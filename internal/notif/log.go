package notif

import (
	"context"

	"github.com/rs/zerolog/log"
)

type LogNotifier struct{ enabled bool }

func NewLog(enabled bool) *LogNotifier { return &LogNotifier{enabled: enabled} }
func (n *LogNotifier) Name() string    { return "log" }
func (n *LogNotifier) Enabled() bool   { return n.enabled }

func (n *LogNotifier) Notify(ctx context.Context, ev Event) error {
	if !n.enabled { return nil }
	log.Info().
		Str("notifier", "log").
		Str("symbol", ev.Symbol).
		Float64("price", ev.Price).
		Float64("threshold", ev.Threshold).
		Str("direction", ev.Direction).
		Msg("ALERT")
	return nil
}
