package notification

import (
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/primexz/KrakenDCA/config"
	log "github.com/sirupsen/logrus"
)

const (
	smtpTimeout = 30 * time.Second

	// port for SMTP over implicit TLS, all other ports use STARTTLS
	smtpsPort = 465
)

var emailLogger = log.WithFields(log.Fields{
	"prefix": "email",
})

func emailConfigured() bool {
	return config.C.SmtpHost != "" && len(config.C.EmailTo) > 0
}

// SendEmailNotification sends the notification as a plain text e-mail to all EMAIL_TO addresses
func SendEmailNotification(title string, messageText string) {
	if !emailConfigured() {
		emailLogger.Debug("Skip sending e-mail notification, not configured")
		return
	}

	from := config.C.EmailFrom
	if from == "" {
		from = config.C.SmtpUsername
	}

	message := buildEmail(from, config.C.EmailTo, "[Kraken DCA] "+title, messageText, time.Now())
	if err := sendEmail(from, config.C.EmailTo, message); err != nil {
		emailLogger.Error("failed to send e-mail notification: ", err)
	}
}

func buildEmail(from string, to []string, subject string, body string, date time.Time) []byte {
	var b strings.Builder

	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + strings.Join(to, ", ") + "\r\n")
	b.WriteString("Subject: " + mime.QEncoding.Encode("utf-8", subject) + "\r\n")
	b.WriteString("Date: " + date.Format(time.RFC1123Z) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	b.WriteString("\r\n")
	b.WriteString(strings.ReplaceAll(body, "\n", "\r\n") + "\r\n")

	return []byte(b.String())
}

func sendEmail(from string, to []string, message []byte) error {
	host := config.C.SmtpHost
	addr := net.JoinHostPort(host, strconv.Itoa(config.C.SmtpPort))
	tlsConfig := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}

	conn, err := net.DialTimeout("tcp", addr, smtpTimeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(smtpTimeout)); err != nil {
		return err
	}

	if config.C.SmtpPort == smtpsPort {
		conn = tls.Client(conn, tlsConfig)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	if config.C.SmtpPort != smtpsPort {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			// never send the password or the purchase details unencrypted
			return errors.New("smtp server does not support STARTTLS")
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return err
		}
	}

	if config.C.SmtpUsername != "" {
		if err := client.Auth(smtp.PlainAuth("", config.C.SmtpUsername, config.C.SmtpPassword, host)); err != nil {
			return fmt.Errorf("smtp authentication failed: %w", err)
		}
	}

	if err := client.Mail(from); err != nil {
		return err
	}
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(message); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return client.Quit()
}
