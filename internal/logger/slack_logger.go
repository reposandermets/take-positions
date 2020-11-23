package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

type SlackRequestBody struct {
	Text    string `json:"text"`
	Channel string `json:"channel"`
}

func SendSlackNotification(message interface{}) {
	webhookUrl := viper.GetString("SLACK_URL")
	channel := viper.GetString("SLACK_CHANNEL")
	go func(msg interface{}) {
		currentTime := time.Now()
		println(fmt.Sprintf(currentTime.Format("2006-01-02 15:04:05 Mon")+" %v", msg))

		slackRequestBody := SlackRequestBody{Text: "take-positions " + currentTime.Format("2006-01-02 15:04:05 Mon") + " - " + fmt.Sprintf("%v", msg), Channel: "#" + channel}

		slackBody, _ := json.Marshal(slackRequestBody)
		req, err := http.NewRequest(http.MethodPost, webhookUrl, bytes.NewBuffer(slackBody))
		if err != nil {
			println("ERROR 1 sending slack" + err.Error())
		} else {
			req.Header.Add("Content-Type", "application/json")
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				println("ERROR 2 sending slack " + err.Error())
			} else {
				buf := new(bytes.Buffer)
				buf.ReadFrom(resp.Body)
				if buf.String() != "ok" {
					println("######### START UNABLE TO SEND TO SLACK")
					println(fmt.Sprintf("%v", msg))
					println("######### END UNABLE TO SEND TO SLACK")
				}
			}
		}
	}(message)
}
