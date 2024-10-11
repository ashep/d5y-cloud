package update

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

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

	appQ := q.Get("app")
	if appQ == "" {
		handlerutil.WriteBadRequest(w, "invalid app", h.l)
		return
	}

	appS := strings.Split(appQ, ":")
	if len(appS) != 4 {
		handlerutil.WriteBadRequest(w, "invalid app", h.l)
		return
	}

	ver, err := semver.NewVersion(appS[3])
	if err != nil {
		h.l.Warn().Err(err).Str("version", appS[3]).Msg("parse app version failed")
		handlerutil.WriteBadRequest(w, "invalid version", h.l)
		return
	}

	toAlpha := q.Get("to_alpha")

	l := handlerutil.ReqLogger(r, h.l).With().
		Str("client_app_name", appS[0]+"/"+appS[1]).
		Str("client_app", appS[2]).
		Str("client_app_v", appS[3]).
		Str("to_alpha", toAlpha).
		Logger()

	l.Info().Msg("firmware update requested")

	rlsSet, err := h.updSvc.List(r.Context(), appS[0], appS[1], appS[2], toAlpha != "0")
	if errors.Is(err, update.ErrAppNotFound) {
		l.Warn().Str("reason", "unknown app").Msg("no updates found")
		handlerutil.WriteNotFound(w, err.Error(), l)
		return
	} else if err != nil {
		handlerutil.WriteInternalServerError(w, err, l)
		return
	}

	rls := rlsSet.Next(ver)
	if rls == nil {
		l.Warn().Str("reason", "no release").Msg("no update found")
		handlerutil.WriteNotFound(w, "no update found", l)
		return
	}

	if len(rls.Assets) == 0 {
		l.Warn().Str("reason", "no assets").Msg("no updates found")
		handlerutil.WriteNotFound(w, "no updates found", l)
		return
	}

	b, err := json.Marshal(rls.Assets[0])
	if err != nil {
		handlerutil.WriteInternalServerError(w, err, l)
		return
	}

	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(b); err != nil {
		l.Error().Err(err).Msg("failed to write response")
		return
	}

	l.Info().
		Str("url", rls.Assets[0].URL).
		Str("version", rls.Version.String()).
		Msg("update found")
}
