package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/ericchiang/css"
	"github.com/tomnomnom/gahttp"
	"golang.org/x/net/html"
)

func extractSelector(r io.Reader, selector string) ([]string, error) {
	out := []string{}

	sel, err := css.Parse(selector)
	if err != nil {
		return out, err
	}

	node, err := html.Parse(r)
	if err != nil {
		return out, err
	}

	var extract func(*html.Node)
	extract = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, ele := range sel.Select(n) {
				if ele.FirstChild != nil && ele.FirstChild.Type == html.TextNode {
					out = append(out, strings.TrimSpace(ele.FirstChild.Data))
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}

	extract(node)
	return out, nil
}

func extractComments(r io.Reader) []string {
	z := html.NewTokenizer(r)

	out := []string{}
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}

		t := z.Token()

		if t.Type == html.CommentToken {
			d := strings.ReplaceAll(t.Data, "\n", " ")
			d = strings.TrimSpace(d)
			if d == "" {
				continue
			}
			out = append(out, d)
		}
	}
	return out
}

func extractAttribs(r io.Reader, attribs []string) []string {
	z := html.NewTokenizer(r)

	out := []string{}
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}

		t := z.Token()

		for _, a := range t.Attr {
			if a.Val != "" {
				for _, attrib := range attribs {
					if attrib == a.Key {
						out = append(out, a.Val)
					}
				}
			}
		}
	}
	return out
}

func extractTags(r io.Reader, tags []string) []string {
	z := html.NewTokenizer(r)

	out := []string{}
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}

		t := z.Token()

		if t.Type == html.StartTagToken {
			for _, tag := range tags {
				if t.Data == tag {
					if z.Next() == html.TextToken {
						text := strings.TrimSpace(z.Token().Data)
						if text != "" {
							out = append(out, text)
						}
					}
				}
			}
		}
	}
	return out
}

type target struct {
	location string
	r        io.ReadCloser
}

func main() {
	flag.Parse()

	mode := flag.Arg(0)
	if mode == "" {
		fmt.Println("Accept URLs or filenames for HTML documents on stdin and extract parts of them.")
		fmt.Println("")
		fmt.Println("Usage: html-tool <mode> [<args>]")
		fmt.Println("")
		fmt.Println("Modes:")
		fmt.Println("	tags <tag-names>        Extract text contained in tags")
		fmt.Println("	attribs <attrib-names>  Extract attribute values")
		fmt.Println("	comments                Extract comments")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("	cat urls.txt | html-tool tags title a strong")
		fmt.Println("	find . -type f -name \"*.html\" | html-tool attribs src href")
		fmt.Println("	cat urls.txt | html-tool comments")
		return
	}

	args := flag.Args()[1:]

	targets := make(chan *target)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for t := range targets {
			var vals []string

			switch mode {
			case "tags":
				vals = extractTags(t.r, args)
			case "attribs":
				vals = extractAttribs(t.r, args)
			case "comments":
				vals = extractComments(t.r)
			case "query":
				var err error
				vals, err = extractSelector(t.r, flag.Arg(1))
				if err != nil {
					fmt.Fprintf(os.Stderr, "failed to parse CSS selector: %s\n", err)
					break
				}
			default:
				fmt.Fprintf(os.Stderr, "unsupported mode '%s'\n", mode)
				break
			}

			for _, v := range vals {
				fmt.Println(v)
			}

			t.r.Close()
		}
		wg.Done()
	}()

	p := gahttp.NewPipeline()
	p.SetClient(gahttp.NewClient(gahttp.SkipVerify))
	p.SetConcurrency(20)

	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		location := strings.TrimSpace(sc.Text())

		nl := strings.ToLower(location)
		if strings.HasPrefix(nl, "http:") || strings.HasPrefix(nl, "https:") {
			p.Get(location, func(req *http.Request, resp *http.Response, err error) {
				if err != nil {
					fmt.Fprintf(os.Stderr, "failed to fetch URL: %s\n", err)
				}
				if resp != nil && resp.Body != nil {
					targets <- &target{req.URL.String(), resp.Body}
				}
			})
			continue
		}

		f, err := os.Open(location)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open file: %s\n", err)
			continue
		}

		targets <- &target{location, f}
	}
	p.Done()
	p.Wait()

	close(targets)
	wg.Wait()
}
