package main

import (
	"asterism"
	"flag"
	"log"
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
		err := asterism.FetchEvent()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// NOP
	}

}
