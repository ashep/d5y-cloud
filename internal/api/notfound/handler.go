package notfound

import (
	"net/http"

	"github.com/ashep/d5y/internal/api/rpcutil"
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

func (h *Handler) Handle(rw http.ResponseWriter, req *http.Request) {
	l := rpcutil.ReqLog(req, h.l)
	l.Warn().Str("path", req.URL.Path).Msg("not found")
	rw.WriteHeader(http.StatusNotFound)
}
