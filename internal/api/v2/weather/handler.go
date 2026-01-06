package weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/ashep/d5y/internal/alertsinua"
	"github.com/ashep/d5y/internal/api/rpcutil"
	"github.com/ashep/go-app/prommetrics"
	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/weatherapi"
)

type AirRaidAlert struct {
	Active bool `json:"active,omitempty"`
}

type Response struct {
	Location weatherapi.Location      `json:"location"`
	Current  weatherapi.ConditionItem `json:"current"`
	AirRaid  AirRaidAlert             `json:"air_raid,omitzero"`
}

type Handler struct {
	weather   *weatherapi.Client
	airAlerts *alertsinua.Client
	l         zerolog.Logger
}

func New(weatherCli *weatherapi.Client, airAlertsCli *alertsinua.Client, l zerolog.Logger) *Handler {
	return &Handler{
		weather:   weatherCli,
		airAlerts: airAlertsCli,
		l:         l,
	}
}

func (h *Handler) Handle(rw http.ResponseWriter, req *http.Request) {
	l := rpcutil.ReqLog(req, h.l)
	l.Info().Msg("weather request")

	m := prommetrics.MeasureHTTPServerRequest(req, "/v2/weather")

	wData, err := h.weather.GetForRequest(req)
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

	rsp := &Response{
		Location: wData.Location,
		Current:  wData.Current,
		AirRaid: AirRaidAlert{
			Active: h.airAlerts.IsActiveInRegion(wData.Location.Region, wData.Location.Name),
		},
	}

	b, err := json.Marshal(rsp)
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
