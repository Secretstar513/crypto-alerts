package notif

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type TelegramNotifier struct {
	botToken string
	chatID   string
	enabled  bool
}

func NewTelegram(botToken, chatID string, enabled bool) *TelegramNotifier {
	return &TelegramNotifier{botToken: botToken, chatID: chatID, enabled: enabled}
}
func (n *TelegramNotifier) Name() string  { return "telegram" }
func (n *TelegramNotifier) Enabled() bool { return n.enabled }

func (n *TelegramNotifier) Notify(ctx context.Context, ev Event) error {
	if !n.enabled || n.botToken == "" || n.chatID == "" {
		return nil
	}
	text := url.QueryEscape(fmt.Sprintf("ALERT %s %s @ %.8f (thr %.8f)", ev.Symbol, ev.Direction, ev.Price, ev.Threshold))
	api := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s", n.botToken, n.chatID, text)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, api, nil)
	_, _ = http.DefaultClient.Do(req)
	return nil
}
