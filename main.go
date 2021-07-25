package main

import (
	"asterism"
	"context"
	"flag"
	"log"
	"os"
	"strconv"

	firebase "firebase.google.com/go"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

type Event struct {
	Id        int    `json:"id"`
	Title     string `json:"title"`
	Start     string `json:"start"`
	End       string `json:"end"`
	IsCapsule bool   `json:"isCapsule"`
}

func getNewEvents() ([]Event, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	databaseURL := os.Getenv("DATABASE_URL")

	opt := option.WithCredentialsFile("key.json")
	config := &firebase.Config{DatabaseURL: databaseURL}
	app, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	client, err := app.Database(ctx)
	if err != nil {
		return nil, err
	}

	var rese map[string]Event
	ref := client.NewRef("hoge")
	if err := ref.Get(ctx, &rese); err != nil {
		return nil, err
	}

	// 既にアップロード済みのイベントの中で最新のID
	var updatedNewstId int
	for _, re := range rese {
		if updatedNewstId < re.Id {
			updatedNewstId = re.Id
		}
	}

	// 現在のwebページから取得できる最新のID
	toIDStr, err := asterism.GetNewestID()
	if err != nil {
		return nil, err
	}

	toID, err := strconv.Atoi(toIDStr)
	if err != nil {
		return nil, err
	}

	log.Println("##### GET NEW ID #####")
	log.Println("# FROM: ", updatedNewstId+1)
	log.Println("# TO  : ", toID)
	events, err := asterism.FetchEvents(updatedNewstId+1, toID)
	if err != nil {
		return nil, err
	}

	var newEvents []Event

	for _, e := range events {
		var isNewEvent = true
		for _, re := range rese {
			if e.Id == re.Id {
				isNewEvent = false
				break
			}
		}
		if isNewEvent != true {
			continue
		}
		event := Event{
			Id:        e.Id,
			Title:     e.Title,
			Start:     e.Period.Start.String(),
			End:       e.Period.Start.String(),
			IsCapsule: e.IsCapsule,
		}
		newEvents = append(newEvents, event)
	}

	return newEvents, nil
}

func updateEvents(events []Event) error {
	if err := godotenv.Load(); err != nil {
		return err
	}

	databaseURL := os.Getenv("DATABASE_URL")

	opt := option.WithCredentialsFile("key.json")
	config := &firebase.Config{DatabaseURL: databaseURL}
	app, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		return err
	}

	ctx := context.Background()
	client, err := app.Database(ctx)
	if err != nil {
		return err
	}

	ref := client.NewRef("hoge")
	for _, e := range events {
		id := strconv.Itoa(e.Id)
		usersRef := ref.Child(id)
		err = usersRef.Set(ctx, map[string]interface{}{
			"id":        e.Id,
			"title":     e.Title,
			"start":     e.Start,
			"end":       e.End,
			"isCapsule": e.IsCapsule,
		})
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func main() {
	flag.Parse()

	f := flag.Arg(0)
	if f == "newest" {
		id, err := asterism.GetNewestID()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("ID:", id)
	} else if f == "fetch" {
		_, err := asterism.FetchEvents(0, 1000)
		if err != nil {
			log.Fatal(err)
		}
	} else if f == "update" {
		newEvents, err := getNewEvents()
		if err != nil {
			log.Fatal(err)
		}
		if err := updateEvents(newEvents); err != nil {
			log.Fatal(err)
		}
	} else {
		// NOP
	}

}
