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
		ExpiryCalculator: otter.ExpiryWriting[string, Info](time.Hour),
	})

	return &Cache{c: c}
}

func (c *Cache) Get(key string) (Info, bool) {
	return c.c.GetIfPresent(key)
}

func (c *Cache) Set(key string, info Info) {
	c.c.Set(key, info)
}
