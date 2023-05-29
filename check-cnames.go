package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type job struct {
	domain, server string
}

type result struct {
	cname   string
	resolved bool
}

func main() {
	servers := []string{
		"8.8.8.8",
		"8.8.4.4",
		"9.9.9.9",
		"1.1.1.1",
		"1.0.0.1",
	}

	rand.Seed(time.Now().Unix())

	jobs := make(chan job)
	results := make(chan result)

	var wg sync.WaitGroup
	const numWorkers = 20

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			for j := range jobs {
				cname, resolved := getCNAME(j.domain, j.server)
				results <- result{cname: cname, resolved: resolved}
			}
			wg.Done()
		}()
	}

	// Start result handler
	go func() {
		for r := range results {
			if !r.resolved {
				fmt.Printf("%s does not resolve (pointed at by %s)\n", r.cname, j.domain)
			}
		}
	}()

	sc := bufio.NewScanner(os.Stdin)

	for sc.Scan() {
		target := strings.ToLower(strings.TrimSpace(sc.Text()))
		if target == "" {
			continue
		}
		server := servers[rand.Intn(len(servers))]

		jobs <- job{target, server}
	}
	close(jobs)

	wg.Wait()
	close(results)
}

func resolves(domain string) bool {
	_, err := net.LookupHost(domain)
	return err == nil
}

func getCNAME(domain, server string) (string, bool) {
	c := new(dns.Client)

	m := new(dns.Msg)
	if domain[len(domain)-1:] != "." {
		domain += "."
	}
	m.SetQuestion(domain, dns.TypeCNAME)
	m.RecursionDesired = true

	r, _, err := c.Exchange(m, net.JoinHostPort(server, "53"))
	if err != nil {
		return "", false
	}

	if len(r.Answer) == 0 {
		return "", false
	}

	for _, ans := range r.Answer {
		if cname, ok := ans.(*dns.CNAME); ok {
			return cname.Target, resolves(cname.Target)
		}
	}

	return "", false
}
