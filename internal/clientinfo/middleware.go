package clientinfo

import (
	"context"
	"net/http"

	"github.com/ashep/go-app/prommetrics"
	"github.com/bigdatacloudapi/bigdatacloud-go/bigdatacloud"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

func WrapHTTP(next http.HandlerFunc, cache *Cache, bdc *bigdatacloud.Client, l zerolog.Logger) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		ci := FromRequest(req, cache, bdc, l)
		labels := prometheus.Labels{
			"id":       ci.ID,
			"vendor":   ci.Vendor,
			"name":     ci.Name,
			"hardware": ci.Hardware,
			"version":  ci.Version,
			"country":  ci.Country,
			"city":     ci.City,
		}

		prommetrics.GetCounter("d5y_cloud_client_info", "D5Y Cloud client info", labels).With(labels).Inc()

		next.ServeHTTP(rw, req.Clone(context.WithValue(req.Context(), ctxKey, ci)))
	}
}
