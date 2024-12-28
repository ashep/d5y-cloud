package clientinfo

import (
	"context"
	"net/http"
	"strings"

	"github.com/ashep/d5y/internal/geoip"
	"github.com/rs/zerolog"
)

type Info struct {
	RemoteAddr string
	UserAgent  string
	ID         string
	Vendor     string
	Name       string
	Version    string
	Hardware   string
	Country    string
	City       string
	Timezone   string
}

type ctxKeyType string

const ctxKey ctxKeyType = "clientInfo"

func FromCtx(ctx context.Context) Info {
	v, _ := ctx.Value(ctxKey).(Info)
	return v
}

func FromRequest(req *http.Request, l zerolog.Logger) Info {
	res := Info{
		ID:         strings.TrimSpace(strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer")),
		RemoteAddr: req.Header.Get("cf-connecting-ip"),
		UserAgent:  req.UserAgent(),
	}

	if res.RemoteAddr == "" {
		res.RemoteAddr = req.Header.Get("x-forwarded-for")
	}
	if res.RemoteAddr == "" {
		res.RemoteAddr = req.RemoteAddr
	}

	uas := strings.Split(res.UserAgent, ":")
	if len(uas) == 4 {
		res.Vendor = uas[0]
		res.Name = uas[1]
		res.Hardware = uas[2]
		res.Version = uas[3]
	}

	gi, err := geoip.Get(res.RemoteAddr)
	if err != nil {
		l.Error().Err(err).Msg("geoip lookup failed")
	} else {
		res.Country = gi.CountryName
		res.City = gi.City
		res.Timezone = gi.Timezone
	}

	return res
}
