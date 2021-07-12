package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
	"github.com/sclevine/agouti"
)

func removeEmpty(arr []string) []string {
	var newArr []string
	for _, a := range arr {
		if len(a) == 0 {
			continue
		}
		newArr = append(newArr, a)
	}
	return newArr
}

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

	// extract title
	var title string
	doc.Find("#news-detail > section > h3").Each(func(i int, s *goquery.Selection) {
		title = s.Text()
	})

	// extract holding period
	var holdingPeriod string
	doc.Find("#news-detail > section > div.text > div > div").Each(func(i int, s *goquery.Selection) {

		txts := strings.Split(s.Text(), "\n")
		txts = removeEmpty(txts)

		f := false
		for _, t := range txts {
			if f {
				holdingPeriod = t
				break
			}

			if t == "開催期間" {
				f = true
			}
		}
	})

	log.Println("TITLE: ", title)
	log.Println("PERIOD: ", holdingPeriod)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	for i := 590; i < 600; i++ {
		id := strconv.Itoa(i)
		url := os.Getenv("TARGET_URL") + id
		getPage(url)
	}

}
