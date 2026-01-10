package app

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ashep/d5y/internal/alertsinua"
	handlerNotFound "github.com/ashep/d5y/internal/api/notfound"
	handlerV1 "github.com/ashep/d5y/internal/api/v1"
	handlerV2 "github.com/ashep/d5y/internal/api/v2"
	handlerV3 "github.com/ashep/d5y/internal/api/v3"
	"github.com/ashep/d5y/internal/clientinfo"
	"github.com/ashep/d5y/internal/fwupdate"
	"github.com/ashep/d5y/internal/weather"
	"github.com/ashep/go-app/health"
	"github.com/ashep/go-app/httpserver"
	"github.com/ashep/go-app/prommetrics"
	"github.com/ashep/go-app/runner"
	"github.com/google/go-github/v63/github"
	"github.com/rs/zerolog"
)

func Run(rt *runner.Runtime[Config]) error {
	ctx := rt.Ctx
	cfg := rt.Cfg
	l := rt.Log

	srv := httpserver.New(httpserver.WithAddr(":9000"))
	health.RegisterServer(srv)
	prommetrics.RegisterServer(rt.AppName, rt.AppVersion, srv)

	airRaidCli := alertsinua.New(cfg.AlertsInUA.APIKey, l.With().Str("pkg", "air_raid_client").Logger())
	if err := airRaidCli.Start(ctx); err != nil {
		l.Error().Err(err).Msg("air raid alerts client start failed")
	}

	weatherCli, err := weather.New(cfg.Weather.APIKey, l.With().Str("pkg", "weather_client").Logger())
	if err != nil {
		return fmt.Errorf("create weather client: %w", err)
	}

	githubCli := github.NewClient(http.DefaultClient).WithAuthToken(cfg.GitHub.Token)
	fwUpdateCli := fwupdate.New(githubCli, l.With().Str("pkg", "fw_update").Logger())

	logV1 := l.With().Str("pkg", "v1_handler").Logger()
	hdlV1 := handlerV1.New(weatherCli, logV1)
	srv.HandleFunc("/api/1", wrapMiddlewares(hdlV1.Handle, logV1)) // BC

	logV2 := l.With().Str("pkg", "v2_handler").Logger()
	hdlV2 := handlerV2.New(weatherCli, fwUpdateCli, logV2)
	srv.Handle("/v2/time", wrapMiddlewares(hdlV2.HandleTime, logV2))
	srv.Handle("/v2/weather", wrapMiddlewares(hdlV2.HandleWeather, logV2))
	srv.Handle("/v2/firmware/update", wrapMiddlewares(hdlV2.HandleUpdate, logV2))

	logV3 := l.With().Str("pkg", "v3_handler").Logger()
	hdlV3 := handlerV3.New(weatherCli, fwUpdateCli, airRaidCli, logV3)
	srv.Handle("/v3/time", wrapMiddlewares(hdlV3.HandleCurrentTime, logV3))
	srv.Handle("/v3/weather/current", wrapMiddlewares(hdlV3.HandleCurrentWeather, logV3))
	srv.Handle("/v3/firmware/update", wrapMiddlewares(hdlV3.HandleFirmwareUpdate, logV3))
	srv.Handle("/v3/air/raid", wrapMiddlewares(hdlV3.HandleAirRaid, logV3))

	log404 := l.With().Str("pkg", "404_handler").Logger()
	hdl404 := handlerNotFound.New(log404)
	srv.HandleFunc("/", wrapMiddlewares(hdl404.Handle, log404))

	var resErr error
	l.Info().Str("addr", srv.Listener().Addr().String()).Msg("starting")
	if srvErr := srv.Run(ctx); srvErr != nil {
		if !errors.Is(srvErr, http.ErrServerClosed) {
			resErr = errors.Join(resErr, fmt.Errorf("server: %w", srvErr))
		}
	}

	return resErr
}

func wrapMiddlewares(h http.HandlerFunc, l zerolog.Logger) http.HandlerFunc {
	h = clientinfo.WrapHTTP(h, l)
	return h
}
