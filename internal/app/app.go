package app

import (
	"context"
	"net/http"

	"github.com/ashep/d5y/internal/auth"
	"github.com/ashep/d5y/internal/geoip"
	"github.com/ashep/d5y/internal/httputil"
	"github.com/ashep/d5y/internal/remoteaddr"
	handlerNotFound "github.com/ashep/d5y/internal/rpc/notfound"
	handlerV1 "github.com/ashep/d5y/internal/rpc/v1"
	handlerV2 "github.com/ashep/d5y/internal/rpc/v2"
	"github.com/ashep/d5y/internal/update"
	"github.com/ashep/d5y/internal/weather"
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
	geoIPSvc := geoip.New()
	weatherSvc := weather.New(cfg.Weather.APIKey)
	githubCli := github.NewClient(http.DefaultClient).WithAuthToken(cfg.GitHub.Token)
	updSvc := update.New(githubCli, l.With().Str("pkg", "update_svc").Logger())

	lv1 := l.With().Str("pkg", "v1_handler").Logger()
	hv1 := handlerV1.New(geoIPSvc, weatherSvc, l)
	mux.HandleFunc("/api/1", wrapMiddlewares(hv1.Handle, geoIPSvc, lv1)) // BC

	logV2 := l.With().Str("pkg", "v2_handler").Logger()
	hdlV2 := handlerV2.New(geoIPSvc, weatherSvc, updSvc, logV2)
	mux.Handle("/v2/time", wrapMiddlewares(hdlV2.HandleTime, geoIPSvc, logV2))
	mux.Handle("/v2/weather", wrapMiddlewares(hdlV2.HandleWeather, geoIPSvc, logV2))
	mux.Handle("/v2/firmware/update", wrapMiddlewares(hdlV2.HandleUpdate, geoIPSvc, logV2))

	l404 := l.With().Str("pkg", "not_found_handler").Logger()
	h404 := handlerNotFound.New(l404)
	mux.HandleFunc("/", wrapMiddlewares(h404.Handle, geoIPSvc, l404))
}

func wrapMiddlewares(h http.HandlerFunc, geoIPSvc *geoip.Service, l zerolog.Logger) http.HandlerFunc {
	h = httputil.ClientInfoMiddleware(h) // called last
	h = httputil.LogRequestMiddleware(h, l)
	h = auth.WrapHTTP(h)
	h = geoip.WrapHTTP(h, geoIPSvc, l)
	h = remoteaddr.WrapHTTP(h) // called first

	return h
}
