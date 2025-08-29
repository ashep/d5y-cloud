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

type App struct {
	rt *runner.Runtime
	l  zerolog.Logger
}

func New(cfg *Config, rt *runner.Runtime) (*App, error) {
	l := rt.Logger

	weatherSvc := weatherapi.New(cfg.Weather.APIKey)
	githubCli := github.NewClient(http.DefaultClient).WithAuthToken(cfg.GitHub.Token)
	updSvc := update.New(githubCli, l.With().Str("pkg", "update_svc").Logger())

	logV1 := l.With().Str("pkg", "v1_handler").Logger()
	hdlV1 := handlerV1.New(weatherSvc, logV1)
	rt.Server.HandleFunc("/api/1", wrapMiddlewares(hdlV1.Handle, logV1)) // BC

	logV2 := l.With().Str("pkg", "v2_handler").Logger()
	hdlV2 := handlerV2.New(weatherSvc, updSvc, logV2)
	rt.Server.Handle("/v2/time", wrapMiddlewares(hdlV2.HandleTime, logV2))
	rt.Server.Handle("/v2/weather", wrapMiddlewares(hdlV2.HandleWeather, logV2))
	rt.Server.Handle("/v2/firmware/update", wrapMiddlewares(hdlV2.HandleUpdate, logV2))

	log404 := l.With().Str("pkg", "404_handler").Logger()
	hdl404 := handlerNotFound.New(log404)
	rt.Server.HandleFunc("/", wrapMiddlewares(hdl404.Handle, log404))

	return &App{
		rt: rt,
		l:  l,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.l.Info().Str("addr", a.rt.Server.Listener().Addr().String()).Msg("starting server")
	return <-a.rt.Server.Start(ctx)
}

func wrapMiddlewares(h http.HandlerFunc, l zerolog.Logger) http.HandlerFunc {
	h = clientinfo.WrapHTTP(h, l)
	return h
}
