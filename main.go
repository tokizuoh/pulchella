package main

import (
	"asterism"
	"context"
	"flag"
	"fmt"
	"log"
	"strconv"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

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
		_, err := asterism.FetchEvents()
		if err != nil {
			log.Fatal(err)
		}
	} else if f == "af" {
		opt := option.WithCredentialsFile("key.json")
		config := &firebase.Config{DatabaseURL: "https://pulchella-37dfa-default-rtdb.firebaseio.com/"}
		app, err := firebase.NewApp(context.Background(), config, opt)
		if err != nil {
			fmt.Errorf("error initializing app: %v", err)
			return
		}

		ctx := context.Background()
		client, err := app.Database(ctx)
		if err != nil {
			log.Fatal(err)
		}

		events, err := asterism.FetchEvents()
		if err != nil {
			log.Fatal(err)
		}

		ref := client.NewRef("hoge")
		for _, e := range events {

			id := strconv.Itoa(e.Id)
			usersRef := ref.Child(id)
			err = usersRef.Set(ctx, map[string]interface{}{
				"id":        e.Id,
				"title":     e.Title,
				"start":     e.Period.Start,
				"end":       e.Period.End,
				"isCapsule": e.IsCapsule,
			})
			if err != nil {
				log.Fatal(err)
			}
		}

		log.Println("Done")

	} else {
		// NOP
	}

}
