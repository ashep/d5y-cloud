package alertsinua

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ashep/d5y/internal/httpcli"
	"github.com/rs/zerolog"
)

const (
	activeAlertsURL = "https://api.alerts.in.ua/v1/alerts/active.json"

	// alerts.in.ua allows 8-10 requests per minute per IP, so 2 per minute is well within the limit.
	refreshInterval = 30 * time.Second

	alertTypeAirRaid = "air_raid"
)

type Alert struct {
	ID           int64  `json:"id"`
	Title        string `json:"location_title"`
	LocationType string `json:"location_type"`
	LocationUID  string `json:"location_uid"`
	Oblast       string `json:"location_oblast"`
	Raion        string `json:"location_raion"`
	AlertType    string `json:"alert_type"`
}

type activeAlertsResponse struct {
	Alerts []Alert `json:"alerts"`
}

// alertState is the set of active air raid alerts indexed for lookups by client place.
type alertState struct {
	// Alerts for whole oblasts and for Kyiv / Sevastopol, keyed by UID.
	oblast map[int]Alert
	// Alerts for raions and cities, keyed by oblast UID and raion key.
	raion map[int]map[string]Alert
	// Oblasts having any raion or city alert; used for clients whose raion is unknown.
	partial map[int]Alert
	// All alerts taken into account, keyed by ID; used for logging.
	all map[int64]Alert
}

type Client struct {
	c     *httpcli.Client
	state alertState
	mux   *sync.RWMutex
	l     zerolog.Logger
}

func New(apiKey string, l zerolog.Logger) *Client {
	if _, err := loadRaions(); err != nil {
		l.Error().Err(err).Msg("load raions failed, raion-level alerts are disabled")
	}

	return &Client{
		c:     httpcli.New(httpcli.WithAuthorization(apiKey)),
		state: buildState(nil, l),
		mux:   &sync.RWMutex{},
		l:     l,
	}
}

// Start performs the initial refresh and starts periodic refreshing in the background.
// The background refreshing is started even if the initial refresh fails; the error is returned for reporting only.
func (c *Client) Start(ctx context.Context) error {
	err := c.refresh()

	go func() {
		t := time.NewTicker(refreshInterval)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if err := c.refresh(); err != nil {
					c.l.Error().Err(err).Msg("get active air raid alerts failed")
				}
			}
		}
	}()

	if err != nil {
		return fmt.Errorf("get active alerts: %w", err)
	}

	return nil
}

func (c *Client) refresh() error {
	res := activeAlertsResponse{}
	if err := c.c.GetJSON(activeAlertsURL, &res); err != nil {
		return fmt.Errorf("get json: %w", err)
	}

	newState := buildState(res.Alerts, c.l)

	c.mux.Lock()
	defer c.mux.Unlock()

	for id, a := range c.state.all {
		if _, ok := newState.all[id]; !ok {
			c.l.Info().Int64("id", id).Str("oblast", a.Oblast).Str("location", a.Title).Msg("air raid alert expired")
		}
	}

	for id, a := range newState.all {
		if _, ok := c.state.all[id]; !ok {
			c.l.Info().Int64("id", id).Str("oblast", a.Oblast).Str("location", a.Title).Msg("new air raid alert")
		}
	}

	c.state = newState

	return nil
}

// buildState indexes air raid alerts. Alerts of other types (shelling, etc.) are ignored.
//
// Hromada-level alerts are ignored as well: on the front line they last for months
// and would keep the whole raion permanently alerted.
func buildState(alerts []Alert, l zerolog.Logger) alertState {
	st := alertState{
		oblast:  make(map[int]Alert),
		raion:   make(map[int]map[string]Alert),
		partial: make(map[int]Alert),
		all:     make(map[int64]Alert),
	}

	known, cities := raionIndex()

	for _, a := range alerts {
		if a.AlertType != alertTypeAirRaid {
			continue
		}

		uid, _ := strconv.Atoi(a.LocationUID)

		switch {
		case a.LocationType == "oblast" && uid != uidUnknown:
			st.oblast[uid] = a

		case a.LocationType == "city" && (uid == uidKyiv || uid == uidSevastopol):
			st.oblast[uid] = a

		case a.LocationType == "raion" || a.LocationType == "city":
			obUID := locationUID(a.Oblast, "")
			if obUID == uidUnknown {
				l.Warn().Int64("id", a.ID).Str("oblast", a.Oblast).Str("location", a.Title).Msg("unknown alert oblast")
				continue
			}

			name := a.Raion
			key := raionKey(name)
			switch {
			case a.LocationType == "raion":
				name = a.Title
				key = raionKey(name)
			case key == "":
				// City alerts come without location_raion, so look the city up among hromada centers.
				key = cities[obUID][cityKey(a.Title)]
			}

			if !known[obUID][key] {
				// Better a false alert in the whole oblast than a missed one.
				l.Warn().Int64("id", a.ID).Str("oblast", a.Oblast).Str("raion", name).Str("location", a.Title).
					Msg("unknown alert raion, treating as oblast-wide")
				st.oblast[obUID] = a
			} else {
				if st.raion[obUID] == nil {
					st.raion[obUID] = make(map[string]Alert)
				}
				st.raion[obUID][key] = a
			}

			st.partial[obUID] = a

		default:
			continue
		}

		st.all[a.ID] = a
	}

	return st
}

// raionIndex returns raion keys by oblast UID and raion keys by oblast UID and city key.
// Cities whose name is not unique within the oblast are left out.
func raionIndex() (map[int]map[string]bool, map[int]map[string]string) {
	known := make(map[int]map[string]bool)
	cities := make(map[int]map[string]string)
	ambiguous := make(map[int]map[string]bool)

	rs, _ := loadRaions()
	for _, r := range rs {
		if known[r.oblastUID] == nil {
			known[r.oblastUID] = make(map[string]bool)
			cities[r.oblastUID] = make(map[string]string)
			ambiguous[r.oblastUID] = make(map[string]bool)
		}
		known[r.oblastUID][r.key] = true

		for _, c := range r.Cities {
			ck := cityKey(c)
			if prev, ok := cities[r.oblastUID][ck]; ok && prev != r.key {
				ambiguous[r.oblastUID][ck] = true
			}
			cities[r.oblastUID][ck] = r.key
		}
	}

	for ob, m := range ambiguous {
		for ck := range m {
			delete(cities[ob], ck)
		}
	}

	return known, cities
}

// IsActive reports whether there is an air raid alert at the place: for the whole oblast or the raion.
// If the raion is unknown, an alert in any raion of the oblast counts.
func (c *Client) IsActive(p Place) bool {
	if p.OblastUID == uidUnknown {
		return false
	}

	c.mux.RLock()
	defer c.mux.RUnlock()

	return c.state.isActive(p)
}

func (st alertState) isActive(p Place) bool {
	if _, ok := st.oblast[p.OblastUID]; ok {
		return true
	}

	if p.Raion == "" {
		_, ok := st.partial[p.OblastUID]
		return ok
	}

	_, ok := st.raion[p.OblastUID][raionKey(p.Raion)]

	return ok
}
