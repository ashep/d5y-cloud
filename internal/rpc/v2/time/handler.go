package time

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/geoip"
	"github.com/ashep/d5y/internal/remoteaddr"
	"github.com/ashep/d5y/internal/rpc/handlerutil"
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

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	l := handlerutil.ReqLogger(r, h.l)

	res := &Response{
		Value: time.Now().Unix(),
	}

	rAddr := remoteaddr.FromCtx(r.Context())
	if rAddr == "" {
		l.Error().Msg("no remote addr in the time request context")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	geo := geoip.FromCtx(r.Context())
	if geo == nil {
		l.Error().Msg("no geo ip data in the time request context")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	res.TZ = geo.Timezone
	res.TZData = tz.ToPosix(geo.Timezone)

	b, err := json.Marshal(res)
	if err != nil {
		l.Error().Err(err).Msg("time response marshal error")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err = w.Write(b); err != nil {
		l.Error().Err(err).Msg("time response write error")
		return
	}

	l.Info().RawJSON("data", b).Msg("time response sent")
}
