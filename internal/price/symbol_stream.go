package price

import (
	"context"
	"sync"
	"time"
)

type symbolStream struct {
	symbol string
	subs   map[Subscriber]struct{}
	mu     sync.Mutex
	cancel context.CancelFunc
}

func newSymbolStream(symbol string) *symbolStream {
	return &symbolStream{
		symbol: symbol,
		subs:   map[Subscriber]struct{}{},
	}
}

func (s *symbolStream) add(sub Subscriber) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subs[sub] = struct{}{}
}

func (s *symbolStream) stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *symbolStream) run(parent context.Context) {
	ctx, cancel := context.WithCancel(parent)
	s.cancel = cancel

	out := make(chan float64, 8)

	go func() {
		for {
			if err := streamBinance(ctx, s.symbol, out); err != nil {
				_ = pollHTTP(ctx, s.symbol, out)
				select {
				case <-ctx.Done():
					return
				case <-time.After(30 * time.Second):
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case p := <-out:
			s.mu.Lock()
			for ch := range s.subs {
				select { case ch <- Update{Symbol: s.symbol, Price: p}:
				default:
				}
			}
			s.mu.Unlock()
		}
	}
}
