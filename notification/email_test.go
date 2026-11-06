package notification

import (
	"strings"
	"testing"
	"time"
)

func TestBuildEmail(t *testing.T) {
	date := time.Date(2026, 11, 6, 7, 17, 51, 0, time.UTC)
	message := string(buildEmail(
		"bot@example.com",
		[]string{"me@example.com", "you@example.com"},
		"[Kraken DCA] Bought btc ₿",
		"Pair: XXBTZEUR\nFee: 0.0148 EUR (0.40%, maker)",
		date,
	))

	headers, body, found := strings.Cut(message, "\r\n\r\n")
	if !found {
		t.Fatalf("no blank line between headers and body:\n%s", message)
	}

	for _, want := range []string{
		"From: bot@example.com",
		"To: me@example.com, you@example.com",
		"Subject: =?utf-8?q?",
		"Date: Fri, 06 Nov 2026 07:17:51 +0000",
		"Content-Type: text/plain; charset=utf-8",
	} {
		if !strings.Contains(headers, want) {
			t.Errorf("headers missing %q:\n%s", want, headers)
		}
	}

	if body != "Pair: XXBTZEUR\r\nFee: 0.0148 EUR (0.40%, maker)\r\n" {
		t.Errorf("unexpected body %q", body)
	}

	if strings.Contains(strings.ReplaceAll(message, "\r\n", ""), "\n") {
		t.Error("message contains bare LF line endings")
	}
}
