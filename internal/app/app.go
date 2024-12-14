package app

import (
	"context"
	"net/http"

	"github.com/ashep/d5y/internal/geoip"
	"github.com/ashep/d5y/internal/server"
	"github.com/ashep/d5y/internal/update"
	"github.com/ashep/d5y/internal/weather"
	"github.com/ashep/go-app/runner"
	"github.com/google/go-github/v63/github"
)

type App struct{}

func New(cfg Config, rt *runner.Runtime) (*App, error) {
	giSvc := geoip.New()
	wthSvc := weather.New(cfg.Weather.APIKey)
	ghCli := github.NewClient(http.DefaultClient).WithAuthToken(cfg.GitHub.Token)
	updSvc := update.New(ghCli, rt.Logger.With().Str("pkg", "update_svc").Logger())

	server.Setup(rt.SrvMux, giSvc, wthSvc, updSvc, rt.Logger)

	return &App{}, nil
}

func (a *App) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
