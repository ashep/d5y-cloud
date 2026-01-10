package v3

import (
	"github.com/ashep/d5y/internal/alertsinua"
	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/fwupdate"
	"github.com/ashep/d5y/internal/weather"
)

type Handler struct {
	weather *weather.Client
	update  *fwupdate.Service
	airRaid *alertsinua.Client
	l       zerolog.Logger
}

func New(weather *weather.Client, updSvc *fwupdate.Service, airRaid *alertsinua.Client, l zerolog.Logger) *Handler {
	return &Handler{
		weather: weather,
		update:  updSvc,
		airRaid: airRaid,
		l:       l,
	}
}
