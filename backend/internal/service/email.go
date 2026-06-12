package service

import (
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"net/smtp"
	"strings"
	"time"

	"github.com/instaagrammeta/somon-crm/backend/internal/config"
)

// EmailService is a tiny SMTP sender used for password resets, lead
// follow-ups, and any outbound transactional mail. The IMAP side (parsing
// inbound replies into the CRM thread) is intentionally left as a Receive()
// stub so a small worker can be added later without touching the surface.
type EmailService struct {
	cfg config.EmailConfig
}

func NewEmailService(cfg config.EmailConfig) *EmailService {
	return &EmailService{cfg: cfg}
}

func (s *EmailService) Enabled() bool {
	return s != nil && s.cfg.Enabled && s.cfg.SMTPHost != "" && s.cfg.From != ""
}

// SendMessage delivers a single text + optional HTML message to one recipient.
// HTML is sent as the alternative body if non-empty; otherwise a plain-text
// message is built. Multipart attachments are not required for the v-2 set —
// when needed, swap this function for a real MIME builder (gomail, etc.).
func (s *EmailService) SendMessage(to, subject, text, html string) error {
	if !s.Enabled() {
		return errors.New("email.disabled")
	}
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.SMTPHost)
	body := s.buildBody(to, subject, text, html)

	if s.cfg.UseTLS {
		return sendTLS(addr, s.cfg.SMTPHost, auth, s.cfg.From, []string{to}, body)
	}
	return smtp.SendMail(addr, auth, s.cfg.From, []string{to}, body)
}

func (s *EmailService) buildBody(to, subject, text, html string) []byte {
	subj := mime.QEncoding.Encode("utf-8", subject)
	headers := []string{
		"From: " + s.cfg.From,
		"To: " + to,
		"Subject: " + subj,
		"Date: " + time.Now().UTC().Format(time.RFC1123Z),
		"MIME-Version: 1.0",
	}
	if html != "" {
		boundary := "somon-bnd-" + fmt.Sprintf("%d", time.Now().UnixNano())
		headers = append(headers, `Content-Type: multipart/alternative; boundary="`+boundary+`"`)
		var b strings.Builder
		b.WriteString(strings.Join(headers, "\r\n"))
		b.WriteString("\r\n\r\n")
		// plain text part
		b.WriteString("--" + boundary + "\r\n")
		b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		b.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
		b.WriteString(text)
		b.WriteString("\r\n\r\n")
		// html part
		b.WriteString("--" + boundary + "\r\n")
		b.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		b.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
		b.WriteString(html)
		b.WriteString("\r\n\r\n--" + boundary + "--\r\n")
		return []byte(b.String())
	}
	headers = append(headers, "Content-Type: text/plain; charset=UTF-8")
	return []byte(strings.Join(headers, "\r\n") + "\r\n\r\n" + text)
}

// sendTLS performs an explicit STARTTLS-on-connect handshake. SMTP submission
// on port 465 typically requires this.
func sendTLS(addr, host string, auth smtp.Auth, from string, to []string, body []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
	if err != nil {
		return err
	}
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer c.Close()
	if err := c.Auth(auth); err != nil {
		return err
	}
	if err := c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err := c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(body); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

// IncomingMail is the future-facing IMAP shape: when the receive worker is
// added it will parse messages into this struct and pipe them into the CRM
// chat thread (or convert them to a lead).
type IncomingMail struct {
	From    string
	To      string
	Subject string
	Body    string
	HTML    string
	At      time.Time
}

// Receive is a placeholder for the IMAP fetcher; wire to go-imap when needed.
func (s *EmailService) Receive() ([]IncomingMail, error) {
	if !s.Enabled() || s.cfg.IMAPHost == "" {
		return nil, errors.New("email.imap_disabled")
	}
	return nil, errors.New("email.imap_not_implemented")
}
