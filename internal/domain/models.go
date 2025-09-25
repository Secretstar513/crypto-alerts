package domain

import "time"

type Direction string

const (
	DirectionUp   Direction = "UP"
	DirectionDown Direction = "DOWN"
)

type Alert struct {
	ID        string    `gorm:"primaryKey"`
	Symbol    string    // e.g., BTCUSDT
	Threshold float64   // price boundary
	Direction Direction // UP or DOWN
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ChannelKind string

const (
	ChannelLog      ChannelKind = "LOG"
	ChannelEmail    ChannelKind = "EMAIL"
	ChannelTelegram ChannelKind = "TELEGRAM"
)

type Channel struct {
	ID        string      `gorm:"primaryKey"`
	Kind      ChannelKind // LOG | EMAIL | TELEGRAM
	Enabled   bool
	Config    string      // JSON blob per channel
	CreatedAt time.Time
	UpdatedAt time.Time
}

type LastPrice struct {
	Symbol    string  `gorm:"primaryKey"`
	Price     float64
	UpdatedAt time.Time
}
