package main

import (
	"log"
	"net/url"
	"os"

	newrelic "github.com/newrelic/go-agent"

	"github.com/ChimeraCoder/anaconda"
)

const (
	uriLanguagePath  = "/text/analytics/v2.1/languages"
	uriSentimentPath = "/text/analytics/v2.1/sentiment"
	threshold        = 0.6
)

var (
	//See apps.twitter.com/ to obtain you credentials
	consumerKey             = mustGetEnv("TWITTER_CONSUMER_KEY")
	consumerSecret          = mustGetEnv("TWITTER_CONSUMER_SECRET")
	accessToken             = mustGetEnv("TWITTER_ACCESS_TOKEN")
	accessTokenSecret       = mustGetEnv("TWITTER_ACCESS_TOKEN_SECRET")
	nrKey                   = mustGetEnv("NEW_RELIC_LICENSE_KEY")
	azTextAnalyticsKey      = mustGetEnv("TEXT_ANALYTICS_SUBSCRIPTION_KEY")
	azTextAnalyticsEndpoint = mustGetEnv("TEXT_ANALYTICS_ENDPOINT")

	mok = 0
	app newrelic.Application
)

func mustGetEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		panic("missing required environment variable " + name)
	}
	return v
}

func isItBad(tweet anaconda.Tweet) {
	log.Println("analyzing a tweet")
	txn := app.StartTransaction("checkSentiment", nil, nil)
	defer txn.End()

	b, s := badFeedback(tweet.FullText, txn)
	if b {
		log.Printf("Sentiment score detected below threshold: %0.2f [%0.2f]\n", s, threshold)
		pushToNR(tweet, s, txn)
	} else {
		log.Printf("nothing bad detected [%s]\n", tweet.FullText)
	}
}

func badFeedback(message string, txn newrelic.Transaction) (bool, float64) {
	uri := azTextAnalyticsEndpoint + uriLanguagePath
	lang := detectLanguage(uri, message, txn)

	if lang != "" {
		log.Printf("Language detected: %s\n", lang)
		uri = azTextAnalyticsEndpoint + uriSentimentPath
		b, s := detectSentiment(uri, lang, message, txn)
		return b, s
	}
	log.Println("no language detected!")
	return false, 0.0
}

func pushToNR(tweet anaconda.Tweet, score float64, txn newrelic.Transaction) {
	log.Println("recording an event to New Relic Insights")
	defer newrelic.StartSegment(txn, "recordCustomEvent").End()
	err := app.RecordCustomEvent("BadFeedback", map[string]interface{}{
		"Source":      "Twitter",
		"message":     tweet.FullText,
		"authorEmail": tweet.User.Email,
		"author":      tweet.User.Name,
		"score":       score,
	})
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Event recorded successfully")
	}
}

func main() {
	var err error
	config := newrelic.NewConfig("twitterbot", nrKey)
	app, err = newrelic.NewApplication(config)
	if err != nil {
		log.Panicf("NR error: %s", err)
	}
	anaconda.SetConsumerKey(consumerKey)
	anaconda.SetConsumerSecret(consumerSecret)
	api := anaconda.NewTwitterApi(accessToken, accessTokenSecret)

	stream := api.PublicStreamFilter(url.Values{
		"track": []string{"sncf"},
	})

	defer stream.Stop()

	for v := range stream.C {
		t, ok := v.(anaconda.Tweet)
		if !ok {
			log.Printf("received unexpected value of type %T", v)
			continue
		}

		go isItBad(t)
	}
}
