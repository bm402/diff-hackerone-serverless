package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"
)

// SlackRequestBody contains data about the slack notification to be sent
type SlackRequestBody struct {
	Channel   string `json:"channel"`
	Username  string `json:"username"`
	Text      string `json:"text"`
	IconEmoji string `json:"icon_emoji"`
}

func sendSlackNotification(channel string, message string, isPaid bool) {
	slackRequestBody := SlackRequestBody{
		Channel:   "#" + channel,
		Username:  "diff-hackerone",
		Text:      message,
		IconEmoji: ":moneybag:",
	}
	if isPaid == false {
		slackRequestBody.IconEmoji = ":ghost:"
	}

	sendRequestToSlack(slackRequestBody)
}

func sendSlackErrorNotification(err error) {
	slackRequestBody := SlackRequestBody{
		Channel:   "#errors",
		Username:  "diff-hackerone",
		Text:      err.Error(),
		IconEmoji: ":x:",
	}

	sendRequestToSlack(slackRequestBody)
}

func sendRequestToSlack(slackRequestBody SlackRequestBody) {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		logger("No slack notification sent due to unset SLACK_WEBHOOK_URL environment variable")
		return
	}

	body, _ := json.Marshal(slackRequestBody)
	request, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(body))
	if err != nil {
		logger("No slack notification sent due to request setup error")
		return
	}
	request.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		logger("No slack notification sent due to request error")
		return
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	if buf.String() != "ok" {
		logger("Non-OK response returned from Slack")
		return
	}

	logger("Slack notification sent")
}
