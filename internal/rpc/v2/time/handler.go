package time

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ashep/d5y/internal/rpc/rpcutil"
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
	l zerolog.Logger
}

func New(l zerolog.Logger) *Handler {
	return &Handler{
		l: l,
	}
}

func (h *Handler) Handle(rw http.ResponseWriter, req *http.Request) {
	l := rpcutil.ReqLog(req, h.l)
	l.Info().Msg("time request")

	m := metrics.HTTPServerRequest(req, "/v2/time")

	res := &Response{
		Value: time.Now().Unix(),
	}

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

	b, err := json.Marshal(res)
	if err != nil {
		m(http.StatusInternalServerError)
		l.Error().Err(fmt.Errorf("load timezone: %w", err)).Msg("time request failed")
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
