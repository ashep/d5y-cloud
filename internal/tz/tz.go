package tz

import (
	"bytes"
	"embed"
	"log"
	"strings"
	"sync"
)

var (
	//go:embed zoneinfo
	zoneInfo embed.FS
	cache    map[string]string
	mux      *sync.Mutex
)

const (
	utcPosixName = "UTC0"
)

func ToPosix(s string) string {
	if s == "" {
		return utcPosixName
	}

	if mux == nil {
		mux = new(sync.Mutex)
	}

	mux.Lock()
	defer mux.Unlock()

	if cache == nil {
		cache = make(map[string]string)
	}

	if res, ok := cache[s]; ok {
		return res
	}

	b, err := zoneInfo.ReadFile("zoneinfo/" + s)
	if err != nil {
		log.Printf("tz: failed to read zone data for %s: %v", s, err)
		return utcPosixName
	}

	bs := bytes.Split(b, []byte("\n"))
	if len(bs) < 2 {
		log.Printf("tz: failed to read zone data for %s: %v", s, err)
		return utcPosixName
	}

	res := strings.TrimSpace(string(bs[len(bs)-2]))
	cache[s] = res

	return res
}
