package v3

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/ashep/d5y/internal/api/rpcutil"
	"github.com/ashep/d5y/internal/fwupdate"
	"github.com/ashep/go-app/prommetrics"
)

func (h *Handler) HandleFirmwareUpdate(rw http.ResponseWriter, req *http.Request) { //nolint:cyclop // later
	l := rpcutil.ReqLog(req, h.l)
	m := prommetrics.MeasureHTTPServerRequest(req, req.URL.Path)

	if req.Method != http.MethodGet {
		m(http.StatusMethodNotAllowed)
		l.Warn().Err(errors.New("method not allowed")).Msg("firmware update request failed")
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	q := req.URL.Query()
	l.Info().Str("query", q.Encode()).Msg("firmware update request")

	appQ := q.Get("app")
	if appQ == "" {
		m(http.StatusBadRequest)
		l.Warn().Err(errors.New("invalid app")).Msg("firmware update request failed")
		rpcutil.WriteBadRequest(rw, "invalid app", l)
		return
	}

	appS := strings.Split(appQ, ":")
	if len(appS) != 4 {
		m(http.StatusBadRequest)
		l.Warn().Err(errors.New("invalid app")).Msg("firmware update request failed")
		rpcutil.WriteBadRequest(rw, "invalid app", l)
		return
	}

	ver, err := semver.NewVersion(appS[3])
	if err != nil {
		m(http.StatusBadRequest)
		l.Warn().Err(errors.New("invalid app version")).Msg("firmware update request failed")
		rpcutil.WriteBadRequest(rw, "invalid version", l)
		return
	}

	rlsSet, err := h.update.List(req.Context(), appS[0], appS[1], appS[2], q.Get("to_alpha") == "1")
	if errors.Is(err, fwupdate.ErrAppNotFound) {
		m(http.StatusNotFound)
		l.Warn().Err(errors.New("unknown client app")).Msg("firmware update request failed")
		rpcutil.WriteNotFound(rw, err.Error(), l)
		return
	} else if err != nil {
		m(http.StatusInternalServerError)
		rpcutil.WriteInternalServerError(rw, err, l)
		return
	}

	rls := rlsSet.Next(ver)
	if rls == nil {
		m(http.StatusOK) // OK is the correct code here
		l.Info().Str("result", "no next release").Msg("firmware update response")
		rpcutil.WriteNotFound(rw, "no firmware update found", l)
		return
	}

	if len(rls.Assets) == 0 {
		m(http.StatusOK) // OK is the correct code here
		l.Warn().
			Str("release", rls.Version.String()).
			Str("result", "no assets in the release").
			Msg("firmware update response")
		rpcutil.WriteNotFound(rw, "no firmware update found", l)
		return
	}

	b, err := json.Marshal(rls.Assets[0])
	if err != nil {
		m(http.StatusInternalServerError)
		l.Error().Err(fmt.Errorf("marshal response: %w", err)).Msg("firmware update request failed")
		rpcutil.WriteInternalServerError(rw, err, l)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	if _, err := rw.Write(b); err != nil {
		m(http.StatusInternalServerError)
		l.Error().Err(fmt.Errorf("write response: %w", err)).Msg("firmware update request failed")
		return
	}

	m(http.StatusOK)

	l.Info().
		Str("download_url", rls.Assets[0].URL).
		Str("version", rls.Version.String()).
		Msg("firmware update response")
}
