package fwupdate

import (
	"context"
	"errors"
	"fmt"
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
		return
	} else if err != nil {
		handlerutil.WriteInternalServerError(w, err, h.l)
		return
	}

	rls := rlsSet.Next(ver)
	if rls == nil || len(rls.Assets) == 0 {
		handlerutil.WriteNotFound(w, "no update found", h.l)
		return
	}

	finURL, err := getFinalURL(r.Context(), rls.Assets[0].URL)
	if err != nil {
		handlerutil.WriteInternalServerError(w, fmt.Errorf("get final url: %w", err), h.l)
		return
	}

	w.Header().Set("Location", finURL)
	w.WriteHeader(http.StatusFound)

	h.l.Info().Str("location", finURL).Str("version", rls.Version.String()).Msg("response")
}

// getFinalURL follows all redirects to find the effective URL.
func getFinalURL(ctx context.Context, s string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s, nil)
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http get: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http get: bad response status code: %d", res.StatusCode)
	}

	return res.Request.URL.String(), nil
}
