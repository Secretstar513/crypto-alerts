package notif

import "context"

type Event struct {
	Symbol    string
	Price     float64
	Threshold float64
	Direction string
}

type Notifier interface {
	Name() string
	Enabled() bool
	Notify(ctx context.Context, ev Event) error
}
