package handlerutil

import (
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/auth"
	"github.com/ashep/d5y/internal/geoip"
	"github.com/ashep/d5y/internal/remoteaddr"
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

func ReqLogger(r *http.Request, l zerolog.Logger) zerolog.Logger {
	ll := l.With().
		Str("method", r.Method).
		Str("uri", r.RequestURI).
		Str("user_agent", r.Header.Get("User-Agent"))

	if rAddr := remoteaddr.FromCtx(r.Context()); rAddr != "" {
		ll.Str("remote_addr", rAddr)
	}

	if geoIP := geoip.FromCtx(r.Context()); geoIP != nil {
		ll.Str("country", geoIP.CountryName).
			Str("region", geoIP.RegionName).
			Str("city", geoIP.City).
			Str("timezone", geoIP.Timezone)
	}

	if authTok := auth.TokenFromCtx(r.Context()); authTok != "" {
		ll.Str("client_id", authTok)
	}

	return ll.Logger()
}
