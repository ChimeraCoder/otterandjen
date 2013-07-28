package main

import (
	"flag"
	"github.com/ChimeraCoder/anaconda"
	"github.com/garyburd/redigo/redis"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const SLEEP_INTERVAL = 120

var httpAddr = flag.String("addr", ":8000", "HTTP server address")

var TWITTER_CONSUMER_KEY = os.Getenv("TWITTER_CONSUMER_KEY")
var TWITTER_CONSUMER_SECRET = os.Getenv("TWITTER_CONSUMER_SECRET")
var TWITTER_ACCESS_TOKEN = os.Getenv("TWITTER_ACCESS_TOKEN")
var TWITTER_ACCESS_TOKEN_SECRET = os.Getenv("TWITTER_ACCESS_TOKEN_SECRET")
var REDIS_ADDRESS = os.Getenv("REDIS_ADDRESS")
var REDIS_PASSWORD = os.Getenv("REDIS_PASSWORD")

var TARGET_USERS = []string{"chimeracoder", "rubinovitz"}

var c redis.Conn

func TweetMentions(tweet anaconda.Tweet, username string) bool {
	return strings.Contains(tweet.Text, "@"+username)
}

func TweetMentionsATarget(tweet anaconda.Tweet) bool {
	//TODO it shouldn't count if the user mentions himself/herself
	for _, username := range TARGET_USERS {
		if TweetMentions(tweet, username) {
			return true
		}
	}
	return false
}

//Retweet the tweet and store the fact that this has been done in Redis
//If the tweet has already been retweeted before, do nothing
func retweetAndLog(api anaconda.TwitterApi, tweet anaconda.Tweet) (err error) {
	already_retweeted, err := alreadyRetweeted(tweet)
	if err != nil {
		return err
	}

	//Only retweet (and log) if the tweet has not already been retweeted
	if !already_retweeted {
		if _, err = api.Retweet(tweet.Id, true); err != nil {
			return
		}

		if _, err = c.Do("SET", tweet.Id_str, strconv.FormatInt(time.Now().Unix(), 10)); err != nil {
			return
		}
		log.Printf("Set %s in redis", tweet.Id_str)

		return
	}

	return
}

//Return true if the tweet was already retweeted previously
func alreadyRetweeted(tweet anaconda.Tweet) (retweeted bool, err error) {
	timestamp, err := c.Do("GET", tweet.Id_str)

	timestamp_b, ok := timestamp.([]byte)

	if !ok || (string(timestamp_b) == "") {
		retweeted = false //This is redundant, since retweeted defaults to false
		log.Print("Was not already retweeted")
	} else {
		retweeted = true
		log.Printf("Was already retweeted on %s", string(timestamp_b))
	}
	return

}

func checkForTweets(api anaconda.TwitterApi) error {
	searchResult, err := api.GetHomeTimeline()
	if err != nil {
		log.Print("error fetching timeline: %v", err)
        return err
	}
	//Assume that we haven't tweeted at each other more than 10 times since the last check
	//Knowing us, this is a very bad assumption.

	log.Printf("We have %d results", len(searchResult))
	//Iterate over the tweets in chronological order (the reverse order from what is returned)
	for i := len(searchResult) - 1; i >= 0; i-- {
		tweet := searchResult[i]
		if TweetMentionsATarget(tweet) {
			if err := retweetAndLog(api, tweet); err != nil {
                log.Print("error when retweeting %v", err)
                continue
			}
			log.Print(tweet.Text)
		} else {
			//log.Printf("Skipping tweet %v", tweet.Text)
		}
	}
    return nil
}

func main() {

	var err error

	c, err = redis.Dial("tcp", REDIS_ADDRESS)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	log.Print("Successfully dialed Redis database")

	auth_result, err := c.Do("AUTH", REDIS_PASSWORD)

	if err != nil {
		panic(err)
	}

	if auth_result == "OK" {
		log.Print("Successfully authenticated Redis database")
	}

	log.Print("Successfully created Redis database connection")

	anaconda.SetConsumerKey(TWITTER_CONSUMER_KEY)
	anaconda.SetConsumerSecret(TWITTER_CONSUMER_SECRET)
	api := anaconda.NewTwitterApi(TWITTER_ACCESS_TOKEN, TWITTER_ACCESS_TOKEN_SECRET)

	for {
		checkForTweets(api)
		log.Printf("Sleeping for %d seconds", SLEEP_INTERVAL)
		time.Sleep(SLEEP_INTERVAL * time.Second)
	}

}
