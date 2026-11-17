// Command genraions converts the OCHA Ukraine admin level 2 (raion) boundaries into a compact,
// simplified dataset embedded into the alertsinua package.
//
// Each raion also gets the list of its hromada administrative centers (cities, towns, villages),
// used to find the raion of city-level alerts.
//
// Source: "Ukraine - Subnational Administrative Boundaries" by OCHA, CC BY-IGO,
// https://data.humdata.org/dataset/cod-ab-ukr (files ukr_admin2.geojson and ukr_admincapitals.geojson
// from ukr_admin_boundaries.geojson.zip).
//
// Usage:
//
//	go run ./internal/alertsinua/genraions -in ukr_admin2.geojson -capitals ukr_admincapitals.geojson \
//		-out internal/alertsinua/data/raions.json.gz
package main

import (
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
)

type srcCollection struct {
	Features []struct {
		Properties struct {
			Oblast string `json:"adm1_name1"`
			Raion  string `json:"adm2_name1"`
			PCode  string `json:"adm2_pcode"`
		} `json:"properties"`
		Geometry struct {
			Type        string          `json:"type"`
			Coordinates json.RawMessage `json:"coordinates"`
		} `json:"geometry"`
	} `json:"features"`
}

type srcCapitals struct {
	Features []struct {
		Properties struct {
			Name  string `json:"name_1"`
			PCode string `json:"adm2_pcode"`
		} `json:"properties"`
	} `json:"features"`
}

// Raion is the output record. Rings holds all rings of all polygons as flat [lng, lat, lng, lat, ...] arrays.
type Raion struct {
	PCode  string      `json:"c"`
	Oblast string      `json:"o"`
	Raion  string      `json:"r"`
	Rings  [][]float64 `json:"g"`
	Cities []string    `json:"t,omitempty"`
}

func main() {
	in := flag.String("in", "", "path to ukr_admin2.geojson")
	capitals := flag.String("capitals", "", "path to ukr_admincapitals.geojson")
	out := flag.String("out", "", "output .json.gz path")
	tolerance := flag.Float64("tolerance", 0.001, "simplification tolerance, degrees")
	flag.Parse()

	if err := run(*in, *capitals, *out, *tolerance); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(in, capitals, out string, tolerance float64) error {
	b, err := os.ReadFile(in)
	if err != nil {
		return err
	}

	src := srcCollection{}
	if err := json.Unmarshal(b, &src); err != nil {
		return fmt.Errorf("unmarshal source: %w", err)
	}

	b, err = os.ReadFile(capitals)
	if err != nil {
		return err
	}

	caps := srcCapitals{}
	if err := json.Unmarshal(b, &caps); err != nil {
		return fmt.Errorf("unmarshal capitals: %w", err)
	}

	cities := make(map[string][]string)
	for _, f := range caps.Features {
		cities[f.Properties.PCode] = append(cities[f.Properties.PCode], f.Properties.Name)
	}

	var res []Raion
	inPts, outPts := 0, 0

	for _, f := range src.Features {
		var polys [][][][]float64

		switch f.Geometry.Type {
		case "Polygon":
			var p [][][]float64
			if err := json.Unmarshal(f.Geometry.Coordinates, &p); err != nil {
				return fmt.Errorf("%s: %w", f.Properties.PCode, err)
			}
			polys = [][][][]float64{p}
		case "MultiPolygon":
			if err := json.Unmarshal(f.Geometry.Coordinates, &polys); err != nil {
				return fmt.Errorf("%s: %w", f.Properties.PCode, err)
			}
		default:
			return fmt.Errorf("%s: unexpected geometry type %q", f.Properties.PCode, f.Geometry.Type)
		}

		r := Raion{
			PCode:  f.Properties.PCode,
			Oblast: f.Properties.Oblast,
			Raion:  f.Properties.Raion,
			Cities: cities[f.Properties.PCode],
		}
		sort.Strings(r.Cities)
		for _, poly := range polys {
			for _, ring := range poly {
				inPts += len(ring)
				s := simplify(ring, tolerance)
				if len(s) < 4 {
					continue // degenerate after simplification (tiny islands, etc.)
				}
				outPts += len(s)

				flat := make([]float64, 0, len(s)*2)
				for _, pt := range s {
					flat = append(flat, round(pt[0]), round(pt[1]))
				}
				r.Rings = append(r.Rings, flat)
			}
		}

		if len(r.Rings) == 0 {
			return fmt.Errorf("%s: no rings left after simplification", r.PCode)
		}

		res = append(res, r)
	}

	sort.Slice(res, func(i, j int) bool { return res[i].PCode < res[j].PCode })

	fh, err := os.Create(out)
	if err != nil {
		return err
	}

	zw, err := gzip.NewWriterLevel(fh, gzip.BestCompression)
	if err != nil {
		_ = fh.Close()
		return err
	}
	if err := json.NewEncoder(zw).Encode(res); err != nil {
		_ = fh.Close()
		return err
	}
	if err := zw.Close(); err != nil {
		_ = fh.Close()
		return err
	}
	if err := fh.Close(); err != nil {
		return err
	}

	fmt.Printf("raions: %d, points: %d -> %d\n", len(res), inPts, outPts)

	return nil
}

func round(v float64) float64 {
	return math.Round(v*1e4) / 1e4
}

// simplify applies the Douglas-Peucker algorithm to a closed ring.
func simplify(pts [][]float64, tolerance float64) [][]float64 {
	if len(pts) < 4 {
		return pts
	}

	keep := make([]bool, len(pts))
	keep[0], keep[len(pts)-1] = true, true

	// A closed ring has equal first and last points, so split it at the farthest point first.
	far, farD := 0, -1.0
	for i := 1; i < len(pts)-1; i++ {
		if d := dist(pts[i], pts[0]); d > farD {
			far, farD = i, d
		}
	}
	keep[far] = true

	dp(pts, 0, far, tolerance, keep)
	dp(pts, far, len(pts)-1, tolerance, keep)

	res := make([][]float64, 0)
	for i, k := range keep {
		if k {
			res = append(res, pts[i])
		}
	}

	return res
}

func dp(pts [][]float64, first, last int, tolerance float64, keep []bool) {
	if last-first < 2 {
		return
	}

	idx, maxD := -1, tolerance
	for i := first + 1; i < last; i++ {
		if d := segDist(pts[i], pts[first], pts[last]); d > maxD {
			idx, maxD = i, d
		}
	}

	if idx < 0 {
		return
	}

	keep[idx] = true
	dp(pts, first, idx, tolerance, keep)
	dp(pts, idx, last, tolerance, keep)
}

func dist(a, b []float64) float64 {
	return math.Hypot(a[0]-b[0], a[1]-b[1])
}

func segDist(p, a, b []float64) float64 {
	dx, dy := b[0]-a[0], b[1]-a[1]
	if dx == 0 && dy == 0 {
		return dist(p, a)
	}

	t := ((p[0]-a[0])*dx + (p[1]-a[1])*dy) / (dx*dx + dy*dy)
	t = math.Max(0, math.Min(1, t))

	return math.Hypot(p[0]-(a[0]+t*dx), p[1]-(a[1]+t*dy))
}
