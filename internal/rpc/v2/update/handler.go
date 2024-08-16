package update

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

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()

	prod := q.Get("prod")
	if prod != "cronus" {
		handlerutil.WriteBadRequest(w, "invalid product", h.l)
		return
	}

	arch := q.Get("arch")
	if arch == "" {
		handlerutil.WriteBadRequest(w, "invalid architecture arch", h.l)
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

	rls, err := h.updSvc.List(r.Context(), prod, arch)
	if errors.Is(err, update.ErrProductNotFound) {
		handlerutil.WriteNotFound(w, err.Error(), h.l)
		return
	} else if err != nil {
		handlerutil.WriteInternalServerError(w, err, h.l)
		return
	}

	rl := rls.Next(ver)
	if rl == nil || len(rl.Assets) == 0 {
		handlerutil.WriteNotFound(w, "no update found", h.l)
		return
	}

	w.Header().Set("Location", rl.Assets[0].URL)
	w.WriteHeader(http.StatusFound)
}
