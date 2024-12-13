package server

import (
	"context"
	"net/http"

	"github.com/ashep/d5y/internal/httputil"
	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/auth"
	"github.com/ashep/d5y/internal/geoip"
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

func Setup(
	mux *http.ServeMux,
	geoipSvc *geoip.Service,
	wthSvc *weather.Service,
	updSvc *update.Service,
	l zerolog.Logger,
) {
	lv1 := l.With().Str("pkg", "v1_handler").Logger()
	hv1 := handlerV1.New(geoipSvc, wthSvc, l)
	mux.HandleFunc("/api/1", wrapMiddlewares(hv1.Handle, geoipSvc, lv1)) // BC

	logV2 := l.With().Str("pkg", "v2_handler").Logger()
	hdlV2 := handlerV2.New(geoipSvc, wthSvc, updSvc, logV2)
	mux.Handle("/v2/time", wrapMiddlewares(hdlV2.HandleTime, geoipSvc, logV2))
	mux.Handle("/v2/weather", wrapMiddlewares(hdlV2.HandleWeather, geoipSvc, logV2))
	mux.Handle("/v2/firmware/update", wrapMiddlewares(hdlV2.HandleUpdate, geoipSvc, logV2))

	l404 := l.With().Str("pkg", "not_found_handler").Logger()
	h404 := handlerNotFound.New(l404)
	mux.HandleFunc("/", wrapMiddlewares(h404.Handle, geoipSvc, l404))
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
	h = httputil.ClientInfoMiddleware(h) // called last
	h = httputil.LogRequestMiddleware(h, l)
	h = auth.WrapHTTP(h)
	h = geoip.WrapHTTP(h, geoIPSvc, l)
	h = remoteaddr.WrapHTTP(h) // called first

	return h
}
