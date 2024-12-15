package update

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/ashep/d5y/internal/httputil"
	"github.com/ashep/d5y/internal/update"
	"github.com/ashep/go-app/metrics"
	"github.com/rs/zerolog"
)

type Handler struct {
	updSvc *update.Service
	l      zerolog.Logger
}

func New(updSvc *update.Service, l zerolog.Logger) *Handler {
	return &Handler{
		updSvc: updSvc,
		l:      l,
	}
}

func (h *Handler) Handle(rw http.ResponseWriter, req *http.Request) { //nolint:cyclop // later
	m := metrics.HTTPServerRequest(req, "/v2/update")

	if req.Method != http.MethodGet {
		m(http.StatusMethodNotAllowed)
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	q := req.URL.Query()

	appQ := q.Get("app")
	if appQ == "" {
		m(http.StatusBadRequest)
		httputil.WriteBadRequest(rw, "invalid app", h.l)
		return
	}

	appS := strings.Split(appQ, ":")
	if len(appS) != 4 {
		m(http.StatusBadRequest)
		httputil.WriteBadRequest(rw, "invalid app", h.l)
		return
	}

	ver, err := semver.NewVersion(appS[3])
	if err != nil {
		m(http.StatusBadRequest)
		h.l.Warn().Err(err).Str("version", appS[3]).Msg("parse app version failed")
		httputil.WriteBadRequest(rw, "invalid version", h.l)
		return
	}

	toAlpha := q.Get("to_alpha")

	l := httputil.ReqLogger(req, h.l).With().
		Str("client_app_name", appS[0]+"/"+appS[1]).
		Str("client_app", appS[2]).
		Str("client_app_v", appS[3]).
		Str("to_alpha", toAlpha).
		Logger()

	l.Info().Msg("firmware update requested")

	rlsSet, err := h.updSvc.List(req.Context(), appS[0], appS[1], appS[2], toAlpha != "0")
	if errors.Is(err, update.ErrAppNotFound) {
		m(http.StatusNotFound)
		l.Warn().Str("reason", "unknown client app").Msg("no updates found")
		httputil.WriteNotFound(rw, err.Error(), l)
		return
	} else if err != nil {
		m(http.StatusInternalServerError)
		httputil.WriteInternalServerError(rw, err, l)
		return
	}

	rls := rlsSet.Next(ver)
	if rls == nil {
		m(http.StatusNotFound)
		l.Info().Str("reason", "no next release").Msg("no firmware update found")
		httputil.WriteNotFound(rw, "no update found", l)
		return
	}

	if len(rls.Assets) == 0 {
		l.Warn().
			Str("release", rls.Version.String()).
			Str("reason", "no assets in the release").
			Msg("no firmware update found")
		m(http.StatusNotFound)
		httputil.WriteNotFound(rw, "no firmware update found", l)
		return
	}

	b, err := json.Marshal(rls.Assets[0])
	if err != nil {
		m(http.StatusInternalServerError)
		httputil.WriteInternalServerError(rw, err, l)
		return
	}

	rw.WriteHeader(http.StatusOK)

	if _, err := rw.Write(b); err != nil {
		m(http.StatusInternalServerError)
		l.Error().Err(err).Msg("failed to write firmware update response")
		return
	}

	m(http.StatusOK)

	l.Info().
		Str("url", rls.Assets[0].URL).
		Str("version", rls.Version.String()).
		Msg("firmware response update sent")
}
