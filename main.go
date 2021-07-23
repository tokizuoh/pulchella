package main

import (
	"asterism"
	"flag"
	"log"
)

func main() {
	flag.Parse()

	f := flag.Arg(0)
	if f == "fetch" {
		id, err := asterism.GetNewestID()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("ID:", id)
	} else {
		// NOP
	}

}
