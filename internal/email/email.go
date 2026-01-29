package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config holds SMTP email configuration
type Config struct {
	SmtpServer string
	SmtpPort   int
	Username   string
	Password   string
	To         string
}

// Summary holds execution result summary
type Summary struct {
	Success int
	Fail    int
	Total   int
}

// SendResultEmail sends an email with the ZIP file attached
func SendResultEmail(cfg Config, zipPath string, taskName string, summary Summary) error {
	from := cfg.Username
	to := strings.Split(cfg.To, ",")
	for i := range to {
		to[i] = strings.TrimSpace(to[i])
	}

	date := time.Now().Format("2006-01-02")
	subject := fmt.Sprintf("[AutoLogCollector] %s - %s", taskName, date)

	body := fmt.Sprintf("Schedule: %s\nDate: %s\n\nResult: %d success, %d failed (total %d)\n",
		taskName, date, summary.Success, summary.Fail, summary.Total)

	// Build MIME message
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Headers
	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=%s\r\n\r\n",
		from, strings.Join(to, ", "), subject, writer.Boundary())
	buf.Reset()
	buf.WriteString(headers)

	// Rewrite with boundary
	writer2 := multipart.NewWriter(&buf)
	// Set same boundary
	buf.Reset()
	writer2 = multipart.NewWriter(&buf)

	headers = fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=%s\r\n\r\n",
		from, strings.Join(to, ", "), subject, writer2.Boundary())

	var msg bytes.Buffer
	msg.WriteString(headers)

	// Text part
	textHeader := make(textproto.MIMEHeader)
	textHeader.Set("Content-Type", "text/plain; charset=utf-8")
	textPart, err := writer2.CreatePart(textHeader)
	if err != nil {
		return fmt.Errorf("failed to create text part: %w", err)
	}
	textPart.Write([]byte(body))

	// Attachment
	zipData, err := os.ReadFile(zipPath)
	if err != nil {
		return fmt.Errorf("failed to read zip file: %w", err)
	}

	attachHeader := make(textproto.MIMEHeader)
	attachHeader.Set("Content-Type", "application/zip")
	attachHeader.Set("Content-Transfer-Encoding", "base64")
	attachHeader.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(zipPath)))
	attachPart, err := writer2.CreatePart(attachHeader)
	if err != nil {
		return fmt.Errorf("failed to create attachment part: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(zipData)
	// Write in 76-char lines
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		attachPart.Write([]byte(encoded[i:end] + "\r\n"))
	}

	writer2.Close()

	msg.Write(buf.Bytes())

	// Send
	addr := fmt.Sprintf("%s:%d", cfg.SmtpServer, cfg.SmtpPort)
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SmtpServer)

	err = smtp.SendMail(addr, auth, from, to, msg.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
