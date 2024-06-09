package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/server"
)

type App struct {
	cfg Config
	l   zerolog.Logger
}

func New(cfg Config, l zerolog.Logger) *App {
	if cfg.Server.Addr == "" {
		cfg.Server.Addr = ":9000"
	}

	return &App{
		cfg: cfg,
		l:   l,
	}
}

func (a *App) Run(ctx context.Context, _ []string) error {
	s := server.New(a.cfg.Server.Addr, a.cfg.Weather.APIKey, a.l)

	done := make(chan struct{})

	go func() {
		err := s.Run()
		if !errors.Is(err, http.ErrServerClosed) {
			a.l.Error().Err(err).Msg("server run failed")
		}

		done <- struct{}{}
	}()

	select {
	case <-done:
		// do nothing
	case <-ctx.Done():
		sdCtx, sdCtxC := context.WithTimeout(context.Background(), time.Second*5)
		defer sdCtxC()

		s.Shutdown(sdCtx) //nolint:contextcheck // ok
	}

	return nil
}
