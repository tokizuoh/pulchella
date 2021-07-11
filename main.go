package main

import (
	"log"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
	"github.com/sclevine/agouti"
)

func getPage(url string) {
	driver := agouti.ChromeDriver(agouti.ChromeOptions("args", []string{
		"--headless",
		"--window-size=1,1",
		"--blink-settings=imagesEnabled=false",
		"--disable-gpu",
		"--disable-dev-shm-usage",
		"no-sandbox",
	}), agouti.Debug)

	if err := driver.Start(); err != nil {
		log.Fatal(err)
	}
	defer driver.Stop()

	page, err := driver.NewPage()
	if err != nil {
		log.Fatal(err)
	}

	page.Navigate(url)

	src, err := page.HTML()
	if err != nil {
		log.Fatal(err)
	}

	r := strings.NewReader(src)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("#news-detail > section > div.text > div > div").Each(func(i int, s *goquery.Selection) {
		log.Println(i, s.Text())
	})
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	id := "599"
	url := os.Getenv("TARGET_URL") + id
	getPage(url)
}
