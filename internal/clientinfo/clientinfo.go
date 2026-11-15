package clientinfo

import (
	"context"
	"crypto/sha1"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ashep/d5y/internal/geoip"
	"github.com/bigdatacloudapi/bigdatacloud-go/bigdatacloud"
	"github.com/rs/zerolog"
)

type Info struct {
	RemoteAddr          string
	UserAgent           string
	ID                  string
	Vendor              string
	Name                string
	Version             string
	Hardware            string
	Country             string
	City                string
	Timezone            string
	Lat                 float64
	Lon                 float64
	LocationSetByDevice bool
}

type ctxKeyType string

const ctxKey ctxKeyType = "clientInfo"

func FromCtx(ctx context.Context) Info {
	v, _ := ctx.Value(ctxKey).(Info)
	return v
}

func FromRequest(req *http.Request, cache *Cache, bdc *bigdatacloud.Client, l zerolog.Logger) Info {
	addr := req.Header.Get("cf-connecting-ip")
	if addr == "" {
		addr = req.Header.Get("x-forwarded-for")
	}
	if addr == "" {
		addr = req.RemoteAddr
	}
	if host, _, err := net.SplitHostPort(addr); err == nil {
		addr = host
	}

	addrSum := sha1.Sum([]byte(addr))

	clientID := strings.TrimSpace(strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer"))
	if clientID == "" {
		clientID = fmt.Sprintf("ip-%x", addrSum[:6])
	}

	lat, lng := parseLocationFromQueryString(req.URL.Query())
	cKey := fmt.Sprintf("%s:%x:%s:%f:%f", clientID, addrSum, req.UserAgent(), lat, lng)

	if v, ok := cache.Get(cKey); ok {
		return v
	}

	info := withUserAgent(Info{ID: clientID}, req, addr)

	if info = withGeolocation(info, req, bdc, l); info.LocationSetByDevice {
		// Cache only cases where location is told by device to avoid excessive requests to reverse geocoding service
		cache.Set(cKey, info)
	}

	return info
}

// withUserAgent sets user-agent-specific data from request
func withUserAgent(info Info, req *http.Request, addr string) Info {
	info.RemoteAddr = addr
	info.UserAgent = req.UserAgent()

	info.Vendor, info.Name, info.Hardware, info.Version = "", "", "", ""
	if ua := strings.Split(info.UserAgent, ":"); len(ua) == 4 {
		info.Vendor = ua[0]
		info.Name = ua[1]
		info.Hardware = ua[2]
		info.Version = ua[3]
	}

	return info
}

// withGeolocation tries to enrich Info with geolocation data.
func withGeolocation(info Info, req *http.Request, bdc *bigdatacloud.Client, l zerolog.Logger) Info {
	if info.RemoteAddr == "" {
		return info
	}

	// If the remote device tells us its location, use it
	info.LocationSetByDevice = false
	if lat, lng := parseLocationFromQueryString(req.URL.Query()); lat != 0 && lng != 0 {
		if data, err := bdc.ReverseGeocoding.ReverseGeocodeWithTimezone(req.Context(), lat, lng, "en"); err != nil {
			l.Error().Err(err).Msg("reverse geo lookup failed")
		} else {
			info.Country = data.CountryName
			info.City = data.City
			info.Timezone = data.TimeZone.IANATimeID
			info.Lat = lat
			info.Lon = lng
			info.LocationSetByDevice = true
			return info
		}
	}

	// If the remote device do not tell us its location, try to detect it by IP address
	if data, err := geoip.Get(info.RemoteAddr); err != nil {
		l.Error().Err(err).Msg("geoip lookup failed")
	} else {
		info.Country = data.CountryName
		info.City = data.City
		info.Timezone = data.Timezone
		info.Lat = data.Lat
		info.Lon = data.Lon
	}

	return info
}

func parseLocationFromQueryString(q url.Values) (lat, lng float64) {
	latS, lngS := q.Get("lat"), q.Get("lng")
	if latS == "" || lngS == "" {
		return
	}

	lat, _ = strconv.ParseFloat(latS, 64)
	lng, _ = strconv.ParseFloat(lngS, 64)

	return
}
