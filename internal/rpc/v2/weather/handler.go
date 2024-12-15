package weather

import (
	"encoding/json"
	"net/http"

	"github.com/ashep/go-app/metrics"
	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/remoteaddr"
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
	m := metrics.HTTPServerRequest(req, "/v2/weather")

	rAddr := remoteaddr.FromCtx(req.Context())
	if rAddr == "" {
		m(http.StatusInternalServerError)
		h.l.Error().Msg("no remote addr in request context")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	d, err := h.weather.GetForIPAddr(rAddr)
	if err != nil {
		m(http.StatusInternalServerError)
		h.l.Error().Err(err).Msg("weather get failed")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(d.Current)
	if err != nil {
		m(http.StatusInternalServerError)
		h.l.Error().Err(err).Msg("response marshal error")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	if _, err = rw.Write(b); err != nil {
		m(http.StatusInternalServerError)
		h.l.Error().Err(err).Msg("response write error")
		return
	}

	m(http.StatusOK)
	h.l.Info().RawJSON("data", b).Msg("response")
}
