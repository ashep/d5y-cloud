package httputil

import (
	"net/http"

	"github.com/ashep/d5y/pkg/pmetrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/auth"
	"github.com/ashep/d5y/internal/geoip"
	"github.com/ashep/d5y/internal/remoteaddr"
)

func LogRequestMiddleware(next http.HandlerFunc, l zerolog.Logger) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		log := l.Info().
			Str("method", req.Method).
			Str("uri", req.RequestURI).
			Str("user_agent", req.Header.Get("User-Agent"))

		if rAddr := remoteaddr.FromCtx(req.Context()); rAddr != "" {
			log.Str("remote_addr", rAddr)
		}

		if geoIP := geoip.FromCtx(req.Context()); geoIP != nil {
			log.Str("country", geoIP.CountryName).
				Str("region", geoIP.RegionName).
				Str("city", geoIP.City).
				Str("timezone", geoIP.Timezone)
		}

		if authTok := auth.TokenFromCtx(req.Context()); authTok != "" {
			log.Str("client_id", authTok)
		}

		log.Msg("request")

		next.ServeHTTP(rw, req)
	}
}

func ClientInfoMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		labels := prometheus.Labels{
			"country":   "unknown",
			"city":      "unknown",
			"client_id": "unknown",
		}

		if geoIP := geoip.FromCtx(req.Context()); geoIP != nil {
			labels["country"] = geoIP.CountryName
			labels["city"] = geoIP.City
		}

		if clientID := auth.TokenFromCtx(req.Context()); clientID != "" {
			labels["client_id"] = clientID
		}

		pmetrics.Counter("app_client_info", "D5Y Cloud clients.", labels).With(labels).Inc()

		next.ServeHTTP(rw, req)
	}
}
