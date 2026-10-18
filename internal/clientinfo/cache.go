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

func (c *Cache) Get(k string) (Info, bool) {
	return c.c.GetIfPresent(k)
}

func (c *Cache) Set(info Info) Info {
	// Location reported by device has priority over IP-address-based one
	cur, curOk := c.c.GetIfPresent(info.RemoteAddr)
	if curOk && !cur.LatLngFromReq && info.LatLngFromReq {
		c.c.Invalidate(info.RemoteAddr)
	}

	// Here the item replaced only if it is expired
	info, _ = c.c.Set(info.RemoteAddr, info)

	return info
}
