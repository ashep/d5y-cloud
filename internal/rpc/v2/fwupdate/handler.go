package fwupdate

import (
	"errors"
	"net/http"

	"github.com/Masterminds/semver/v3"
	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/rpc/handlerutil"
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

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) { //nolint:cyclop // later
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()

	app := q.Get("app")
	if app == "" {
		handlerutil.WriteBadRequest(w, "invalid app", h.l)
		return
	}

	arch := q.Get("arch")
	if arch == "" {
		handlerutil.WriteBadRequest(w, "invalid arch", h.l)
		return
	}

	hw := q.Get("hw")
	if hw == "" {
		handlerutil.WriteBadRequest(w, "invalid hw", h.l)
		return
	}

	qVer := q.Get("ver")
	if qVer == "" {
		qVer = "0.0.0"
	}

	ver, err := semver.NewVersion(qVer)
	if err != nil {
		handlerutil.WriteBadRequest(w, "invalid version", h.l)
		return
	}

	rlsSet, err := h.updSvc.List(r.Context(), app, arch, hw)
	if errors.Is(err, update.ErrAppNotFound) {
		handlerutil.WriteNotFound(w, err.Error(), h.l)
		h.l.Info().Str("result", "app not found").Msg("response")

		return
	} else if err != nil {
		handlerutil.WriteInternalServerError(w, err, h.l)
		return
	}

	rls := rlsSet.Next(ver)
	if rls == nil || len(rls.Assets) == 0 {
		handlerutil.WriteNotFound(w, "no update found", h.l)
		h.l.Info().Str("result", "no update found").Msg("response")

		return
	}

	w.Header().Set("Location", rls.Assets[0].URL)
	w.WriteHeader(http.StatusFound)

	if _, err := w.Write([]byte(rls.Assets[0].URL)); err != nil {
		h.l.Error().Err(err).Msg("failed to write response")
		return
	}

	h.l.Info().Str("url", rls.Assets[0].URL).Str("version", rls.Version.String()).Msg("response")
}
