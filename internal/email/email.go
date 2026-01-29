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
	subject := fmt.Sprintf("[AutoLogCollector] %s — %d success, %d failed (%s)", taskName, summary.Success, summary.Fail, date)

	statusColor := "#22c55e"
	statusText := "ALL SUCCESS"
	if summary.Fail > 0 && summary.Success > 0 {
		statusColor = "#f59e0b"
		statusText = "PARTIAL FAILURE"
	} else if summary.Fail > 0 {
		statusColor = "#ef4444"
		statusText = "FAILED"
	}

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="margin:0;padding:0;background:#f4f4f5;font-family:'Segoe UI',Arial,sans-serif;">
<table width="100%%" cellpadding="0" cellspacing="0" style="background:#f4f4f5;padding:32px 0;">
<tr><td align="center">
<table width="520" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 1px 3px rgba(0,0,0,0.1);">
  <tr>
    <td style="background:#18181b;padding:24px 32px;">
      <span style="color:#ffffff;font-size:18px;font-weight:700;letter-spacing:-0.3px;">AutoLogCollector</span>
    </td>
  </tr>
  <tr>
    <td style="padding:28px 32px 12px;">
      <span style="display:inline-block;background:%s;color:#fff;font-size:12px;font-weight:700;padding:4px 12px;border-radius:20px;letter-spacing:0.5px;">%s</span>
    </td>
  </tr>
  <tr>
    <td style="padding:16px 32px 24px;">
      <table width="100%%" cellpadding="0" cellspacing="0" style="border-collapse:collapse;">
        <tr>
          <td style="padding:12px 0;border-bottom:1px solid #e4e4e7;color:#71717a;font-size:13px;width:100px;">Schedule</td>
          <td style="padding:12px 0;border-bottom:1px solid #e4e4e7;color:#18181b;font-size:15px;font-weight:600;">%s</td>
        </tr>
        <tr>
          <td style="padding:12px 0;border-bottom:1px solid #e4e4e7;color:#71717a;font-size:13px;">Date</td>
          <td style="padding:12px 0;border-bottom:1px solid #e4e4e7;color:#18181b;font-size:15px;">%s</td>
        </tr>
        <tr>
          <td style="padding:12px 0;border-bottom:1px solid #e4e4e7;color:#71717a;font-size:13px;">Total</td>
          <td style="padding:12px 0;border-bottom:1px solid #e4e4e7;color:#18181b;font-size:15px;font-weight:600;">%d</td>
        </tr>
        <tr>
          <td style="padding:12px 0;border-bottom:1px solid #e4e4e7;color:#71717a;font-size:13px;">Success</td>
          <td style="padding:12px 0;border-bottom:1px solid #e4e4e7;color:#22c55e;font-size:15px;font-weight:600;">%d</td>
        </tr>
        <tr>
          <td style="padding:12px 0;color:#71717a;font-size:13px;">Failed</td>
          <td style="padding:12px 0;color:#ef4444;font-size:15px;font-weight:600;">%d</td>
        </tr>
      </table>
    </td>
  </tr>
  <tr>
    <td style="padding:0 32px 24px;">
      <p style="margin:0;color:#a1a1aa;font-size:12px;">Log files are attached as a ZIP archive.</p>
    </td>
  </tr>
</table>
</td></tr>
</table>
</body>
</html>`,
		statusColor, statusText,
		taskName, date,
		summary.Total, summary.Success, summary.Fail)

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
	textHeader.Set("Content-Type", "text/html; charset=utf-8")
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
