package notif

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type EmailConfig struct {
	Host string
	Port string
	User string
	Pass string
	From string
	To   string
}

type EmailNotifier struct {
	cfg     EmailConfig
	enabled bool
}

func NewEmail(cfg EmailConfig, enabled bool) *EmailNotifier {
	return &EmailNotifier{cfg: cfg, enabled: enabled}
}

func (n *EmailNotifier) Name() string  { return "email" }
func (n *EmailNotifier) Enabled() bool { return n.enabled }

func (n *EmailNotifier) Notify(ctx context.Context, ev Event) error {
	if !n.enabled { return nil }
	if n.cfg.Host == "" || n.cfg.Port == "" || n.cfg.From == "" || n.cfg.To == "" {
		return nil
	}
	sub := fmt.Sprintf("[Crypto Alert] %s %s %.2f (thr=%.2f)", ev.Symbol, ev.Direction, ev.Price, ev.Threshold)
	body := fmt.Sprintf("Symbol: %s\nDirection: %s\nPrice: %.8f\nThreshold: %.8f\nTime: %s\n",
		ev.Symbol, ev.Direction, ev.Price, ev.Threshold, time.Now().Format(time.RFC3339))

	addr := net.JoinHostPort(n.cfg.Host, n.cfg.Port)
	msg := strings.Builder{}
	msg.WriteString("From: " + n.cfg.From + "\r\n")
	msg.WriteString("To: " + n.cfg.To + "\r\n")
	msg.WriteString("Subject: " + sub + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
	msg.WriteString(body)

	var auth smtp.Auth
	if n.cfg.User != "" || n.cfg.Pass != "" {
		auth = smtp.PlainAuth("", n.cfg.User, n.cfg.Pass, n.cfg.Host)
	}

	_ = smtp.SendMail(addr, auth, n.cfg.From, []string{n.cfg.To}, []byte(msg.String()))
	conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
	if err == nil {
		_ = conn.Close()
	}
	return nil
}
