package v1

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ashep/go-app/metrics"
	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/geoip"
	"github.com/ashep/d5y/internal/remoteaddr"
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
	geoIP   *geoip.Service
	weather *weather.Service
	l       zerolog.Logger
}

func New(g *geoip.Service, w *weather.Service, l zerolog.Logger) *Handler {
	return &Handler{
		geoIP:   g,
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

	m := metrics.HTTPServerRequest(req, "/v1")

	rAddr := remoteaddr.FromRequest(req)
	if rAddr == "" {
		m(http.StatusInternalServerError)
		h.l.Error().Msg("empty remote address")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	geo := geoip.FromCtx(req.Context())
	if geo == nil {
		m(http.StatusInternalServerError)
		h.l.Warn().Msg("empty geo ip data")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	tz, err := time.LoadLocation(geo.Timezone)
	if err != nil {
		h.l.Warn().Err(err).Msg("time zone detect failed")
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
	weatherData, err := h.weather.GetForIPAddr(rAddr)
	if err == nil {
		resp.Weather = true
		resp.Temp = weatherData.Current.Temp
		resp.FeelsLike = weatherData.Current.FeelsLike
	} else {
		h.l.Error().Err(err).Msg("weather get failed")
	}

	d, err := json.Marshal(resp)
	if err != nil {
		m(http.StatusInternalServerError)
		h.l.Error().Err(err).Msg("response marshal failed")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	if _, err = rw.Write(d); err != nil {
		m(http.StatusInternalServerError)
		h.l.Error().Err(err).Msg("response write failed")
	}

	m(http.StatusOK)
	h.l.Info().RawJSON("data", d).Msg("response")
}
