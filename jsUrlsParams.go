// params.txt && urls.txt /path


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
	"github.com/edoardottt/golazy"\
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

	params, err := readParamsFromFile("params.txt")
	if err != nil {
		fmt.Printf("Failed to read params from file: %v\n", err)
		return result
	}

	for _, domain := range input {
		limiter <- domain

		wg.Add(1)

		go func(domain string) {
			defer wg.Done()
			defer func() { <-limiter }()

			resp, err := http.Get(domain)
			if err != nil {
				fmt.Printf("Failed to make a request to %s: %v\n", domain, err)
				return
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Failed to read response body from %s: %v\n", domain, err)
				return
			}

			sb := string(body)
			for _, param := range params {
				if strings.Contains(sb, param+"=FUZZ") {
					mutex.Lock()
					results := r.FindAllString(sb, -1)
					for _, res := range golazy.RemoveDuplicateValues(results) {
						endpoint := strings.Trim(res, "\"")
						fullURL := buildFullURL(domain, endpoint)
						result = append(result, fullURL)
					}
					mutex.Unlock()
					break
				}
			}
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

// readParamsFromFile reads parameters from a file.
func readParamsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var params []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		param := scanner.Text()
		params = append(params, param)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return params, nil
}
