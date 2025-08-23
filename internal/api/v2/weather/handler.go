package weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/ashep/d5y/internal/api/rpcutil"
	"github.com/ashep/go-app/metrics"
	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/weatherapi"
)

type Handler struct {
	wAPI *weatherapi.Service
	l    zerolog.Logger
}

func New(wAPI *weatherapi.Service, l zerolog.Logger) *Handler {
	return &Handler{
		wAPI: wAPI,
		l:    l,
	}
}

func (h *Handler) Handle(rw http.ResponseWriter, req *http.Request) {
	l := rpcutil.ReqLog(req, h.l)
	l.Info().Msg("weather request")

	m := metrics.HTTPServerRequest(req, "/v2/weather")

	data, err := h.wAPI.GetFromRequest(req)
	if errors.Is(err, weatherapi.ErrInvalidArgument) {
		m(http.StatusBadRequest)
		l.Warn().Err(fmt.Errorf("call weather api: %w", err)).Msg("weather request failed")
		rw.WriteHeader(http.StatusBadRequest)
		return
	} else if err != nil {
		m(http.StatusInternalServerError)
		l.Error().Err(fmt.Errorf("call weather api: %w", err)).Msg("weather request failed")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(data.Current)
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
