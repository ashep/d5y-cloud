package alertsinua

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ashep/d5y/internal/httpcli"
	"github.com/rs/zerolog"
)

const (
	regionUnknown = iota
	regionKyiv
	regionKyivska
	regionMykolaivska
	regionZakarpatska
	regionZaporizka
)

type Alert struct {
	City   string `json:"location_title"`
	Region string `json:"location_oblast"`
}

type activeAlertsResponse struct {
	Alerts []Alert `json:"alerts"`
}

type Client struct {
	c            *httpcli.Client
	activeAlerts map[int]Alert
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

func (c *Client) Start(ctx context.Context) error {
	if err := c.refresh(); err != nil {
		return fmt.Errorf("get active alerts: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Minute):
				if err := c.refresh(); err != nil {
					c.l.Error().Err(err).Msg("get active air raid alerts failed")
					continue
				}
			}
		}
	}()

	return nil
}

func (c *Client) refresh() error {
	res := activeAlertsResponse{}
	if err := c.c.GetJSON("https://api.alerts.in.ua/v1/alerts/active.json", &res); err != nil {
		return fmt.Errorf("get json: %w", err)
	}

	newAlerts := make(map[int]Alert)
	for _, alert := range res.Alerts {
		regID := regionID(alert.Region, alert.City)
		if regID == regionUnknown {
			continue
		}

		// We count only regions, while alerts.in.ua provides per-city alerts.
		if _, exists := newAlerts[regID]; !exists {
			newAlerts[regID] = alert
		}
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	curRegions := make([]int, 0, len(c.activeAlerts))
	for k := range c.activeAlerts {
		curRegions = append(curRegions, k)
	}

	// Delete expired alerts
	for _, regID := range curRegions {
		if _, exists := newAlerts[regID]; !exists {
			c.l.Info().Str("region", c.activeAlerts[regID].Region).Msg("air raid alert expired")
			delete(c.activeAlerts, regID)
		}
	}

	// Add new alerts
	for k, v := range newAlerts {
		if _, exists := c.activeAlerts[k]; !exists {
			c.l.Info().Str("region", v.Region).Msg("new air raid alert")
			c.activeAlerts[k] = v
		}
	}

	return nil
}

func (c *Client) IsActiveInRegion(region, city string) bool {
	c.mux.RLock()
	defer c.mux.RUnlock()

	_, ok := c.activeAlerts[regionID(region, city)]

	return ok
}

func regionID(region, city string) int {
	if city == "Kyiv" || city == "м. Київ" {
		return regionKyiv
	}

	// TODO: add missing regions:
	// https://devs.alerts.in.ua/#modeluid
	// https://simplemaps.com/data/ua-cities
	switch region {
	case "Kyyivs'ka Oblast'", "Київська область":
		return regionKyivska
	case "Mykolayivs'ka Oblast'", "Миколаївська область":
		return regionMykolaivska
	case "Zakarpats'ka Oblast'", "Закарпатська область":
		return regionZakarpatska
	case "Zaporiz'ka Oblast'", "Запорізька область":
		return regionZaporizka
	default:
		return regionUnknown
	}
}
