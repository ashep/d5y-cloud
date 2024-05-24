package remoteaddr

import (
	"net/http"
)

func FromRequest(r *http.Request) string {
	res := r.Header.Get("cf-connecting-ip")

	if res == "" {
		res = r.Header.Get("x-forwarded-for")
	}

	if res == "" {
		res = r.RemoteAddr
	}

	return res
}
