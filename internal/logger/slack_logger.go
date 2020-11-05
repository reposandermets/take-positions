package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

type SlackRequestBody struct {
	Text    string `json:"text"`
	Channel string `json:"channel"`
}

func SendSlackNotification(msg interface{}) error {
	println(fmt.Sprintf("%v", msg))

	// TODO make it gopher
	webhookUrl := viper.GetString("SLACK_URL")
	channel := viper.GetString("SLACK_CHANNEL")

	slackRequestBody := SlackRequestBody{Text: "take-positions: " + fmt.Sprintf("%v", msg), Channel: "#" + channel}

	slackBody, _ := json.Marshal(slackRequestBody)
	req, err := http.NewRequest(http.MethodPost, webhookUrl, bytes.NewBuffer(slackBody))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		println("######### START UNABLE TO SEND TO SLACK")
		println(fmt.Sprintf("%v", msg))
		println("######### END UNABLE TO SEND TO SLACK")
		return errors.New("Non-ok response returned from Slack")
	}
	return nil
}
