package v3

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ashep/d5y/internal/api/rpcutil"
	"github.com/ashep/go-app/prommetrics"
)

type AirRaidAlertResponse struct {
	Active bool `json:"active"`
}

func (h *Handler) HandleAirRaid(rw http.ResponseWriter, req *http.Request) {
	l := rpcutil.ReqLog(req, h.l)
	l.Info().Msg("air raid alert request")

	m := prommetrics.MeasureHTTPServerRequest(req, req.URL.Path)

	loc, err := h.weather.LocationForRequest(req)
	if err != nil {
		m(http.StatusInternalServerError)
		l.Error().Err(fmt.Errorf("call weather api: %w", err)).Msg("air raid alert request failed")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	res := AirRaidAlertResponse{
		Active: h.airRaid.IsActiveInRegion(loc.Region, loc.Name),
	}

	b, err := json.Marshal(res)
	if err != nil {
		m(http.StatusInternalServerError)
		l.Error().Err(fmt.Errorf("marshal response: %w", err)).Msg("air raid alert request failed")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	if _, err = rw.Write(b); err != nil {
		m(http.StatusInternalServerError)
		l.Error().Err(fmt.Errorf("write response: %w", err)).Msg("air raid alert request failed")
		return
	}

	m(http.StatusOK)
	l.Info().RawJSON("data", b).Msg("air raid alert response")
}
