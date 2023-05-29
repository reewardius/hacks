package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func main() {
	file, err := os.Open("subdomains.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	outputFile, err := os.OpenFile("output.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		subdomain := scanner.Text()
		cname, err := resolveCNAME(subdomain)
		if err != nil {
			continue
		}

		if matchRegex(cname) {
			appendToFile(outputFile, cname)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

// Разрешение CNAME для поддомена
func resolveCNAME(subdomain string) (string, error) {
	cmd := exec.Command("dig", subdomain, "CNAME")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	// Извлечение CNAME из вывода команды dig
	re := regexp.MustCompile(`CNAME\s+(.*)\.`)
	match := re.FindStringSubmatch(string(output))
	if len(match) < 2 {
		return "", fmt.Errorf("CNAME not found for %s", subdomain)
	}

	return match[1], nil
}

// Проверка совпадения регулярного выражения
func matchRegex(cname string) bool {
	re := regexp.MustCompile(`s3\.amazonaws\.com|storage\.googleapis\.com|blob\.core\.windows\.net`)
	return re.MatchString(cname)
}

func appendToFile(file *os.File, content string) {
	if _, err := file.WriteString(content + "\n"); err != nil {
		log.Fatal(err)
	}
}
