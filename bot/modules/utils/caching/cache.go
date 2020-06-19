package caching

import (
	"time"

	"github.com/anandpskerala/Martha/bot/modules/utils/error_handling"
	"github.com/allegro/bigcache"
)

var CACHE *bigcache.BigCache

func InitCache() {
	config := bigcache.Config{
		Shards:             1024,
		LifeWindow:         1 * time.Minute, // Life time of 1 minute
		CleanWindow:        1 * time.Minute,  // delete dead entries in one minute
		MaxEntriesInWindow: 1000 * 10 * 60,
		MaxEntrySize:       500,
		HardMaxCacheSize:   8192,
		OnRemove:           nil,
		OnRemoveWithReason: nil,
	}
	cache, err := bigcache.NewBigCache(config)
	error_handling.HandleErr(err)
	CACHE = cache
}
