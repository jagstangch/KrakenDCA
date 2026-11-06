package main

import (
	"runtime"
	"time"

	"github.com/primexz/KrakenDCA/bot"
	"github.com/primexz/KrakenDCA/metrics"
	log "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

func init() {
	log.SetFormatter(&prefixed.TextFormatter{
		TimestampFormat:  "2006/01/02 - 15:04:05",
		FullTimestamp:    true,
		QuoteEmptyFields: true,
		SpacePadding:     45,
	})

	log.SetReportCaller(true)
}

func main() {
	log.WithFields(log.Fields{
		"commit":  commit,
		"runtime": runtime.Version(),
		"arch":    runtime.GOARCH,
	}).Infof("Kraken DCA 🐙 %s modded by Jag, built %s", version, buildDate())

	metrics.StartServer()

	bot.NewBot().StartBot()
}

// buildDate returns the build date set via ldflags in the same format as the log timestamps
func buildDate() string {
	if t, err := time.Parse(time.RFC3339, date); err == nil {
		return t.UTC().Format("2006/01/02 - 15:04:05") + " UTC"
	}
	return date
}
