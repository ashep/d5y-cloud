package httputil

import (
	"net/http"
	"strings"

	"github.com/ashep/go-app/metrics"
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
			"country":         "",
			"city":            "",
			"client_id":       "",
			"client_vendor":   "",
			"client_name":     "",
			"client_hardware": "",
			"client_version":  "",
		}

		if ua := req.Header.Get("User-Agent"); ua != "" {
			if uaSplit := strings.SplitN(ua, ":", 4); len(uaSplit) == 4 {
				labels["client_vendor"] = uaSplit[0]
				labels["client_name"] = uaSplit[1]
				labels["client_hardware"] = uaSplit[2]
				labels["client_version"] = uaSplit[3]
			}
		}

		if geoIP := geoip.FromCtx(req.Context()); geoIP != nil {
			labels["country"] = geoIP.CountryName
			labels["city"] = geoIP.City
		}

		if clientID := auth.TokenFromCtx(req.Context()); clientID != "" {
			labels["client_id"] = clientID
		}

		metrics.Counter("app_client_info", "D5Y Cloud clients.", labels).With(labels).Inc()

		next.ServeHTTP(rw, req)
	}
}
