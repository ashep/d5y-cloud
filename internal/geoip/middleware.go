package geoip

import (
	"context"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/ashep/d5y/internal/remoteaddr"
)

type ctxKeyType string

const ctxKey ctxKeyType = "geoIP"

func WrapHTTP(next http.HandlerFunc, svc *Service, l zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rAddr := remoteaddr.FromCtx(r.Context())
		if rAddr == "" {
			next.ServeHTTP(w, r)
			return
		}

		data, err := svc.Get(rAddr)
		if err != nil {
			l.Error().Err(err).Msg("get geoip data failed")
			next.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, r.Clone(context.WithValue(r.Context(), ctxKey, data)))
	}
}

func FromCtx(ctx context.Context) *Data {
	d, ok := ctx.Value(ctxKey).(*Data)
	if !ok {
		return nil
	}

	return d
}
