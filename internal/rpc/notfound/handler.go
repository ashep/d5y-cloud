package notfound

import (
	"net/http"

	"github.com/rs/zerolog"
)

type Handler struct {
	l zerolog.Logger
}

func New(l zerolog.Logger) *Handler {
	return &Handler{
		l: l,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	h.l.Warn().Str("path", r.URL.Path).Msg("not found")
	w.WriteHeader(http.StatusNotFound)
}
