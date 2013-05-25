package main

import (
	"github.com/ChimeraCoder/anaconda"
	"fmt"
    "os"
    "strings"
)

var TWITTER_CONSUMER_KEY = os.Getenv("TWITTER_CONSUMER_KEY")
var TWITTER_CONSUMER_SECRET = os.Getenv("TWITTER_CONSUMER_SECRET")
var TWITTER_ACCESS_TOKEN = os.Getenv("TWITTER_ACCESS_TOKEN")
var TWITTER_ACCESS_TOKEN_SECRET = os.Getenv("TWITTER_ACCESS_TOKEN_SECRET")

var TARGET_USERS = []string{"chimeracoder", "rubinovitz"}

func TweetMentions(tweet anaconda.Tweet, username string) bool {
    return strings.Contains(tweet.Text, "@" + username)
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


func main() {

	anaconda.SetConsumerKey(TWITTER_CONSUMER_KEY)
	anaconda.SetConsumerSecret(TWITTER_CONSUMER_SECRET)
	api := anaconda.NewTwitterApi(TWITTER_ACCESS_TOKEN, TWITTER_ACCESS_TOKEN_SECRET)

    searchResult, _ := api.GetHomeTimeline()
    for _ , tweet := range searchResult {
        if TweetMentionsATarget(tweet){
            fmt.Println(tweet.Text)
        }
        }   
}
