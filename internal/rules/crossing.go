package rules

import "github.com/Secretstar513/crypto-alerts/internal/domain"

// Crosses returns true if price crosses threshold in the alert's direction.
// Requires previous price (may be 0 if unknown).
func Crosses(prev, current float64, a *domain.Alert) bool {
	if prev == 0 {
		return false // need a baseline
	}
	switch a.Direction {
	case domain.DirectionUp:
		// crossing up: prev < thr && current >= thr
		return prev < a.Threshold && current >= a.Threshold
	case domain.DirectionDown:
		// crossing down: prev > thr && current <= thr
		return prev > a.Threshold && current <= a.Threshold
	default:
		return false
	}
}
