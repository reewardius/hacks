package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	// Получение имени файла из флага -l
	filePath := flag.String("l", "", "Путь к файлу со списком сайтов")
	flag.Parse()

	if *filePath == "" {
		log.Fatal("Необходимо указать путь к файлу со списком сайтов через флаг -l")
	}

	// Чтение списка сайтов из файла
	siteList, err := readSiteList(*filePath)
	if err != nil {
		log.Fatalf("Ошибка при чтении файла: %v", err)
	}

	// Проверка CSP политики для каждого сайта
	for _, site := range siteList {
		checkCSP(site)
	}
}

func readSiteList(filePath string) ([]string, error) {
	// Открытие файла
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var siteList []string

	// Чтение списка сайтов из файла
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		site := strings.TrimSpace(scanner.Text())

		// Добавление префикса "https://" к сайту, если он отсутствует
		if !strings.HasPrefix(site, "http://") && !strings.HasPrefix(site, "https://") {
			site = "https://" + site
		}

		siteList = append(siteList, site)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return siteList, nil
}

func checkCSP(site string) {
	// Отправка GET-запроса к сайту
	resp, err := http.Get(site)
	if err != nil {
		log.Printf("Ошибка при отправке запроса к %s: %v", site, err)
		return
	}
	defer resp.Body.Close()

	// Проверка наличия CSP заголовка
	if _, ok := resp.Header["Content-Security-Policy"]; ok {
		// CSP заголовок присутствует, игнорируем этот сайт
		fmt.Println(site)
	} else {
		// CSP заголовок отсутствует, выводим этот сайт
		fmt.Println(site)
	}
}
