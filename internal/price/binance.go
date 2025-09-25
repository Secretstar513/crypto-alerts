package price

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"nhooyr.io/websocket"
)

type binanceTicker struct {
	C string `json:"c"` // last price as string
}

func wsURL(symbol string) string {
	return fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@ticker", lower(symbol))
}
func lower(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 32
		}
	}
	return string(b)
}

func streamBinance(ctx context.Context, symbol string, out chan<- float64) error {
	c, _, err := websocket.Dial(ctx, wsURL(symbol), nil)
	if err != nil {
		return err
	}
	defer c.Close(websocket.StatusNormalClosure, "bye")

	for {
		_, data, err := c.Read(ctx)
		if err != nil {
			return err
		}
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			continue
		}
		if cval, ok := raw["c"].(string); ok {
			if p, err := parseFloat(cval); err == nil {
				select {
				case out <- p:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscan(s, &f)
	return f, err
}
