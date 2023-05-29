// Если размеры равны, результат не записывается в файл и переходит к следующему URL-адресу
// go run redircheck.go -l domains.txt -o redircheck.txt

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
)

func main() {
	filePath := flag.String("l", "", "Path to the file containing URLs")
	outputFilePath := flag.String("o", "", "Path to the output file")
	concurrency := flag.Int("c", 10, "Number of concurrent requests")
	flag.Parse()

	if *filePath == "" {
		fmt.Println("Please provide a file path using the -l flag")
		return
	}

	if *outputFilePath == "" {
		fmt.Println("Please provide an output file path using the -o flag")
		return
	}

	urls, err := readURLsFromFile(*filePath)
	if err != nil {
		fmt.Printf("Error reading URLs from file: %s\n", err)
		return
	}

	outputFile, err := os.Create(*outputFilePath)
	if err != nil {
		fmt.Printf("Error creating output file: %s\n", err)
		return
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, *concurrency)

	for _, url := range urls {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(url string) {
			defer func() {
				<-semaphore
				wg.Done()
			}()

			resp, err := http.Get(url)
			if err != nil {
				fmt.Printf("Error fetching URL %s: %s\n", url, err)
				return
			}

			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response body from URL %s: %s\n", url, err)
				return
			}

			output := fmt.Sprintf("URL: %s - %d status code - %d bytes", url, resp.StatusCode, len(body))

			// Check for redirects
			if resp.StatusCode >= 300 && resp.StatusCode < 400 {
				redirectURL, err := resp.Location()
				if err != nil {
					fmt.Printf("Error getting redirect URL: %s\n", err)
					return
				}

				redirectResp, err := http.Get(redirectURL.String())
				if err != nil {
					fmt.Printf("Error fetching redirect URL %s: %s\n", redirectURL, err)
					return
				}

				defer redirectResp.Body.Close()

				redirectBody, err := ioutil.ReadAll(redirectResp.Body)
				if err != nil {
					fmt.Printf("Error reading response body from redirect URL %s: %s\n", redirectURL, err)
					return
				}

				if len(body) == len(redirectBody) {
					return // Пропустить запись в файл, если размеры тел ответов равны
				}

				output += fmt.Sprintf(" - Redirect: %s - %d status code - %d bytes", redirectURL, redirectResp.StatusCode, len(redirectBody))
			}

			fmt.Println(output)
			//fmt.Fprintln(writer, output) - Удалена строка записи результата в файл
		}(url)
	}

	wg.Wait()
	writer.Flush()
	fmt.Printf("Results saved to %s\n", *outputFilePath)
}

func readURLsFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var urls []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}
