package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	TIME_FORMAT = "2006-01-02 15:04:05"
)

type H map[string]interface{}

func getTimeStamp() string {
	t := time.Now().Format(TIME_FORMAT)
	return t
}

func DumpObject(obj interface{}) []byte {
	buff := &bytes.Buffer{}
	enc := json.NewEncoder(buff)
	enc.Encode(obj)
	return buff.Bytes()
}

var lastAlert time.Time

func fangtangNotify(title, content string) {

	ftqqURL := *flagAlertUrl

	ts := getTimeStamp()
	postData := url.Values{}
	postData.Add("text", title)
	postData.Add("desp", ts+" "+content)

	c := new(http.Client)

	logrus.Debugf("ftqq url = %v", ftqqURL)
	resp, err := c.PostForm(ftqqURL, postData)
	if err != nil {
		logrus.ErrorError(err, "call FTQQ http api failed")
		return
	}
	resp.Body.Close()
}

func couldAlert() bool {
	now := time.Now()
	if now.Sub(lastAlert).Seconds() < 1200 {
		return false
	}
	lastAlert = now
	return true
}

func dingtalkNotify(title, content string) error {
	alertUrl := *flagAlertUrl
	logrus.Debugf("dingtalk alert url = %v", alertUrl)

	ts := getTimeStamp()

	var data = H{}
	data["msgtype"] = "text"
	data["text"] = H{"content": ts + content}
	bindata := bytes.NewReader(DumpObject(data))

	c := new(http.Client)
	resp, err := c.Post(alertUrl, "application/json", bindata)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func sendAlert(title, content string) {
	if !couldAlert() {
		return
	}
	dingtalkNotify(title, content)
}
