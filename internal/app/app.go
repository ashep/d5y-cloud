package app

import (
	"context"
	"net/http"

	handlerNotFound "github.com/ashep/d5y/internal/api/notfound"
	handlerV1 "github.com/ashep/d5y/internal/api/v1"
	handlerV2 "github.com/ashep/d5y/internal/api/v2"
	"github.com/ashep/d5y/internal/clientinfo"
	"github.com/ashep/d5y/internal/update"
	"github.com/ashep/d5y/internal/weatherapi"
	"github.com/ashep/go-app/runner"
	"github.com/google/go-github/v63/github"
	"github.com/rs/zerolog"
)

type App struct{}

func New(cfg Config, rt *runner.Runtime) (*App, error) {
	setupServer(cfg, rt.SrvMux, rt.Logger)
	return &App{}, nil
}

func (a *App) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func setupServer(cfg Config, mux *http.ServeMux, l zerolog.Logger) {
	weatherSvc := weatherapi.New(cfg.Weather.APIKey)
	githubCli := github.NewClient(http.DefaultClient).WithAuthToken(cfg.GitHub.Token)
	updSvc := update.New(githubCli, l.With().Str("pkg", "update_svc").Logger())

	lv1 := l.With().Str("pkg", "v1_handler").Logger()
	hv1 := handlerV1.New(weatherSvc, l)
	mux.HandleFunc("/api/1", wrapMiddlewares(hv1.Handle, lv1)) // BC

	logV2 := l.With().Str("pkg", "v2_handler").Logger()
	hdlV2 := handlerV2.New(weatherSvc, updSvc, logV2)
	mux.Handle("/v2/time", wrapMiddlewares(hdlV2.HandleTime, logV2))
	mux.Handle("/v2/weather", wrapMiddlewares(hdlV2.HandleWeather, logV2))
	mux.Handle("/v2/firmware/update", wrapMiddlewares(hdlV2.HandleUpdate, logV2))

	l404 := l.With().Str("pkg", "not_found_handler").Logger()
	h404 := handlerNotFound.New(l404)
	mux.HandleFunc("/", wrapMiddlewares(h404.Handle, l404))
}

func wrapMiddlewares(h http.HandlerFunc, l zerolog.Logger) http.HandlerFunc {
	h = clientinfo.WrapHTTP(h, l)
	return h
}
