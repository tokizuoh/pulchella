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

type Period struct {
	Start time.Time
	End   time.Time
}

type Event struct {
	Id        int
	Title     string
	Period    Period
	IsCapsule bool
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
func getPeriod2(words []string) (Period, error) {
	var start time.Time
	var end time.Time

	d := strings.Split(words[0], "～")
	l := "2006年1月2日"
	t, err := time.Parse(l, d[0])
	if err != nil {
		return Period{}, err
	}
	start = t

	et := d[1] + words[1]
	l = "2006年1月2日15:04まで"
	t, err = time.Parse(l, et)
	if err != nil {
		return Period{}, err
	}
	end = t

	pd := Period{Start: start, End: end}
	return pd, err
}

// [2021年6月21日～2021年6月30日, 11:59, まで]
func getPeriod3(words []string) (Period, error) {
	var start time.Time
	var end time.Time

	d := strings.Split(words[0], "～")
	l := "2006年1月2日"
	t, err := time.Parse(l, d[0])
	if err != nil {
		return Period{}, err
	}
	start = t

	et := d[1] + words[1] + words[2]
	l = "2006年1月2日15:04まで"
	t, err = time.Parse(l, et)
	if err != nil {
		return Period{}, err
	}
	end = t

	pd := Period{Start: start, End: end}
	return pd, err
}

// [2021年6月30日, ～, 2021年7月10日, 14:59まで]
func getPeriod4A(words []string) (Period, error) {
	var start time.Time
	var end time.Time

	l := "2006年1月2日"
	t, err := time.Parse(l, words[0])
	if err != nil {
		return Period{}, err
	}
	start = t

	l = "2006年1月2日15:04まで"
	v := strings.Join(words[2:4], "")
	t, err = time.Parse(l, v)
	if err != nil {
		return Period{}, err
	}
	end = t

	pd := Period{Start: start, End: end}
	return pd, err
}

// [2021年6月30日, ～, 7月4日, 23:59まで]
func getPeriod4B(words []string) (Period, error) {
	var start time.Time
	var end time.Time

	l := "2006年1月2日"
	t, err := time.Parse(l, words[0])
	if err != nil {
		return Period{}, err
	}
	start = t

	et := words[2] + words[3]
	l = "1月2日15:04まで"
	t, err = time.Parse(l, et)
	if err != nil {
		return Period{}, err
	}
	end = time.Date(start.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())

	pd := Period{Start: start, End: end}
	return pd, err
}

// [2021年6月30日, ～, 2021年7月31日, 14:59, まで]
func getPeriod5(words []string) (Period, error) {
	var start time.Time
	var end time.Time

	l := "2006年1月2日"
	t, err := time.Parse(l, words[0])
	if err != nil {
		return Period{}, err
	}
	start = t

	et := words[2] + words[3]
	l = "2006年1月2日15:04"
	t, err = time.Parse(l, et)
	if err != nil {
		return Period{}, err
	}
	end = t

	pd := Period{Start: start, End: end}
	return pd, err
}

func convertPeriod(from string) (Period, error) {
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

	return Period{}, fmt.Errorf("doesn't meet existing conditions")
}

func getEvent(url string, id int) (Event, bool, error) {
	driver := agouti.ChromeDriver(agouti.ChromeOptions("args", []string{
		"--headless",
		"--window-size=1,1",
		"--blink-settings=imagesEnabled=false",
		"--disable-gpu",
		"--disable-dev-shm-usage",
		"no-sandbox",
	}), agouti.Debug)

	if err := driver.Start(); err != nil {
		return Event{}, false, err
	}
	defer driver.Stop()

	page, err := driver.NewPage()
	if err != nil {
		return Event{}, false, err
	}

	page.Navigate(url)

	src, err := page.HTML()
	if err != nil {
		return Event{}, false, err
	}

	r := strings.NewReader(src)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return Event{}, false, err
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
		return Event{}, false, nil
	}

	pd, err := convertPeriod(holdingPeriod)
	if err != nil {
		return Event{}, false, err
	}

	e := Event{
		Id:        id,
		Title:     title,
		Period:    pd,
		IsCapsule: strings.Contains(title, "ガシャ"),
	}
	return e, true, nil
}

func FetchEvents(fromID, toID int) ([]Event, error) {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	targetURL := os.Getenv("TARGET_URL_1")
	var events []Event
	// TODO: 初期値を最新のIDを取得（get_news_id.go）して設定し、終端を過去に取得した最新のIDにする
	for i := fromID; i <= toID; i++ {
		id := strconv.Itoa(i)
		url := targetURL + id

		e, ok, err := getEvent(url, i)
		if err != nil {
			log.Printf("WARNING: ID [%v] convert error", i)
			continue
		}

		if ok != true {
			continue
		}

		// TODO: DBに保存する
		// title(string), start(date), end(date), type(string)
		//   type: event or capsule
		log.Println("/-------------------------")
		log.Println("I   D: ", e.Id)
		log.Println("TITLE: ", e.Title)
		log.Println("CAPSU: ", e.IsCapsule)
		log.Println("start: ", e.Period.Start)
		log.Println("e n d: ", e.Period.End)
		log.Println("--------------------------")
		events = append(events, e)
	}

	return events, nil
}
