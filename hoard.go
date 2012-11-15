package hoard

import (
	"sync"
	"time"
)

// container contains the cached data as well as metadata for the caching engine
type container struct {
	// key is the string key at which this object is cached
	key string

	// data is the actual cached data
	data interface{}

	// created is the time this entry was first created
	created time.Time

	// accessed is the time this entry was last accessed
	accessed time.Time

	// expiration holds the expiration properties for this item
	expiration *Expiration
}

// Hoard is a type that manages the actual caching
type Hoard struct {
	// cache is a map containing the container objects
	cache map[string]container

	// defaultExpiration is an expiration object applied to all objects that
	// do not explicitly provide an expiration
	defaultExpiration *Expiration

	// ticker controls how often the flush check is run
	ticker *time.Ticker

	// deadbolt is used to lock the cache object
	deadbolt sync.RWMutex
}

// HoardFunc is a type for the function signature used to place data into the 
// caching system, as well as an optional expiration set
type HoardFunc func() (interface{}, *Expiration)

// StartFlushManager starts the ticker to check for expired entries and
// flushes those that are expired
func (h *Hoard) StartFlushManager() {

	// Tick every second to check for expired data
	h.ticker = time.NewTicker(1 * time.Second)

	go func() {
		for currentTime := range h.ticker.C {

			var expirations []string

			if len(h.cache) != 0 {

				h.deadbolt.RLock()

				for key, value := range h.cache {

					if value.expiration != nil {
						if value.expiration.IsExpired(value.accessed, currentTime) {
							expirations = append(expirations, key)
						}
					}
				}

				h.deadbolt.RUnlock()

				if len(expirations) != 0 {

					h.deadbolt.Lock()
					for _, key := range expirations {
						delete(h.cache, key)
					}
					h.deadbolt.Unlock()

				}
			}
		}
	}()
}

var sharedHoard *Hoard
var initOnce sync.Once

// SharedHoard returns a shared hoard object
// The shared Hoard object does not have a default expiration policy
func SharedHoard() *Hoard {

	initOnce.Do(func() {
		sharedHoard = MakeHoard(ExpiresNever)
	})

	return sharedHoard

}

// MakeHoard creates a new hoard object
func MakeHoard(defaultExpiration *Expiration) *Hoard {

	h := new(Hoard)

	h.cache = make(map[string]container)
	h.defaultExpiration = defaultExpiration
	h.StartFlushManager()

	return h

}

/* Get performs several functions:
* 1. If the key is in cache, it returns the object immediately
* 2. If the key is not in the cache:
*	a. It retrieves the object and expiration properties from the hoardFunc
*	b. It creates a container object in which to store the data and metadata
*	c. It stores the new container in the cache */
func (h *Hoard) Get(key string, hoardFunc HoardFunc) interface{} {

	item, ok := h.cacheGet(key)

	if !ok {
		data, expiration := hoardFunc()

		if expiration == nil {
			expiration = h.defaultExpiration
		}

		item = container{key, data, time.Now(), time.Now(), expiration}
		h.cacheSet(key, item)

	} else {
		item.accessed = time.Now()
		h.cacheSet(key, item)
	}

	return item.data

}

// Expire removes the item from the map
func (h *Hoard) Expire(item container) {

	h.deadbolt.Lock()
	delete(h.cache, item.key)
	h.deadbolt.Unlock()
}

// cacheGet retrieves an item from the cache atomically
func (h *Hoard) cacheGet(key string) (container, bool) {
	h.deadbolt.RLock()
	item, ok := h.cache[key]
	h.deadbolt.RUnlock()
	return item, ok
}

// cacheSet sets an item in the cache atomically
func (h *Hoard) cacheSet(key string, item container) {
	h.deadbolt.Lock()
	h.cache[key] = item
	h.deadbolt.Unlock()
}
