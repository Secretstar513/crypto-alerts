package domain

import (
	"errors"
	"strings"
)

func ValidateAlert(a *Alert) error {
	if a.Symbol == "" {
		return errors.New("symbol required")
	}
	if strings.ToUpper(a.Symbol) != a.Symbol {
		return errors.New("symbol must be uppercase, e.g., BTCUSDT")
	}
	if a.Threshold <= 0 {
		return errors.New("threshold must be > 0")
	}
	if a.Direction != DirectionUp && a.Direction != DirectionDown {
		return errors.New("direction must be UP or DOWN")
	}
	return nil
}
