package rpcutil

import (
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/clientinfo"
)

func WriteBadRequest(w http.ResponseWriter, msg string, l zerolog.Logger) {
	w.WriteHeader(http.StatusBadRequest)

	if _, err := w.Write([]byte(msg)); err != nil {
		l.Error().Err(err).Msg("failed to write response")
	}
}

func WriteNotFound(w http.ResponseWriter, msg string, l zerolog.Logger) {
	w.WriteHeader(http.StatusNotFound)

	if _, err := w.Write([]byte(msg)); err != nil {
		l.Error().Err(err).Msg("failed to write response")
	}
}

func WriteInternalServerError(w http.ResponseWriter, err error, l zerolog.Logger) {
	w.WriteHeader(http.StatusInternalServerError)

	c := strconv.Itoa(int(time.Now().UnixMilli()))
	l.Error().Err(err).Str("code", c).Msg("internal server error")

	if _, wErr := w.Write([]byte("error #" + c)); wErr != nil {
		l.Error().Err(wErr).Msg("failed to write response")
	}
}

func ReqLog(req *http.Request, l zerolog.Logger) zerolog.Logger {
	ci := clientinfo.FromCtx(req.Context())

	ll := l.With().
		Str("req_method", req.Method).
		Str("req_uri", req.RequestURI).
		Str("user_agent", ci.UserAgent).
		Str("remote_addr", ci.RemoteAddr).
		Str("client_id", ci.ID).
		Str("client_vendor", ci.Vendor).
		Str("client_name", ci.Name).
		Str("client_version", ci.Version).
		Str("client_hardware", ci.Hardware).
		Str("client_country", ci.Country).
		Str("client_city", ci.City).
		Str("client_timezone", ci.Timezone)

	return ll.Logger()
}
