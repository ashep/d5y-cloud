package clientinfo

import (
	"time"

	"github.com/maypok86/otter/v2"
	_ "github.com/maypok86/otter/v2/stats"
)

type Cache struct {
	c *otter.Cache[string, Info]
}

func NewCache() *Cache {
	c := otter.Must(&otter.Options[string, Info]{
		MaximumSize:      1000,
		ExpiryCalculator: otter.ExpiryAccessing[string, Info](time.Hour),
	})

	return &Cache{c: c}
}

func (c *Cache) Get(id string) (Info, bool) {
	return c.c.GetIfPresent(id)
}

func (c *Cache) Set(id string, cliInfo Info) Info {
	// Location reported by device has priority over IP-address-based one
	cur, curOk := c.c.GetIfPresent(id)
	if curOk && !cur.LatLngFromReq && cliInfo.LatLngFromReq {
		c.c.Invalidate(id)
	}

	// Here the item replaced only if it is expired
	cliInfo, _ = c.c.Set(id, cliInfo)

	return cliInfo
}
