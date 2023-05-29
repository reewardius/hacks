package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/net/html"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	z := html.NewTokenizer(reader)
	out := bufio.NewWriter(os.Stdout)

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}

		t := z.Token()

		if t.Type == html.CommentToken {
			d := string(t.Data)
			d = strings.ReplaceAll(d, "\n", " ")
			d = strings.TrimSpace(d)
			if d == "" {
				continue
			}
			out.WriteString(d)
			out.WriteByte('\n')
		}
	}

	out.Flush()
}
