package price

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type httpTicker struct{ Price string `json:"price"` }

func pollHTTP(ctx context.Context, symbol string, out chan<- float64) error {
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s", symbol)
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()

	for {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err := http.DefaultClient.Do(req)
		if err == nil && resp.StatusCode == 200 {
			var v httpTicker
			if json.NewDecoder(resp.Body).Decode(&v) == nil {
				if p, err := parseFloat(v.Price); err == nil {
					select {
					case out <- p:
					case <-ctx.Done():
						resp.Body.Close()
						return ctx.Err()
					}
				}
			}
			resp.Body.Close()
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}
	}
}
