package weather

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/remoteaddr"
	"github.com/ashep/d5y/internal/rpc/handlerutil"
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

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	l := handlerutil.ReqLogger(r, h.l)

	rAddr := remoteaddr.FromCtx(r.Context())
	if rAddr == "" {
		l.Error().Msg("no remote addr in the weather request context")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	d, err := h.weather.GetForIPAddr(rAddr)
	if err != nil {
		l.Error().Err(err).Msg("weather get failed")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	b, err := json.Marshal(d.Current)
	if err != nil {
		l.Error().Err(err).Msg("weather response marshal error")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err = w.Write(b); err != nil {
		l.Error().Err(err).Msg("weather response write error")
		return
	}

	l.Info().RawJSON("data", b).Msg("weather response sent")
}
