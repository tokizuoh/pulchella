package topaz

import (
	"fmt"
	"os"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/joho/godotenv"
)

type TwitterConfig struct {
	ConsumerKey      string
	ConsumerSecret   string
	UserAccessToken  string
	UserAccessSecret string
}

func Tweet(msg string) error {
	if err := godotenv.Load(); err != nil {
		return err
	}

	consumerKey := os.Getenv("TWITTER_CONSUMER_KEY")
	consumerSecret := os.Getenv("TWITTER_CONSUMER_SECRET")
	userAccessToken := os.Getenv("TWITTER_USER_ACCESS_TOKEN")
	userAccessSecret := os.Getenv("TWITTER_USER_ACCESS_SECRET")

	config := &TwitterConfig{
		ConsumerKey:      consumerKey,
		ConsumerSecret:   consumerSecret,
		UserAccessToken:  userAccessToken,
		UserAccessSecret: userAccessSecret,
	}

	oauthConfig := oauth1.NewConfig(config.ConsumerKey, config.ConsumerSecret)
	token := oauth1.NewToken(config.UserAccessToken, config.UserAccessSecret)
	httpClient := oauthConfig.Client(oauth1.NoContext, token)
	twitterClient := twitter.NewClient(httpClient)

	_, resp, err := twitterClient.Statuses.Update(msg, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf(resp.Status)
		return err
	}

	return nil
}
