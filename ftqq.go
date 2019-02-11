package main

import (
	"net/http"
	"net/url"
	"time"

	"github.com/morya/utils/log"
)

const (
	TIME_FORMAT = "2006-01-02 15:04:05"
)

func getTimeStamp() string {
	t := time.Now().Format(TIME_FORMAT)
	return t
}


var lastAlert time.Time

func fangtangNotify(title, content string) {
	now := time.Now()
	if now.Sub(lastAlert).Seconds() < 1200 {
		log.Info("prehibit warn from last call")
		return
	}
	lastAlert = now

	ftqqURL := *flagAlertUrl

	ts := getTimeStamp()
	postData := url.Values{}
	postData.Add("text", title)
	postData.Add("desp", ts+" "+content)

	c := new(http.Client)

	log.Debugf("ftqq url = %v", ftqqURL)
	resp, err := c.PostForm(ftqqURL, postData)
	if err != nil {
		log.ErrorError(err, "call FTQQ http api failed")
		return
	}
	resp.Body.Close()
}
