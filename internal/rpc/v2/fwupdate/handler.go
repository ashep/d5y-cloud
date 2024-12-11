package fwupdate

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/ashep/d5y/internal/httputil"
	"github.com/ashep/d5y/pkg/pmetrics"
	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/update"
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
	m := pmetrics.HTTPServerRequest(req)

	if req.Method != http.MethodGet {
		m(http.StatusMethodNotAllowed)
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	q := req.URL.Query()

	app := q.Get("app")
	if app == "" {
		m(http.StatusBadRequest)
		httputil.WriteBadRequest(rw, "invalid app", h.l)
		return
	}

	appS := strings.Split(app, "/")
	if len(appS) != 2 {
		m(http.StatusBadRequest)
		httputil.WriteBadRequest(rw, "invalid app", h.l)
		return
	}

	arch := q.Get("arch")
	if arch == "" {
		m(http.StatusBadRequest)
		httputil.WriteBadRequest(rw, "invalid arch", h.l)
		return
	}

	hw := q.Get("hw")
	if hw == "" {
		m(http.StatusBadRequest)
		httputil.WriteBadRequest(rw, "invalid hw", h.l)
		return
	}

	qVer := q.Get("ver")
	if qVer == "" {
		qVer = "0.0.0"
	}

	ver, err := semver.NewVersion(qVer)
	if err != nil {
		m(http.StatusBadRequest)
		httputil.WriteBadRequest(rw, "invalid version", h.l)
		return
	}

	rlsSet, err := h.updSvc.List(req.Context(), appS[0], appS[1], arch, hw)
	if errors.Is(err, update.ErrAppNotFound) {
		m(http.StatusNotFound)
		httputil.WriteNotFound(rw, err.Error(), h.l)
		h.l.Info().Str("result", "app not found").Msg("response")
		return
	} else if err != nil {
		m(http.StatusInternalServerError)
		httputil.WriteInternalServerError(rw, err, h.l)
		return
	}

	rls := rlsSet.Next(ver)
	if rls == nil || len(rls.Assets) == 0 {
		m(http.StatusNotFound)
		httputil.WriteNotFound(rw, "no update found", h.l)
		h.l.Info().Str("result", "no update found").Msg("response")
		return
	}

	b, err := json.Marshal(rls.Assets[0])
	if err != nil {
		m(http.StatusInternalServerError)
		httputil.WriteInternalServerError(rw, err, h.l)
		return
	}

	rw.WriteHeader(http.StatusOK)

	if _, err := rw.Write(b); err != nil {
		m(http.StatusInternalServerError)
		h.l.Error().Err(err).Msg("failed to write response")
		return
	}

	m(http.StatusOK)
	h.l.Info().Str("url", rls.Assets[0].URL).Str("version", rls.Version.String()).Msg("response")
}
