package clientinfo

import (
	"context"
	"net/http"

	"github.com/ashep/go-app/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

func WrapHTTP(next http.HandlerFunc, l zerolog.Logger) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		ci := FromRequest(req, l)

		labels := prometheus.Labels{
			"id":       ci.ID,
			"vendor":   ci.Vendor,
			"name":     ci.Name,
			"hardware": ci.Hardware,
			"version":  ci.Version,
			"country":  ci.Country,
			"city":     ci.City,
		}

		metrics.Counter("client_info", "D5Y Cloud clients.", labels).With(labels).Inc()

		next.ServeHTTP(rw, req.Clone(context.WithValue(req.Context(), ctxKey, ci)))
	}
}
