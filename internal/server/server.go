package server

import (
	"context"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/auth"
	"github.com/ashep/d5y/internal/geoip"
	"github.com/ashep/d5y/internal/httplog"
	"github.com/ashep/d5y/internal/remoteaddr"
	"github.com/ashep/d5y/internal/weather"

	handlerV1 "github.com/ashep/d5y/internal/api/v1"
	handlerV2 "github.com/ashep/d5y/internal/api/v2"
)

type Server struct {
	s *http.Server
	l zerolog.Logger
}

func New(addr, weatherAPIKey string, l zerolog.Logger) *Server {
	mux := http.NewServeMux()

	geoIPSvc := geoip.New()
	weatherSvc := weather.New(weatherAPIKey)

	hv1 := handlerV1.New(geoIPSvc, weatherSvc, l)
	mux.HandleFunc("/", wrapMiddlewares(hv1.Handle, geoIPSvc, l.With().Str("pkg", "v1_middleware").Logger())) // BC

	hv2 := handlerV2.New(geoIPSvc, weatherSvc, l)
	mux.Handle("/v2/time", wrapMiddlewares(hv2.HandleTime, geoIPSvc, l.With().Str("pkg", "time_middleware").Logger()))
	mux.Handle("/v2/weather", wrapMiddlewares(hv2.HandleWeather, geoIPSvc, l.With().Str("pkg", "weather_middleware").Logger()))

	return &Server{
		s: &http.Server{Addr: addr, Handler: mux},
		l: l,
	}
}

func (s *Server) Run() error {
	s.l.Info().Str("addr", s.s.Addr).Msg("server starting")
	return s.s.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) {
	s.l.Info().Msg("shutting down the server")

	err := s.s.Shutdown(ctx)
	if err != nil {
		s.l.Error().Err(err).Msg("server shutdown failed")
	}

	s.l.Info().Msg("server stopped")
}

func wrapMiddlewares(h http.HandlerFunc, geoIPSvc *geoip.Service, l zerolog.Logger) http.HandlerFunc {
	h = httplog.LogRequest(h, l) // called last
	h = auth.WrapHTTP(h)
	h = geoip.WrapHTTP(h, geoIPSvc, l)
	h = remoteaddr.WrapHTTP(h) // called first

	return h
}
