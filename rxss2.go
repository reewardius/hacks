// go run rxss.go -l urls.txt -o results.txt -c 40

package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type paramCheck struct {
	url   string
	param string
}

var (
	transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: time.Second,
			DualStack: true,
		}).DialContext,
	}

	httpClient = &http.Client{
		Transport: transport,
	}
)

func main() {
	httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	linksFile := flag.String("l", "", "Path to the file containing links")
	outputFile := flag.String("o", "", "Path to the output file")
	concurrency := flag.Int("c", 40, "Number of concurrent workers")
	flag.Parse()

	var sc *bufio.Scanner
	if *linksFile != "" {
		file, err := os.Open(*linksFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open links file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		sc = bufio.NewScanner(file)
	} else {
		sc = bufio.NewScanner(os.Stdin)
	}

	initialChecks := make(chan paramCheck, *concurrency)
	appendChecks := makePool(initialChecks, func(c paramCheck, output chan paramCheck) {
		reflected, err := checkReflected(c.url)
		if err != nil {
			return
		}

		if len(reflected) == 0 {
			return
		}

		for _, param := range reflected {
			output <- paramCheck{c.url, param}
		}
	})

	charChecks := makePool(appendChecks, func(c paramCheck, output chan paramCheck) {
		wasReflected, err := checkAppend(c.url, c.param, "iy3j4h234hjb23234")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error from checkAppend for URL %s with param %s: %s", c.url, c.param, err)
			return
		}

		if wasReflected {
			output <- paramCheck{c.url, c.param}
		}
	})

	done := makePool(charChecks, func(c paramCheck, output chan paramCheck) {
		outputURL := []string{c.url, c.param}
		filtered := true
		characters := []string{"\"", "'", "<", ">", "$", "|", "(", ")", "`", ":", ";", "{", "}", "%"}
		charLen := len(characters)
		results := make([]string, 0, charLen)

		var wg sync.WaitGroup
		wg.Add(charLen)

		for _, char := range characters {
			go func(c paramCheck, char string) {
				defer wg.Done()

				wasReflected, err := checkAppend(c.url, c.param, "aprefix"+char+"asuffix")
				if err != nil {
					fmt.Fprintf(os.Stderr, "error from checkAppend for URL %s with param %s with %s: %s", c.url, c.param, char, err)
					return
				}

				if wasReflected {
					results = append(results, char)
					filtered = false
				}
			}(c, char)
		}

		wg.Wait()

		if !filtered && len(results) >= 1 {
			outputURL = append(outputURL, results...)
			fmt.Printf("URL: %s Param: %s Unfiltered: %v\n", outputURL[0], outputURL[1], outputURL[2:])

			if *outputFile != "" {
				file, err := os.OpenFile(*outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to open output file: %v\n", err)
					return
				}
				defer file.Close()

				fmt.Fprintf(file, "URL: %s Param: %s Unfiltered: %v\n", outputURL[0], outputURL[1], outputURL[2:])
			}
		}
	})

	for sc.Scan() {
		initialChecks <- paramCheck{url: sc.Text()}
	}

	close(initialChecks)
	<-done
}

func checkReflected(targetURL string) ([]string, error) {
	out := make([]string, 0)

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return out, err
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.100 Safari/537.36")

	resp, err := httpClient.Do(req)
	if err != nil {
		return out, err
	}
	if resp.Body == nil {
		return out, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return out, err
	}

	if strings.HasPrefix(resp.Status, "3") {
		return out, nil
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "" && !strings.Contains(ct, "html") {
		return out, nil
	}

	body := string(b)

	u, err := url.Parse(targetURL)
	if err != nil {
		return out, err
	}

	for key, vv := range u.Query() {
		for _, v := range vv {
			if !strings.Contains(body, v) {
				continue
			}

			out = append(out, key)
		}
	}

	return out, nil
}

func checkAppend(targetURL, param, suffix string) (bool, error) {
	u, err := url.Parse(targetURL)
	if err != nil {
		return false, err
	}

	qs := u.Query()
	val := qs.Get(param)

	qs.Set(param, val+suffix)
	u.RawQuery = qs.Encode()

	reflected, err := checkReflected(u.String())
	if err != nil {
		return false, err
	}

	for _, r := range reflected {
		if r == param {
			return true, nil
		}
	}

	return false, nil
}

type workerFunc func(paramCheck, chan paramCheck)

func makePool(input chan paramCheck, fn workerFunc) chan paramCheck {
	var wg sync.WaitGroup

	output := make(chan paramCheck)
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			for c := range input {
				fn(c, output)
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(output)
	}()

	return output
}
