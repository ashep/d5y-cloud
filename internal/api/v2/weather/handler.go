package weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/ashep/d5y/internal/api/rpcutil"
	"github.com/ashep/go-app/metrics"
	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/clientinfo"
	"github.com/ashep/d5y/internal/weather"
)

type Response *weather.Data

type Handler struct {
	weather *weather.Service
	l       zerolog.Logger
}

func New(weatherCli *weather.Service, l zerolog.Logger) *Handler {
	return &Handler{
		weather: weatherCli,
		l:       l,
	}
}

func (h *Handler) Handle(rw http.ResponseWriter, req *http.Request) {
	l := rpcutil.ReqLog(req, h.l)
	l.Info().Msg("weather request")

	m := metrics.HTTPServerRequest(req, "/v2/weather")

	ci := clientinfo.FromCtx(req.Context())
	if ci.RemoteAddr == "" {
		m(http.StatusInternalServerError)
		l.Error().Err(errors.New("missing remote address")).Msg("weather request failed")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	d, err := h.weather.GetForIPAddr(ci.RemoteAddr)
	if err != nil {
		m(http.StatusInternalServerError)
		l.Error().Err(fmt.Errorf("get weather: %w", err)).Msg("weather request failed")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(d.Current)
	if err != nil {
		m(http.StatusInternalServerError)
		l.Error().Err(fmt.Errorf("marshal response: %w", err)).Msg("weather request failed")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	if _, err = rw.Write(b); err != nil {
		m(http.StatusInternalServerError)
		l.Error().Err(fmt.Errorf("write response: %w", err)).Msg("weather request failed")
		return
	}

	m(http.StatusOK)
	l.Info().RawJSON("data", b).Msg("weather response")
}
