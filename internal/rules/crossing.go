package rules

import "github.com/Secretstar513/crypto-alerts/internal/domain"

func Crosses(prev, current float64, a *domain.Alert) bool {
	if prev == 0 {
		return false
	}
	switch a.Direction {
	case domain.DirectionUp:
		return prev < a.Threshold && current >= a.Threshold
	case domain.DirectionDown:
		return prev > a.Threshold && current <= a.Threshold
	default:
		return false
	}
}
