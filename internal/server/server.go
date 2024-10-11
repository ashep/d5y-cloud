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
	handlerNotFound "github.com/ashep/d5y/internal/rpc/notfound"
	handlerV1 "github.com/ashep/d5y/internal/rpc/v1"
	handlerV2 "github.com/ashep/d5y/internal/rpc/v2"
	"github.com/ashep/d5y/internal/update"
	"github.com/ashep/d5y/internal/weather"
)

type Server struct {
	s *http.Server
	l zerolog.Logger
}

func New(
	addr string,
	geoipSvc *geoip.Service,
	wthSvc *weather.Service,
	updSvc *update.Service,
	l zerolog.Logger,
) *Server {
	mux := http.NewServeMux()

	logV1 := l.With().Str("pkg", "v1_handler").Logger()
	ndlV1 := handlerV1.New(geoipSvc, wthSvc, l)
	mux.HandleFunc("/api/1", wrap(httplog.LogRequest(ndlV1.Handle, logV1), geoipSvc, logV1)) // BC

	logV2 := l.With().Str("pkg", "v2_handler").Logger()
	hdlV2 := handlerV2.New(geoipSvc, wthSvc, updSvc, logV2)
	mux.Handle("/v2/time", wrap(httplog.LogRequest(hdlV2.HandleTime, logV2), geoipSvc, logV2))
	mux.Handle("/v2/weather", wrap(httplog.LogRequest(hdlV2.HandleWeather, logV2), geoipSvc, logV2))
	mux.Handle("/v2/firmware/update", wrap(hdlV2.HandleUpdate, geoipSvc, logV2))

	l404 := l.With().Str("pkg", "not_found_handler").Logger()
	h404 := handlerNotFound.New(l404)
	mux.HandleFunc("/", wrap(h404.Handle, geoipSvc, l404))

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

func wrap(h http.HandlerFunc, geoIPSvc *geoip.Service, l zerolog.Logger) http.HandlerFunc {
	h = auth.WrapHTTP(h)
	h = geoip.WrapHTTP(h, geoIPSvc, l)
	h = remoteaddr.WrapHTTP(h)

	return h
}
