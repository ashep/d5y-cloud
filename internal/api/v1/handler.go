package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ashep/d5y/internal/api/rpcutil"
	"github.com/ashep/go-app/metrics"
	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/clientinfo"
	"github.com/ashep/d5y/internal/weather"
)

type Response struct {
	Second int `json:"second"`
	Minute int `json:"minute"`
	Hour   int `json:"hour"`
	Dow    int `json:"dow"`
	Day    int `json:"day"`
	Month  int `json:"month"`
	Year   int `json:"year"`

	Weather   bool    `json:"weather"`
	Temp      float64 `json:"temp"`
	FeelsLike float64 `json:"feels_like"`
}

type Handler struct {
	weather *weather.Service
	l       zerolog.Logger
}

func New(w *weather.Service, l zerolog.Logger) *Handler {
	return &Handler{
		weather: w,
		l:       l,
	}
}

func (h *Handler) Handle(rw http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" && req.URL.Path != "/api/1" {
		h.l.Warn().Str("path", req.URL.Path).Msg("unexpected path")
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	l := rpcutil.ReqLog(req, h.l)
	l.Info().Msg("v1 request")

	m := metrics.HTTPServerRequest(req, "/v1")

	ci := clientinfo.FromCtx(req.Context())
	if ci.RemoteAddr == "" {
		m(http.StatusInternalServerError)
		l.Error().Err(errors.New("missing remote address")).Msg("v1 request failed")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	if ci.Timezone == "" {
		m(http.StatusInternalServerError)
		l.Warn().Err(errors.New("missing timezone")).Msg("v1 request failed")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	tz, err := time.LoadLocation(ci.Timezone)
	if err != nil {
		m(http.StatusInternalServerError)
		l.Error().Err(fmt.Errorf("load timezone: %w", err)).Msg("v1 request failed")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	t := time.Now().In(tz)

	// Fix weekday number
	dow := int(t.Weekday())
	if dow == 0 {
		dow = 7
	}

	resp := Response{
		Second: t.Second(),
		Minute: t.Minute(),
		Hour:   t.Hour(),
		Dow:    dow,
		Day:    t.Day(),
		Month:  int(t.Month()),
		Year:   t.Year() - 2000,
	}

	// Add weather data
	weatherData, err := h.weather.GetForIPAddr(ci.RemoteAddr)
	if err == nil {
		resp.Weather = true
		resp.Temp = weatherData.Current.Temp
		resp.FeelsLike = weatherData.Current.FeelsLike
	} else {
		l.Error().Err(fmt.Errorf("get weather: %w", err)).Msg("v1 request warning")
	}

	d, err := json.Marshal(resp)
	if err != nil {
		m(http.StatusInternalServerError)
		l.Error().Err(fmt.Errorf("marshal response: %w", err)).Msg("v1 request failed")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	if _, err = rw.Write(d); err != nil {
		m(http.StatusInternalServerError)
		l.Error().Err(fmt.Errorf("write response: %w", err)).Msg("v1 request failed")
		return
	}

	m(http.StatusOK)

	l.Info().RawJSON("data", d).Msg("v1 response")
}
