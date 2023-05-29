package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/edoardottt/golazy"
	"github.com/fatih/color"
	"github.com/rodaine/table"
)

var Reset = "\033[0m"
var Red = "\033[31m"

var myTransport http.RoundTripper = &http.Transport{
	Proxy:                 http.ProxyFromEnvironment,
	ResponseHeaderTimeout: time.Second * 15,
}

var myClient = &http.Client{Transport: myTransport}

// main.
func main() {
	helpPtr := flag.Bool("h", false, "Show usage.")
	flag.Parse()

	if *helpPtr {
		help()
	}

	TestMethods(ScanTargets())
}

// help shows the usage.
func help() {
	usage := `Take as input on stdin a list of urls and print on stdout all the status codes and body sizes for HTTP methods.
	$> cat urls | tahm`

	fmt.Println()
	fmt.Println(usage)
	fmt.Println()
	os.Exit(0)
}

// ScanTargets return the array of elements
// taken as input on stdin.
func ScanTargets() []string {
	var result []string

	// accept domains on stdin
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		domain := strings.ToLower(sc.Text())
		result = append(result, domain)
	}

	return golazy.RemoveDuplicateValues(result)
}

// TestMethods.
func TestMethods(input []string) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	for _, elem := range input {
		fmt.Printf("= %s%s%s =\n", Red, elem, Reset)

		tbl := table.New("METHOD", "STATUS", "SIZE")
		tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

		testMethod(tbl, elem, "GET")
		testMethod(tbl, elem, "POST")
		testMethod(tbl, elem, "PUT")
		testMethod(tbl, elem, "DELETE")
		testMethod(tbl, elem, "HEAD")
		testMethod(tbl, elem, "CONNECT")
		testMethod(tbl, elem, "OPTIONS")
		testMethod(tbl, elem, "TRACE")
		testMethod(tbl, elem, "PATCH")

		tbl.Print()
		fmt.Println("---------------------------")
		fmt.Println()
	}
}

// testMethod performs a HTTP request for the given method and adds a row to the table.
func testMethod(tbl *table.Table, elem string, method string) {
	status, size, err := Request(elem, method)
	if err == nil {
		tbl.AddRow(method, status, strconv.Itoa(size))
	}
}

// Request performs an HTTP request.
func Request(target string, method string) (string, int, error) {
	var req *http.Request
	var err error

	if method == "POST" {
		req, err = createPostRequest(target)
	} else if method == "PUT" {
		req, err = createPutRequest(target)
	} else {
		req, err = http.NewRequest(method, target, nil)
	}

	if err != nil {
		return "", 0, err
	}

	resp, err := myClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}

	return resp.Status, len(body), nil
}

// createPostRequest creates a POST request.
func createPostRequest(target string) (*http.Request, error) {
	postBody, _ := json.Marshal("{data}")
	responseBody := bytes.NewBuffer(postBody)

	req, err := http.NewRequest(http.MethodPost, target, responseBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// createPutRequest creates a PUT request.
func createPutRequest(target string) (*http.Request, error) {
	jsonData, _ := json.Marshal("{data}")
	req, err := http.NewRequest(http.MethodPut, target, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	return req, nil
}
