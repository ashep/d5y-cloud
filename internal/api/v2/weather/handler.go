package weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ashep/d5y/internal/api/rpcutil"
	"github.com/ashep/go-app/metrics"
	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/clientinfo"
	"github.com/ashep/d5y/internal/weather"
)

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

	data, err := h.getForLocation(req.URL.Query().Get("lat"), req.URL.Query().Get("lng"))
	if err != nil {
		m(http.StatusInternalServerError)
		l.Error().Err(fmt.Errorf("get for location: %w", err)).Msg("weather request failed")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	if data == nil {
		ci := clientinfo.FromCtx(req.Context())
		if ci.RemoteAddr == "" {
			m(http.StatusInternalServerError)
			l.Error().Err(errors.New("missing remote ip address")).Msg("weather request failed")
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		if data, err = h.weather.GetForIPAddr(ci.RemoteAddr); err != nil {
			m(http.StatusInternalServerError)
			l.Error().Err(fmt.Errorf("get for remote ip address: %w", err)).Msg("weather request failed")
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
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

func (h *Handler) getForLocation(lat, lng string) (*weather.Data, error) {
	if lat == "" || lng == "" {
		return nil, nil
	}

	latF, err := strconv.ParseFloat(lat, 32)
	if err != nil {
		return nil, fmt.Errorf("parse lat: %w", err)
	}
	if latF == 0 {
		return nil, nil
	}

	lngF, err := strconv.ParseFloat(lng, 32)
	if err != nil {
		return nil, fmt.Errorf("parse lng: %w", err)
	}
	if lngF == 0 {
		return nil, nil
	}

	return h.weather.GetForLocation(latF, lngF)
}
