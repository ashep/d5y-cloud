package time

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ashep/d5y/pkg/pmetrics"
	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/geoip"
	"github.com/ashep/d5y/internal/remoteaddr"
	"github.com/ashep/d5y/internal/tz"
)

type Response struct {
	TZ     string `json:"tz"`
	TZData string `json:"tz_data"`
	Value  int64  `json:"value"`
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

func (h *Handler) Handle(rw http.ResponseWriter, req *http.Request) {
	m := pmetrics.HTTPServerRequest(req)

	res := &Response{
		Value: time.Now().Unix(),
	}

	rAddr := remoteaddr.FromCtx(req.Context())
	if rAddr == "" {
		m(http.StatusInternalServerError)
		h.l.Error().Msg("no remote addr in request context")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	geo := geoip.FromCtx(req.Context())
	if geo == nil {
		m(http.StatusInternalServerError)
		h.l.Error().Msg("no geo ip data in request context")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.TZ = geo.Timezone
	res.TZData = tz.ToPosix(geo.Timezone)

	b, err := json.Marshal(res)
	if err != nil {
		m(http.StatusInternalServerError)
		h.l.Error().Err(err).Msg("response marshal error")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	if _, err = rw.Write(b); err != nil {
		m(http.StatusInternalServerError)
		h.l.Error().Err(err).Msg("response write error")
		return
	}

	m(http.StatusOK)
	h.l.Info().RawJSON("data", b).Msg("response")
}
