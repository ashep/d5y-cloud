package geoip

import (
	"encoding/json"
	"sync"

	"github.com/ashep/d5y/internal/httpcli"
)

var (
	cli   = httpcli.New()
	cache = make(map[string]*Data)
	mux   = &sync.Mutex{}
)

type Data struct {
	City        string  `json:"city,omitempty"`
	CountryCode string  `json:"countryCode,omitempty"`
	CountryName string  `json:"country,omitempty"`
	IP          string  `json:"ip,omitempty"`
	Latitude    float64 `json:"lat,omitempty"`
	Longitude   float64 `json:"lon,omitempty"`
	RegionCode  string  `json:"region,omitempty"`
	RegionName  string  `json:"regionName,omitempty"`
	Timezone    string  `json:"timezone,omitempty"`
}

func (d *Data) String() string {
	b, err := json.Marshal(d)
	if err != nil {
		return err.Error()
	}

	return string(b)
}

func Get(addr string) (*Data, error) {
	mux.Lock()
	defer mux.Unlock()

	d, ok := cache[addr]
	if ok {
		return d, nil
	}

	d = &Data{}

	if err := cli.GetJSON("http://ip-api.com/json/"+addr, d); err != nil {
		return nil, err
	}

	cache[addr] = d

	return d, nil
}
