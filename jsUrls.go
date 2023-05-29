// Например это было найдено в JS файлах, значит в вывод будет записаны следующее урлы, что ниже
// Идея от сюда https://github.com/edoardottt/lit-bb-hack-tools/tree/main/eefjsf
// /api/users
// /api/posts
// /api/comments

// http://example.com/api/users
// http://example.com/api/posts
// http://example.com/api/comments


package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
)

import (
	"github.com/edoardottt/golazy"
)

func main() {
	helpPtr := flag.Bool("h", false, "Show usage.")
	flag.Parse()

	if *helpPtr {
		help()
	}

	input := ScanTargets()

	results := RetrieveContents(golazy.RemoveDuplicateValues(input), 10)
	for _, elem := range results {
		fmt.Println(elem)
	}
}

// help shows the usage.
func help() {
	var usage = `Take as input on stdin a list of js file urls and print on stdout all the unique endpoints found.
	$> cat js-urls | eefjsf`

	fmt.Println()
	fmt.Println(usage)
	fmt.Println()
	os.Exit(0)
}

// ScanTargets return the array of elements
// taken as input on stdin.
func ScanTargets() []string {
	var result []string

	// accept domains on stdin.
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		domain := strings.ToLower(sc.Text())
		result = append(result, domain)
	}

	return result
}

// RetrieveContents.
func RetrieveContents(input []string, channels int) []string {
	var (
		result = []string{}
		mutex  = &sync.Mutex{}
	)

	r := regexp.MustCompile(`\"\/[a-zA-Z0-9_\/?=&]*\"`)

	limiter := make(chan string, channels) // Limits simultaneous requests.
	wg := sync.WaitGroup{}                 // Needed to not prematurely exit before all requests have been finished.

	for _, domain := range input {
		limiter <- domain

		wg.Add(1)

		go func(domain string) {
			defer wg.Done()
			defer func() { <-limiter }()

			resp, err := http.Get(domain)

			mutex.Lock()

			if err == nil {
				body, err := ioutil.ReadAll(resp.Body)
				if err == nil && len(body) != 0 {
					// Convert the body to type string.
					sb := string(body)
					results := r.FindAllString(sb, -1)
					for _, res := range golazy.RemoveDuplicateValues(results) {
						endpoint := strings.Trim(res, "\"")
						fullURL := buildFullURL(domain, endpoint)
						result = append(result, fullURL)
					}
				}

				resp.Body.Close()
			}
			mutex.Unlock()
		}(domain)
	}

	wg.Wait()

	return golazy.RemoveDuplicateValues(result)
}

// buildFullURL builds the full URL using the domain and endpoint.
func buildFullURL(domain, endpoint string) string {
	baseURL, _ := url.Parse(domain)
	endpointURL, _ := url.Parse(endpoint)
	fullURL := baseURL.ResolveReference(endpointURL).String()
	return fullURL
}
