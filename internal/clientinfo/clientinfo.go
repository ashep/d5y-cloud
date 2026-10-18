package clientinfo

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/ashep/d5y/internal/geoip"
	"github.com/bigdatacloudapi/bigdatacloud-go/bigdatacloud"
	"github.com/rs/zerolog"
)

type Info struct {
	RemoteAddr    string
	UserAgent     string
	ID            string
	Vendor        string
	Name          string
	Version       string
	Hardware      string
	Country       string
	City          string
	Timezone      string
	Lat           float64
	Lon           float64
	LatLngFromReq bool
}

type ctxKeyType string

const ctxKey ctxKeyType = "clientInfo"

func FromCtx(ctx context.Context) Info {
	v, _ := ctx.Value(ctxKey).(Info)
	return v
}

func FromRequest(req *http.Request, cache *Cache, bdc *bigdatacloud.Client, l zerolog.Logger) Info {
	info := Info{
		ID:         strings.TrimSpace(strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer")),
		RemoteAddr: req.Header.Get("cf-connecting-ip"),
		UserAgent:  req.UserAgent(),
	}

	if info.RemoteAddr == "" {
		info.RemoteAddr = req.Header.Get("x-forwarded-for")
	}
	if info.RemoteAddr == "" {
		info.RemoteAddr = req.RemoteAddr
	}
	if idx := strings.LastIndex(info.RemoteAddr, ":"); idx != -1 { // strip port
		info.RemoteAddr = info.RemoteAddr[:idx]
	}

	if info.RemoteAddr != "" {
		if v, ok := cache.Get(info.RemoteAddr); ok && v.LatLngFromReq {
			return v
		}
	}

	if ua := strings.Split(info.UserAgent, ":"); len(ua) == 4 {
		info.Vendor = ua[0]
		info.Name = ua[1]
		info.Hardware = ua[2]
		info.Version = ua[3]
	}

	return cache.Set(enrichGeoData(req, info, bdc, l))
}

func enrichGeoData(req *http.Request, info Info, bdc *bigdatacloud.Client, l zerolog.Logger) Info {
	if info.RemoteAddr == "" {
		return info
	}

	if lat, lng := parseLatLonFromRequest(req); lat != 0 && lng != 0 {
		if data, err := bdc.ReverseGeocoding.ReverseGeocodeWithTimezone(req.Context(), lat, lng, "en"); err != nil {
			l.Error().Err(err).Msg("reverse geo lookup failed")
		} else {
			info.Country = data.CountryName
			info.City = data.City
			info.Timezone = data.TimeZone.IANATimeID
			info.Lat = lat
			info.Lon = lng
			info.LatLngFromReq = true
			return info
		}
	}

	if data, err := geoip.Get(info.RemoteAddr); err != nil {
		l.Error().Err(err).Msg("geoip lookup failed")
	} else {
		info.Country = data.CountryName
		info.City = data.City
		info.Timezone = data.Timezone
		info.Lat = data.Lat
		info.Lon = data.Lon
		info.LatLngFromReq = false
	}

	return info
}

func parseLatLonFromRequest(req *http.Request) (lat, lng float64) {
	q := req.URL.Query()

	latS, lngS := q.Get("lat"), q.Get("lng")
	if latS == "" || lngS == "" {
		return
	}

	lat, _ = strconv.ParseFloat(latS, 64)
	lng, _ = strconv.ParseFloat(lngS, 64)

	return
}
