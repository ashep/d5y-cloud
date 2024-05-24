package time

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/geoip"
	"github.com/ashep/d5y/internal/remoteaddr"
	"github.com/ashep/d5y/internal/tz"
)

type ResponseTimestamp struct {
	TZ     string `json:"tz"`
	TZData string `json:"tz_data"`
	Value  int64  `json:"value"`
}

type ResponseGeo struct {
	Country  string `json:"country"`
	Region   string `json:"region"`
	City     string `json:"city"`
	Timezone string `json:"timezone"`
}

type Response struct {
	Timestamp *ResponseTimestamp `json:"timestamp,omitempty"`
}

type Handler struct {
	geoIP *geoip.Service
	l     zerolog.Logger
}

func New(geoIPSvc *geoip.Service, l zerolog.Logger) *Handler {
	return &Handler{
		geoIP: geoIPSvc,
		l:     l,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	res := &Response{
		Timestamp: &ResponseTimestamp{
			Value: time.Now().Unix(),
		},
	}

	rAddr := remoteaddr.FromCtx(r.Context())
	if rAddr == "" {
		h.l.Error().Msg("no remote addr in request context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	geo := geoip.FromCtx(r.Context())
	if geo == nil {
		h.l.Error().Msg("no geo ip data in request context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Timestamp.TZ = geo.Timezone
	res.Timestamp.TZData = tz.ToPosix(geo.Timezone)

	b, err := json.Marshal(res)
	if err != nil {
		h.l.Error().Err(err).Msg("response marshal error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(b); err != nil {
		h.l.Error().Err(err).Msg("response write error")
		return
	}

	h.l.Info().RawJSON("data", b).Msg("response")
}
