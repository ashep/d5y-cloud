package v2

import (
	"net/http"

	"github.com/rs/zerolog"

	timeh "github.com/ashep/d5y/internal/api/v2/time"
	weatherh "github.com/ashep/d5y/internal/api/v2/weather"
	"github.com/ashep/d5y/internal/geoip"
	"github.com/ashep/d5y/internal/weather"
)

type Handler struct {
	time    *timeh.Handler
	weather *weatherh.Handler
}

func New(geoIPCli *geoip.Service, weatherCli *weather.Client, l zerolog.Logger) *Handler {
	return &Handler{
		time:    timeh.New(geoIPCli, l.With().Str("handler", "time").Logger()),
		weather: weatherh.New(weatherCli, l.With().Str("handler", "weather").Logger()),
	}
}

func (h *Handler) HandleTime(w http.ResponseWriter, r *http.Request) {
	h.time.Handle(w, r)
}

func (h *Handler) HandleWeather(w http.ResponseWriter, r *http.Request) {
	h.weather.Handle(w, r)
}
