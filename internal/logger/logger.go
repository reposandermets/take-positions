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
	Content string `json:"content"`
}

func SendLoggerNotification(message interface{}) {
	webhookUrl := viper.GetString("LOGGER_URL")
	loggerEnabled := viper.GetBool("LOGGER_ENABLED")
	currentTime := time.Now()
	content := currentTime.Format("2006-01-02 15:04:05 Mon") + " - " + fmt.Sprintf("%v", message)
	println(content)
	if !loggerEnabled {
		return
	}

	go func(msg string) {
		slackRequestBody := SlackRequestBody{Content: msg}

		slackBody, _ := json.Marshal(slackRequestBody)
		req, err := http.NewRequest(http.MethodPost, webhookUrl, bytes.NewBuffer(slackBody))
		if err != nil {
			println("ERROR preparing logger payload" + err.Error())
		} else {
			req.Header.Add("Content-Type", "application/json")
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			if resp.StatusCode == 204 && err == nil {
				return
			}
			println("ERROR logger http " + err.Error())
		}
	}(content)
}
