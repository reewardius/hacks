package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func extractParametersFromFile(filename string) ([]string, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var parameters []string
	// Регулярное выражение для поиска параметров в JS-файле
	regex := regexp.MustCompile(`\b(?:var|let|const)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\b`)

	matches := regex.FindAllStringSubmatch(string(content), -1)
	for _, match := range matches {
		// Извлекаем только параметры, пропуская функции и другие конструкции
		if !strings.HasPrefix(match[1], "function") {
			parameters = append(parameters, match[1])
		}
	}

	return parameters, nil
}

func extractHiddenParamsFromURL(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var hiddenParams []string
	// Регулярное выражение для поиска скрытых параметров в HTML-странице
	regex := regexp.MustCompile(`<input\s+type="hidden"\s+name="([^"]+)"`)

	matches := regex.FindAllStringSubmatch(string(body), -1)
	for _, match := range matches {
		hiddenParams = append(hiddenParams, match[1])
	}

	return hiddenParams, nil
}

func main() {
	urlsFile := "urls.txt"
	paramsFile := "params.txt"

	urls, err := ioutil.ReadFile(urlsFile)
	if err != nil {
		fmt.Printf("Failed to read URLs file: %s\n", err)
		return
	}

	urlList := strings.Split(string(urls), "\n")

	outputFile, err := os.Create(paramsFile)
	if err != nil {
		fmt.Printf("Failed to create params file: %s\n", err)
		return
	}
	defer outputFile.Close()

	for _, url := range urlList {
		parameters, err := extractParametersFromFile(url)
		if err != nil {
			fmt.Printf("Failed to extract parameters from JS file: %s\n", err)
			continue
		}

		hiddenParams, err := extractHiddenParamsFromURL(url)
		if err != nil {
			fmt.Printf("Failed to extract hidden parameters from URL: %s\n", err)
			continue
		}

		fmt.Fprintln(outputFile, "URL:", url)
		for _, param := range parameters {
			fmt.Fprintln(outputFile, param)
		}

		for _, param := range hiddenParams {
			fmt.Fprintln(outputFile, param)
		}

		fmt.Fprintln(outputFile, "")
	}

	fmt.Println("Parameter extraction completed. Results are written to", paramsFile)
}
