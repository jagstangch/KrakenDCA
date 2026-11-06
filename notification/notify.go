package notification

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

// Notify sends a notification to every configured channel (Gotify, e-mail)
func Notify(title string, messageText string) {
	SendPushNotification(title, messageText)
	SendEmailNotification(title, messageText)
}

// LogChannels logs which notification channels are configured
func LogChannels() {
	channels := []string{}
	if gotifyConfigured() {
		channels = append(channels, "gotify")
	}
	if emailConfigured() {
		channels = append(channels, "e-mail")
	}

	if len(channels) == 0 {
		log.WithField("prefix", "notification").Info("No notification channel configured")
		return
	}

	log.WithFields(log.Fields{
		"prefix":   "notification",
		"channels": strings.Join(channels, ", "),
	}).Info("Notifications enabled")
}
