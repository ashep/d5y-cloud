package handlerutil

import (
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog"
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
