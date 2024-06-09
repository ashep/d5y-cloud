package remoteaddr

import (
	"context"
	"net/http"
)

type ctxKeyType string

const ctxKey ctxKeyType = "remoteAddress"

func WrapHTTP(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.Clone(context.WithValue(r.Context(), ctxKey, FromRequest(r))))
	}
}

func FromCtx(ctx context.Context) string {
	v, _ := ctx.Value(ctxKey).(string)
	return v
}
