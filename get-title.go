package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/tomnomnom/gahttp"
	"golang.org/x/net/html"
)

func extractTitle(z *html.Tokenizer, reqURL string, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}

		t := z.Token()

		if t.Type == html.StartTagToken && t.Data == "title" {
			if z.Next() == html.TextToken {
				title := strings.TrimSpace(z.Token().Data)
				fmt.Printf("%s (%s)\n", title, reqURL)
				break
			}
		}
	}
}

func main() {
	var concurrency int
	flag.IntVar(&concurrency, "c", 20, "Concurrency")
	flag.Parse()

	p := gahttp.NewPipeline()
	p.SetConcurrency(concurrency)

	extractFn := func(req *http.Request, resp *http.Response, err error) {
		if err != nil {
			return
		}

		z := html.NewTokenizer(resp.Body)
		wg := &sync.WaitGroup{}

		for {
			tt := z.Next()
			if tt == html.ErrorToken {
				break
			}

			t := z.Token()

			if t.Type == html.StartTagToken && t.Data == "title" {
				if z.Next() == html.TextToken {
					wg.Add(1)
					go extractTitle(z.Clone(), req.URL.String(), wg)
					break
				}
			}
		}

		wg.Wait()
		resp.Body.Close()
	}

	sc := bufio.NewScanner(os.Stdin)

	client := gahttp.NewClient(gahttp.SkipVerify)
	p.Client = client

	for sc.Scan() {
		p.Get(sc.Text(), extractFn)
	}

	p.Done()
	p.Wait()
}
