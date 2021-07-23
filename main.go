package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

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

type date struct {
	time      time.Time
	isUpdated bool
}

type period struct {
	start date
	end   date
}

type event struct {
	title  string
	period period
}

func convertPeriod(from string) (period, bool) {
	words := strings.Fields(from)

	var start date
	var end date
	for i, w := range words {
		_, err := strconv.Atoi(w[:1])
		if err != nil {
			continue
		}

		if start.isUpdated == false {
			// スタート生成
			if len(words) == 3 {
				// 2021年6月21日～2021年6月30日 11:59 まで

				// [2021年6月21日, 2021年6月30日]
				d := strings.Split(words[0], "～")
				l := "2006年1月2日"
				t, err := time.Parse(l, d[0])
				if err != nil {
					return period{}, false
				}
				start.time = t
				start.isUpdated = true

				et := d[1] + words[1] + words[2]
				l = "2006年1月2日15:04まで"
				t, err = time.Parse(l, et)
				if err != nil {
					return period{}, false
				}
				end.time = t
				end.isUpdated = true

			} else if len(words) == 2 {
				// 2021年6月11日～2021年6月30日 11:59まで
				d := strings.Split(words[0], "～")
				l := "2006年1月2日"
				t, err := time.Parse(l, d[0])
				if err != nil {
					return period{}, false
				}
				start.time = t
				start.isUpdated = true

				et := d[1] + words[1]
				l = "2006年1月2日15:04まで"
				t, err = time.Parse(l, et)
				if err != nil {
					return period{}, false
				}
				end.time = t
				end.isUpdated = true

			} else if len(words) == 5 {
				// 2021年6月30日 ～ 2021年7月31日 14:59 まで
				l := "2006年1月2日"
				t, err := time.Parse(l, words[0])
				if err != nil {
					return period{}, false
				}
				start.time = t
				start.isUpdated = true

				et := words[2] + words[3]
				l = "2006年1月2日15:04"
				t, err = time.Parse(l, et)
				if err != nil {
					return period{}, false
				}
				end.time = t
				end.isUpdated = true

			} else {
				// len(word) == 4 のとき
				if strings.Contains(words[2], "年") {
					// 2021年6月11日　～　2021年6月21日 11:59まで
					l := "2006年1月2日"
					t, err := time.Parse(l, w)
					if err != nil {
						return period{}, false
					}
					start.time = t
					start.isUpdated = true
				} else {
					// 2021年6月30日 ～ 7月4日 23:59まで
					l := "2006年1月2日"
					t, err := time.Parse(l, words[0])
					if err != nil {
						return period{}, false
					}
					start.time = t
					start.isUpdated = true

					et := words[2] + words[3]
					l = "1月2日15:04まで"
					t, err = time.Parse(l, et)
					if err != nil {
						return period{}, false
					}
					end.time = time.Date(start.time.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
					end.isUpdated = true
				}
			}
		} else if end.isUpdated == false {
			if i+1 < len(words) {
				// [2021年6月30日, ～, 2021年7月10日, 14:59まで]
				if len(words) == 4 {
					l := "2006年1月2日15:04まで"
					v := strings.Join(words[i:i+2], "")
					t, err := time.Parse(l, v)
					if err != nil {
						return period{}, false
					}
					end.time = t
					end.isUpdated = true
				}
			} else if i+2 < len(words) {
				// [2021年6月30日, ～, 2021年7月10日, 14:59, まで]
				l := "2006年1月2日15:04まで"
				v := strings.Join(words[i:i+3], "")
				t, err := time.Parse(l, v)
				if err != nil {
					return period{}, false
				}
				end.time = t
				end.isUpdated = true
			}
		}
	}

	success := start.isUpdated && end.isUpdated
	if !success {
		return period{}, false
	}

	return period{start: start, end: end}, true
}

func getEvent(url string) (event, bool) {
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

			if t == "開催期間" || t == "イベント開催期間" {
				f = true
			}
		}
	})

	if len(holdingPeriod) == 0 {
		return event{}, false
	}

	pd, ok := convertPeriod(holdingPeriod)
	if ok != true {
		log.Println("TITLE:", title)
		return event{}, false
	}

	e := event{
		title:  title,
		period: pd,
	}
	return e, true
}

func fetchEvent() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	targetURL := os.Getenv("TARGET_URL_1")
	for i := 595; i < 606; i++ {
		id := strconv.Itoa(i)
		url := targetURL + id

		e, ok := getEvent(url)
		if ok != true {
			errMsg := fmt.Sprintf("id: %s", id)
			log.Println(errMsg)
			continue
		}

		log.Println("/-------------------------")
		log.Println("TITLE: ", e.title)
		log.Println("start: ", e.period.start.time)
		log.Println("e n d: ", e.period.end.time)
		log.Println("--------------------------")
	}
}

func getNewestID() (string, error) {
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

	return needURL, nil
}

func main() {
	flag.Parse()

	f := flag.Arg(0)
	if f == "fetch" {
		// getNewestID()
		// fetchEvent()
	} else {
		// NOP
	}

}
