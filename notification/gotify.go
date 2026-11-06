package notification

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

var logger *log.Entry

func init() {
	logger = log.WithFields(log.Fields{
		"prefix": "gotify",
	})
}

type GotifyMessage struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Priority int    `json:"priority"`
}

func gotifyConfigured() bool {
	return os.Getenv("GOTIFY_APP_TOKEN") != "" && os.Getenv("GOTIFY_URL") != ""
}

func SendPushNotification(title string, messageText string) {
	appToken := os.Getenv("GOTIFY_APP_TOKEN")
	rawUrl := os.Getenv("GOTIFY_URL")

	if appToken == "" || rawUrl == "" {
		logger.Debug("Skip sending push notification, not configured")
		return
	}

	reqUrl := url.URL{
		Host:   rawUrl,
		Path:   "/message",
		Scheme: "http",
	}

	q := reqUrl.Query()
	q.Set("token", appToken)
	reqUrl.RawQuery = q.Encode()

	message := &GotifyMessage{
		Title:    title,
		Message:  messageText,
		Priority: 5,
	}

	bodyBytes, err := json.Marshal(message)
	if err != nil {
		logger.Error(err)
		return
	}

	client := &http.Client{Timeout: 15 * time.Second}
	rsp, err := client.Post(reqUrl.String(), "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		logger.Error("failed to send push notification", err)
		return
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		logger.Error("failed to send push notification, status code", rsp.StatusCode)
	}
}
