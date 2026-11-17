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
	Title        string `json:"location_title"`
	LocationType string `json:"location_type"`
	LocationUID  string `json:"location_uid"`
	Oblast       string `json:"location_oblast"`
	AlertType    string `json:"alert_type"`
}

type activeAlertsResponse struct {
	Alerts []Alert `json:"alerts"`
}

type Client struct {
	c            *httpcli.Client
	activeAlerts map[int]Alert // keyed by location UID
	mux          *sync.RWMutex
	l            zerolog.Logger
}

func New(apiKey string, l zerolog.Logger) *Client {
	return &Client{
		c:            httpcli.New(httpcli.WithAuthorization(apiKey)),
		activeAlerts: make(map[int]Alert),
		mux:          &sync.RWMutex{},
		l:            l,
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

	newAlerts := airRaidAlertsByUID(res.Alerts)

	c.mux.Lock()
	defer c.mux.Unlock()

	for uid, a := range c.activeAlerts {
		if _, ok := newAlerts[uid]; !ok {
			c.l.Info().Int("uid", uid).Str("location", a.Title).Msg("air raid alert expired")
		}
	}

	for uid, a := range newAlerts {
		if _, ok := c.activeAlerts[uid]; !ok {
			c.l.Info().Int("uid", uid).Str("location", a.Title).Msg("new air raid alert")
		}
	}

	c.activeAlerts = newAlerts

	return nil
}

// airRaidAlertsByUID returns air raid alerts keyed by the UID of the oblast (or special-status city) they belong to.
//
// The client location is known only up to the oblast, while most alerts are declared per raion,
// so an oblast is considered alerted when it has an air raid alert for the whole oblast, any of its raions
// or cities. Hromada-level alerts are ignored: on the front line they last for months and would keep
// the whole oblast permanently alerted. Alerts of other types (shelling, etc.) are ignored as well.
func airRaidAlertsByUID(alerts []Alert) map[int]Alert {
	res := make(map[int]Alert)

	for _, a := range alerts {
		if a.AlertType != alertTypeAirRaid {
			continue
		}

		var uid int
		switch a.LocationType {
		case "oblast":
			var err error
			if uid, err = strconv.Atoi(a.LocationUID); err != nil {
				continue
			}
		case "raion", "city":
			// Kyiv and Sevastopol are special-status cities with their own UIDs; location_oblast points to them.
			uid = locationUID(a.Oblast, "")
		default:
			continue
		}

		if uid == uidUnknown {
			continue
		}

		// Prefer the whole-oblast alert for logging.
		if prev, ok := res[uid]; ok && prev.LocationType == "oblast" {
			continue
		}

		res[uid] = a
	}

	return res
}

// IsActiveInRegion reports whether an air raid alert is declared anywhere in the oblast
// (or the special-status city, like Kyiv) the given location belongs to, see airRaidAlertsByUID.
func (c *Client) IsActiveInRegion(region, city string) bool {
	uid := locationUID(region, city)
	if uid == uidUnknown {
		return false
	}

	c.mux.RLock()
	defer c.mux.RUnlock()

	_, ok := c.activeAlerts[uid]

	return ok
}
