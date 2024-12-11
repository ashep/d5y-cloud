package server

import (
	"context"
	"net/http"
	"time"

	"github.com/ashep/d5y/internal/httputil"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

func New(
	addr string,
	giSvc *geoip.Service,
	wthSvc *weather.Service,
	updSvc *update.Service,
	l zerolog.Logger,
) *Server {
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())

	lv1 := l.With().Str("pkg", "v1_handler").Logger()
	hv1 := handlerV1.New(giSvc, wthSvc, l)
	mux.HandleFunc("/api/1", wrapMiddlewares(hv1.Handle, giSvc, lv1)) // BC

	lv2 := l.With().Str("pkg", "v2_handler").Logger()
	hv2 := handlerV2.New(giSvc, wthSvc, updSvc, lv2)
	mux.Handle("/v2/time", wrapMiddlewares(hv2.HandleTime, giSvc, lv2))
	mux.Handle("/v2/weather", wrapMiddlewares(hv2.HandleWeather, giSvc, lv2))
	mux.Handle("/v2/firmware/update", wrapMiddlewares(hv2.HandleUpdate, giSvc, lv2))

	l404 := l.With().Str("pkg", "not_found_handler").Logger()
	h404 := handlerNotFound.New(l404)
	mux.HandleFunc("/", wrapMiddlewares(h404.Handle, giSvc, l404))

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
	h = httputil.ClientInfoMiddleware(h) // called last
	h = httputil.LogRequestMiddleware(h, l)
	h = auth.WrapHTTP(h)
	h = geoip.WrapHTTP(h, geoIPSvc, l)
	h = remoteaddr.WrapHTTP(h) // called first

	return h
}
