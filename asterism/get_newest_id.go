package asterism

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
	"github.com/sclevine/agouti"
)

func GetNewestID() (string, error) {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	targetURL := os.Getenv("TARGET_URL_2")

	driver := agouti.ChromeDriver(agouti.ChromeOptions("args", []string{
		"--headless",
		"--window-size=1,1",
		"--blink-settings=imagesEnabled=false",
		"--disable-gpu",
		"--disable-dev-shm-usage",
		"no-sandbox",
	}), agouti.Debug)

	if err := driver.Start(); err != nil {
		return "", err
	}
	defer driver.Stop()

	page, err := driver.NewPage()
	if err != nil {
		return "", err
	}

	page.Navigate(targetURL)

	src, err := page.HTML()
	if err != nil {
		return "", err
	}

	r := strings.NewReader(src)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", err
	}

	regex := regexp.MustCompile(`article.html\?id=\d{1,}`)
	var needURL string
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		if needURL != "" {
			return
		}
		url, _ := s.Attr("href")
		ok := regex.MatchString(url)
		if ok {
			needURL = url
			return
		}
	})

	if needURL == "" {
		return "", fmt.Errorf("Not found href.")
	}

	// article.html?id=606 -> [article.html?id, 606]
	arr := strings.Split(needURL, "=")

	return arr[1], nil
}
