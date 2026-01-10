package v3

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ashep/d5y/internal/api/rpcutil"
	"github.com/ashep/d5y/internal/clientinfo"
	"github.com/ashep/d5y/internal/tz"
	"github.com/ashep/go-app/prommetrics"
)

type TimeResponse struct {
	TZ     string `json:"tz"`
	TZData string `json:"tz_data"`
	Value  int64  `json:"value"`
}

func (h *Handler) HandleCurrentTime(rw http.ResponseWriter, req *http.Request) {
	l := rpcutil.ReqLog(req, h.l)
	l.Info().Msg("time request")

	m := prommetrics.MeasureHTTPServerRequest(req, req.URL.Path)

	res := &TimeResponse{
		Value: time.Now().Unix(),
	}

	// Try Weather API, then GeoIP from client info
	if wAPIData, err := h.weather.WeatherForRequest(req); err == nil && wAPIData.Location.Timezone != "" {
		res.TZ = wAPIData.Location.Timezone
		res.TZData = tz.ToPosix(wAPIData.Location.Timezone)
	} else {
		ci := clientinfo.FromCtx(req.Context())
		if ci.RemoteAddr == "" {
			m(http.StatusInternalServerError)
			l.Error().Err(errors.New("missing remote address")).Msg("time request failed")
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		if ci.Timezone == "" {
			m(http.StatusInternalServerError)
			l.Error().Err(errors.New("missing timezone")).Msg("time request failed")
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.TZ = ci.Timezone
		res.TZData = tz.ToPosix(ci.Timezone)
	}

	b, err := json.Marshal(res)
	if err != nil {
		m(http.StatusInternalServerError)
		l.Error().Err(fmt.Errorf("marshal response: %w", err)).Msg("time request failed")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	if _, err = rw.Write(b); err != nil {
		m(http.StatusInternalServerError)
		l.Error().Err(fmt.Errorf("write response: %w", err)).Msg("time request failed")
		return
	}

	m(http.StatusOK)

	l.Info().RawJSON("data", b).Msg("time response")
}
