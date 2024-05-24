package auth

import (
	"context"
	"net/http"
	"strings"
)

var ctxKey = "authToken"

func WrapHTTP(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer"))
		if tok == "" {
			next.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, r.Clone(context.WithValue(r.Context(), ctxKey, tok)))
	}
}

func TokenFromCtx(ctx context.Context) string {
	tok, ok := ctx.Value(ctxKey).(string)

	if !ok {
		return ""
	}

	return tok
}
