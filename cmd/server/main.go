package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/Secretstar513/crypto-alerts/internal/app"
	"github.com/Secretstar513/crypto-alerts/internal/config"
	"github.com/Secretstar513/crypto-alerts/internal/server"
)

func main() {
	cfg := config.Load()
	a := app.New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a.Start(ctx)

	h := server.NewHandlers(a)
	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: server.Routes(h),
	}

	go func() {
		log.Info().Str("addr", cfg.Addr).Msg("server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
	log.Info().Msg("shutting down...")
	_ = srv.Shutdown(context.Background())
	a.Stop()
	time.Sleep(300 * time.Millisecond)
}
