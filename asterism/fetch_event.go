package asterism

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
	"github.com/sclevine/agouti"
)

type period struct {
	start time.Time
	end   time.Time
}

type event struct {
	title  string
	period period
}

// ["", "1", "", "a"] -> ["1", "a"]
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

// [2021年6月11日～2021年6月30日, 11:59まで]
func getPeriod2(words []string) (period, error) {
	var start time.Time
	var end time.Time

	d := strings.Split(words[0], "～")
	l := "2006年1月2日"
	t, err := time.Parse(l, d[0])
	if err != nil {
		return period{}, err
	}
	start = t

	et := d[1] + words[1]
	l = "2006年1月2日15:04まで"
	t, err = time.Parse(l, et)
	if err != nil {
		return period{}, err
	}
	end = t

	pd := period{start: start, end: end}
	return pd, err
}

// [2021年6月21日～2021年6月30日, 11:59, まで]
func getPeriod3(words []string) (period, error) {
	var start time.Time
	var end time.Time

	d := strings.Split(words[0], "～")
	l := "2006年1月2日"
	t, err := time.Parse(l, d[0])
	if err != nil {
		return period{}, err
	}
	start = t

	et := d[1] + words[1] + words[2]
	l = "2006年1月2日15:04まで"
	t, err = time.Parse(l, et)
	if err != nil {
		return period{}, err
	}
	end = t

	pd := period{start: start, end: end}
	return pd, err
}

// [2021年6月30日, ～, 2021年7月10日, 14:59まで]
func getPeriod4A(words []string) (period, error) {
	var start time.Time
	var end time.Time

	l := "2006年1月2日"
	t, err := time.Parse(l, words[0])
	if err != nil {
		return period{}, err
	}
	start = t

	l = "2006年1月2日15:04まで"
	v := strings.Join(words[2:4], "")
	t, err = time.Parse(l, v)
	if err != nil {
		return period{}, err
	}
	end = t

	pd := period{start: start, end: end}
	return pd, err
}

// [2021年6月30日, ～, 7月4日, 23:59まで]
func getPeriod4B(words []string) (period, error) {
	var start time.Time
	var end time.Time

	l := "2006年1月2日"
	t, err := time.Parse(l, words[0])
	if err != nil {
		return period{}, err
	}
	start = t

	et := words[2] + words[3]
	l = "1月2日15:04まで"
	t, err = time.Parse(l, et)
	if err != nil {
		return period{}, err
	}
	end = time.Date(start.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())

	pd := period{start: start, end: end}
	return pd, err
}

// [2021年6月30日, ～, 2021年7月31日, 14:59, まで]
func getPeriod5(words []string) (period, error) {
	var start time.Time
	var end time.Time

	l := "2006年1月2日"
	t, err := time.Parse(l, words[0])
	if err != nil {
		return period{}, err
	}
	start = t

	et := words[2] + words[3]
	l = "2006年1月2日15:04"
	t, err = time.Parse(l, et)
	if err != nil {
		return period{}, err
	}
	end = t

	pd := period{start: start, end: end}
	return pd, err
}

func convertPeriod(from string) (period, error) {
	words := strings.Fields(from)
	lw := len(words)

	switch lw {
	case 2:
		return getPeriod2(words)
	case 3:
		return getPeriod3(words)
	case 4:
		if strings.Contains(words[2], "年") {
			return getPeriod4A(words)
		} else {
			return getPeriod4B(words)
		}
	case 5:
		return getPeriod5(words)
	}

	return period{}, fmt.Errorf("doesn't meet existing conditions")
}

func getEvent(url string) (event, bool, error) {
	driver := agouti.ChromeDriver(agouti.ChromeOptions("args", []string{
		"--headless",
		"--window-size=1,1",
		"--blink-settings=imagesEnabled=false",
		"--disable-gpu",
		"--disable-dev-shm-usage",
		"no-sandbox",
	}), agouti.Debug)

	if err := driver.Start(); err != nil {
		return event{}, false, err
	}
	defer driver.Stop()

	page, err := driver.NewPage()
	if err != nil {
		return event{}, false, err
	}

	page.Navigate(url)

	src, err := page.HTML()
	if err != nil {
		return event{}, false, err
	}

	r := strings.NewReader(src)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return event{}, false, err
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
		return event{}, false, nil
	}

	pd, err := convertPeriod(holdingPeriod)
	if err != nil {
		return event{}, false, err
	}

	e := event{
		title:  title,
		period: pd,
	}
	return e, true, nil
}

func FetchEvent() error {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	targetURL := os.Getenv("TARGET_URL_1")
	for i := 595; i < 606; i++ {
		id := strconv.Itoa(i)
		url := targetURL + id

		e, ok, err := getEvent(url)
		if err != nil {
			log.Printf("WARNING: ID [%v] convert error", i)
			continue
		}

		if ok != true {
			continue
		}

		log.Println("/-------------------------")
		log.Println("TITLE: ", e.title)
		log.Println("start: ", e.period.start)
		log.Println("e n d: ", e.period.end)
		log.Println("--------------------------")
	}

	return nil
}
