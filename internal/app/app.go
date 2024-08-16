package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/go-github/v63/github"
	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/geoip"
	"github.com/ashep/d5y/internal/server"
	"github.com/ashep/d5y/internal/update"
	"github.com/ashep/d5y/internal/weather"
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
	giSvc := geoip.New()
	wthSvc := weather.New(a.cfg.Weather.APIKey)
	ghCli := github.NewClient(http.DefaultClient).WithAuthToken(a.cfg.GitHub.Token)
	updSvc := update.New(ghCli, a.l.With().Str("pkg", "update_svc").Logger())

	s := server.New(a.cfg.Server.Addr, giSvc, wthSvc, updSvc, a.l)
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
