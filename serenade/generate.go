package serenade

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type MessageType int

const (
	CapsuleType MessageType = iota
	EventType
)

type Event struct {
	Id        int    `json:"id"`
	Title     string `json:"title"`
	Start     string `json:"start"`
	End       string `json:"end"`
	IsCapsule bool   `json:"isCapsule"`
}

func Generate(msgType MessageType, events []Event) (string, error) {
	if err := godotenv.Load(); err != nil {
		return "", err
	}

	var msg string

	switch msgType {
	case CapsuleType:

		for _, e := range events {
			if e.IsCapsule != true {
				continue
			}

			if len(msg) == 0 {
				msg += os.Getenv("CAPSULE_TWEET_TXT")
				msg += "\n"
			}

			var title string
			log.Println(e.Title)
			title = strings.Replace(e.Title, "開催！", "", -1)
			title = strings.Replace(title, "開催!", "", -1)
			msg += fmt.Sprintf("・%v\n", title)
		}

		if len(msg) == 0 {
			msg += os.Getenv("EMPTY_CAPSULE_TWEET_TXT")
		}

	case EventType:
		for _, e := range events {
			if e.IsCapsule == true {
				continue
			}

			if len(msg) == 0 {
				msg += os.Getenv("EVENT_TWEET_TXT")
				msg += "\n"
			}

			var title string
			title = strings.Replace(e.Title, "開催！", "", -1)
			title = strings.Replace(e.Title, "開催!", "", -1)
			msg += fmt.Sprintf("・%v\n", title)
		}

		if len(msg) == 0 {
			msg += os.Getenv("EMPTY_EVENT_TWEET_TXT")
		}
	}

	return msg, nil
}
