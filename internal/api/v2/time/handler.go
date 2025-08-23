package time

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ashep/d5y/internal/api/rpcutil"
	"github.com/ashep/d5y/internal/weatherapi"
	"github.com/ashep/go-app/metrics"
	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/clientinfo"
	"github.com/ashep/d5y/internal/tz"
)

type Response struct {
	TZ     string `json:"tz"`
	TZData string `json:"tz_data"`
	Value  int64  `json:"value"`
}

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
	l.Info().Msg("time request")

	m := metrics.HTTPServerRequest(req, "/v2/time")

	res := &Response{
		Value: time.Now().Unix(),
	}

	// Try Weather API, then GeoIP from client info
	if wAPIData, err := h.wAPI.GetFromRequest(req); err == nil && wAPIData.Location.Timezone != "" {
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
