package price

import (
	"context"
	"sync"
)

type Update struct {
	Symbol string
	Price  float64
}

type Subscriber chan Update

type Router struct {
	mu       sync.Mutex
	streams  map[string]*symbolStream
}

func NewRouter() *Router {
	return &Router{streams: map[string]*symbolStream{}}
}

func (r *Router) Subscribe(ctx context.Context, symbol string) Subscriber {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.streams[symbol]
	if !ok {
		s = newSymbolStream(symbol)
		r.streams[symbol] = s
		go s.run(ctx)
	}
	ch := make(Subscriber, 16)
	s.add(ch)
	return ch
}

func (r *Router) StopAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.streams {
		s.stop()
	}
}
