package v2

import (
	"net/http"

	"github.com/rs/zerolog"

	timeh "github.com/ashep/d5y/internal/api/v2/time"
	updateh "github.com/ashep/d5y/internal/api/v2/update"
	weatherh "github.com/ashep/d5y/internal/api/v2/weather"
	"github.com/ashep/d5y/internal/update"
	"github.com/ashep/d5y/internal/weather"
)

type Handler struct {
	time    *timeh.Handler
	weather *weatherh.Handler
	update  *updateh.Handler
}

func New(wthSvc *weather.Service, updSvc *update.Service, l zerolog.Logger) *Handler {
	return &Handler{
		time:    timeh.New(l.With().Str("handler", "time").Logger()),
		weather: weatherh.New(wthSvc, l.With().Str("handler", "weather").Logger()),
		update:  updateh.New(updSvc, l),
	}
}

func (h *Handler) HandleTime(w http.ResponseWriter, r *http.Request) {
	h.time.Handle(w, r)
}

func (h *Handler) HandleWeather(w http.ResponseWriter, r *http.Request) {
	h.weather.Handle(w, r)
}

func (h *Handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	h.update.Handle(w, r)
}
