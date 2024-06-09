package httplog

import (
	"net/http"

	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/auth"
	"github.com/ashep/d5y/internal/geoip"
	"github.com/ashep/d5y/internal/remoteaddr"
)

func LogRequest(next http.HandlerFunc, l zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := l.Info().
			Str("method", r.Method).
			Str("uri", r.RequestURI).
			Str("user_agent", r.Header.Get("User-Agent"))

		if rAddr := remoteaddr.FromCtx(r.Context()); rAddr != "" {
			log.Str("remote_addr", rAddr)
		}

		if geoIP := geoip.FromCtx(r.Context()); geoIP != nil {
			log.Str("country", geoIP.CountryName).
				Str("region", geoIP.RegionName).
				Str("city", geoIP.City).
				Str("timezone", geoIP.Timezone)
		}

		if authTok := auth.TokenFromCtx(r.Context()); authTok != "" {
			log.Str("client_id", authTok)
		}

		log.Msg("request")

		next.ServeHTTP(w, r)
	}
}
