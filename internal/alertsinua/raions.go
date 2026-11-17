package alertsinua

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
)

// Raion boundaries, simplified from "Ukraine - Subnational Administrative Boundaries" by OCHA (CC BY-IGO),
// https://data.humdata.org/dataset/cod-ab-ukr. Regenerate with ./genraions.
//
//go:embed data/raions.json.gz
var raionsData []byte

// Place is a client location resolved to the alerts.in.ua oblast UID and, if possible, the raion.
type Place struct {
	OblastUID int
	Raion     string // Ukrainian raion name without the "район" word; empty if the raion is unknown
}

type raion struct {
	PCode     string      `json:"c"`
	Oblast    string      `json:"o"`
	Name      string      `json:"r"`
	Rings     [][]float64 `json:"g"`
	Cities    []string    `json:"t"`
	oblastUID int
	key       string
	minLng    float64
	minLat    float64
	maxLng    float64
	maxLat    float64
}

var (
	raionsOnce sync.Once
	raions     []raion
	raionsErr  error
)

func loadRaions() ([]raion, error) {
	raionsOnce.Do(func() {
		zr, err := gzip.NewReader(bytes.NewReader(raionsData))
		if err != nil {
			raionsErr = fmt.Errorf("open raions data: %w", err)
			return
		}

		var rs []raion
		if err := json.NewDecoder(zr).Decode(&rs); err != nil {
			raionsErr = fmt.Errorf("decode raions data: %w", err)
			return
		}

		for i := range rs {
			r := &rs[i]

			r.oblastUID = locationUID(r.Oblast, "")
			if r.oblastUID == uidUnknown {
				raionsErr = fmt.Errorf("raion %s: unknown oblast %q", r.PCode, r.Oblast)
				return
			}
			r.key = raionKey(r.Name)

			r.minLng, r.minLat, r.maxLng, r.maxLat = math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
			for _, ring := range r.Rings {
				for j := 0; j+1 < len(ring); j += 2 {
					r.minLng, r.maxLng = math.Min(r.minLng, ring[j]), math.Max(r.maxLng, ring[j])
					r.minLat, r.maxLat = math.Min(r.minLat, ring[j+1]), math.Max(r.maxLat, ring[j+1])
				}
			}
		}

		raions = rs
	})

	return raions, raionsErr
}

// cityKey normalizes a settlement name, dropping a settlement type prefix like "м." or "смт".
func cityKey(name string) string {
	name = strings.TrimSpace(name)
	for _, p := range []string{"м.", "смт.", "смт", "с.", "селище", "місто"} {
		if strings.HasPrefix(name, p+" ") {
			name = strings.TrimPrefix(name, p+" ")
			break
		}
	}

	return normalize(name)
}

// raionKey normalizes a raion name, so "Кам'янський район" from alerts.in.ua and "Кам’янський" from OCHA match.
func raionKey(name string) string {
	return strings.TrimSuffix(normalize(name), "район")
}

// findRaion returns the raion containing the point, or nil.
func findRaion(lat, lng float64) *raion {
	rs, err := loadRaions()
	if err != nil {
		return nil
	}

	for i := range rs {
		r := &rs[i]
		if lng < r.minLng || lng > r.maxLng || lat < r.minLat || lat > r.maxLat {
			continue
		}

		// Even-odd rule over all rings handles both multipolygons and holes.
		in := false
		for _, ring := range r.Rings {
			if ringContains(ring, lat, lng) {
				in = !in
			}
		}

		if in {
			return r
		}
	}

	return nil
}

func ringContains(ring []float64, lat, lng float64) bool {
	in := false
	n := len(ring) / 2

	for i, j := 0, n-1; i < n; j, i = i, i+1 {
		xi, yi := ring[i*2], ring[i*2+1]
		xj, yj := ring[j*2], ring[j*2+1]

		if (yi > lat) != (yj > lat) && lng < (xj-xi)*(lat-yi)/(yj-yi)+xi {
			in = !in
		}
	}

	return in
}

// Locate resolves the client location. Coordinates are preferred; if they are unknown or outside
// the dataset, the oblast is resolved from the region and city names and the raion is left empty.
func Locate(lat, lng float64, region, city string) Place {
	if lat != 0 || lng != 0 {
		if r := findRaion(lat, lng); r != nil {
			return Place{OblastUID: r.oblastUID, Raion: r.Name}
		}
	}

	return Place{OblastUID: locationUID(region, city)}
}
