package main

import (
	"flag"
	"fmt"
	"os"

	"golang.org/x/net/html"
)

func main() {
	flag.Parse()
	keys := flag.Args()

	z := html.NewTokenizer(os.Stdin)

	for tt := z.Next(); tt != html.ErrorToken; tt = z.Next() {
		t := z.Token()

		for _, a := range t.Attr {
			if a.Val == "" {
				continue
			}

			// If no keys are specified, output all attribute values
			if len(keys) == 0 {
				fmt.Println(a.Val)
				continue
			}

			// Check if the attribute key matches any of the specified keys
			for _, k := range keys {
				if k == a.Key {
					fmt.Println(a.Val)
					break
				}
			}
		}
	}
}
