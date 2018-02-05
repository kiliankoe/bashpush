package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/SlyMarbo/rss"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"github.com/sideshow/apns2/payload"
)

func debuglog(msg string) {
	if os.Getenv("BASHPUSH_DEBUG") == "1" {
		fmt.Println(msg)
	}
}

func main() {
	quote, link, id := getLastQuote()
	if isNew(id) {
		err := ioutil.WriteFile("lastquote.txt", []byte(id), 0644)
		if err == nil {
			// Panicking on the error but sending the notifications nonetheless
			// would spam pushover and slack every minute until I fix it due
			// to cron. Let's not do that.
			sendSlackNotification(id, link)
			sendiOSNotifications(id, quote)
		} else {
			debuglog(fmt.Sprintf("Couldn't write lastquote.txt, exiting. Err: %s", err))
		}
	} else {
		debuglog("Quote isn't new, exiting")
	}
}

func getLastQuote() (string, string, string) {

	bashFeedURL := os.Getenv("BASH_FEED_URL")
	if bashFeedURL == "" {
		panic("BASH_FEED_URL not set")
	}

	feed, err := rss.Fetch(bashFeedURL)
	if err != nil {
		panic(err)
	}

	r, _ := regexp.Compile(`\d+`)
	lastQuote := feed.Items[0]
	content := lastQuote.Content
	content = html.UnescapeString(content)

	debuglog(fmt.Sprintf("Found %s", lastQuote.Link))
	return content, lastQuote.Link, r.FindString(lastQuote.Link)
}

func isNew(id string) bool {
	if _, err := os.Stat("lastquote.txt"); os.IsNotExist(err) {
		// lastquote.txt does not yet exist
		err := ioutil.WriteFile("lastquote.txt", []byte{}, 0644)
		if err != nil {
			panic(err)
		}
	}
	last, err := ioutil.ReadFile("lastquote.txt")
	if err != nil {
		panic(err)
	}
	return id != strings.TrimSpace(string(last))
}

func sendiOSNotifications(id, quote string) {
	debuglog("Starting sending of iOS notifications")
	cert, err := certificate.FromP12File("./bashfsrPush.p12", "bashpush")
	if err != nil {
		log.Fatal(err)
	}

	devicesStr := os.Getenv("IOS_PUSH_TOKENS")
	devices := strings.Split(devicesStr, "\n")

	client := apns2.NewClient(cert).Development()

	for i := 0; i < len(devices); i++ {
		notification := &apns2.Notification{}
		notification.DeviceToken = devices[i]

		debuglog(fmt.Sprintf("Sending notification to %s", devices[i]))

		payload := payload.NewPayload()
		payload.AlertBody(quote)
		payload.Badge(1)
		payload.Sound("default")
		payload.Category("quote_category")
		payload.Custom("quote_id", id)
		notification.Payload = payload

		res, err := client.Push(notification)

		if err != nil {
			log.Fatal(err)
		}

		debuglog(fmt.Sprintf("%v %v %v\n", res.StatusCode, res.ApnsID, res.Reason))
	}
}

func sendSlackNotification(id, link string) {
	debuglog("Starting sending of Slack notification")
	hc := http.Client{}

	slackHookURL := os.Getenv("SLACK_HOOK_URL")
	if slackHookURL == "" {
		log.Fatal("SLACK_HOOK_URL not set")
	}

	quoteString := fmt.Sprintf("Neues Bash Zitat -> <%s|%s> 🤐", link, id)
	payload := map[string]string{"text": quoteString}
	jsonPayload, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", slackHookURL, bytes.NewBufferString(string(jsonPayload)))
	resp, err := hc.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatal("Post to Slack failed /o\\")
	} else {
		debuglog("Finished sending of Slack notification without errors (I think)")
	}
}
