package server

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/auth"
	"github.com/ashep/d5y/internal/geoip"
	"github.com/ashep/d5y/internal/httplog"
	"github.com/ashep/d5y/internal/remoteaddr"
	"github.com/ashep/d5y/internal/weather"

	handlerNotFound "github.com/ashep/d5y/internal/api/notfound"
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

	lv1 := l.With().Str("pkg", "v1_handler").Logger()
	hv1 := handlerV1.New(geoIPSvc, weatherSvc, l)
	mux.HandleFunc("/api/1", wrapMiddlewares(hv1.Handle, geoIPSvc, lv1)) // BC

	lv2 := l.With().Str("pkg", "v2_handler").Logger()
	hv2 := handlerV2.New(geoIPSvc, weatherSvc, lv2)
	mux.Handle("/v2/time", wrapMiddlewares(hv2.HandleTime, geoIPSvc, lv2))
	mux.Handle("/v2/weather", wrapMiddlewares(hv2.HandleWeather, geoIPSvc, lv2))

	l404 := l.With().Str("pkg", "not_found_handler").Logger()
	h404 := handlerNotFound.New(l404)
	mux.HandleFunc("/", wrapMiddlewares(h404.Handle, geoIPSvc, l404))

	return &Server{
		s: &http.Server{
			Addr:        addr,
			Handler:     mux,
			ReadTimeout: time.Second * 5,
		},
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
