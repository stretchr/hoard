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

	// expiration holds the expiration properties for this object
	expiration *Expiration
}

// Hoard is a type that manages the actual caching
type Hoard struct {
	// cache is a map containing the container objects
	cache map[string]container

	// expirationCache is a map containing container objects
	expirationCache map[string]container

	// defaultExpiration is an expiration object applied to all objects that
	// do not explicitly provide an expiration
	defaultExpiration *Expiration

	// ticker controls how often the flush check is run
	ticker *time.Ticker

	// tickerRunning stores whether the ticker is started or not
	tickerRunning bool

	// cacheDeadbolt is used to lock the cache object
	cacheDeadbolt sync.RWMutex

	// expirationDeadbolt is used to lock the expirationCache object
	expirationDeadbolt sync.RWMutex

	// tickerDeadbolt is used to lock the ticker object
	tickerDeadbolt sync.Mutex
}

// HoardFunc is a type for the function signature used to place data into the 
// caching system, as well as an optional expiration set
type HoardFunc func() (interface{}, *Expiration)

// HoardFuncWithError is a type for the function signature used to place data into the 
// caching system, as well as an optional expiration set
type HoardFuncWithError func() (interface{}, error, *Expiration)

// StartFlushManager starts the ticker to check for expired entries and
// flushes those that are expired
func (h *Hoard) StartFlushManager() {

	if !h.getTickerRunning() {
		h.setTickerRunning(true)

		// Tick every second to check for expired data
		h.ticker = time.NewTicker(1 * time.Second)

		go func() {
			for currentTime := range h.ticker.C {

				var expirations []string

				if len(h.expirationCache) != 0 {

					h.expirationDeadbolt.RLock()

					for key, value := range h.expirationCache {

						if value.expiration != nil {
							if value.expiration.IsExpired(value.accessed, currentTime) {
								expirations = append(expirations, key)
							}
						}
					}

					h.expirationDeadbolt.RUnlock()

					if len(expirations) != 0 {

						h.cacheDeadbolt.Lock()
						h.expirationDeadbolt.Lock()
						for _, key := range expirations {
							delete(h.cache, key)
							delete(h.expirationCache, key)
						}
						h.cacheDeadbolt.Unlock()
						h.expirationDeadbolt.Unlock()

					}
				} else {
					h.ticker.Stop()
					h.setTickerRunning(false)
				}
			}
		}()
	}
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
	h.expirationCache = make(map[string]container)
	h.defaultExpiration = defaultExpiration

	return h

}

/* Get is the most concise way to put data into the cache. However, it only
* works when a single possible return value exists. Otherwise, use Set.
* This method performs several functions:
* 1. If the key is in cache, it returns the object immediately
* 2. If the key is not in the cache:
*	a. It retrieves the object and expiration properties from the hoardFunc
*	b. It creates a container object in which to store the data and metadata
*	c. It stores the new container in the cache
*
* If no hoardfunc is passed and the key is not in the cache, returns nil.
* Only the first hoardFunc is used. All others will be ignored */
func (h *Hoard) Get(key string, hoardFunc ...HoardFunc) interface{} {

	var data interface{}
	object, ok := h.cacheGet(key)

	if !ok {
		if len(hoardFunc) == 0 {
			return nil
		}

		var expiration *Expiration

		data, expiration = hoardFunc[0]()

		if expiration == nil {
			expiration = h.defaultExpiration
		}

		h.Set(key, data, expiration)

	} else {
		data = object.data
		object.accessed = time.Now()
		h.cacheSet(key, object)

		if object.expiration != nil && object.expiration != ExpiresNever {
			h.expirationCacheSet(key, object)
		}
	}

	return data

}

// GetWithError does the same as Get, but handles error cases. If an error is
// encountered, it does not cache anything and returns the error.
func (h *Hoard) GetWithError(key string, hoardFuncWithError ...HoardFuncWithError) (interface{}, error) {

	var data interface{}
	object, ok := h.cacheGet(key)

	if !ok {
		if len(hoardFuncWithError) == 0 {
			return nil, nil
		}

		var expiration *Expiration
		var err error

		data, err, expiration = hoardFuncWithError[0]()

		if err != nil {
			return data, err
		}

		if expiration == nil {
			expiration = h.defaultExpiration
		}

		h.Set(key, data, expiration)

	} else {
		data = object.data
		object.accessed = time.Now()
		h.cacheSet(key, object)

		if object.expiration != nil && object.expiration != ExpiresNever {
			h.expirationCacheSet(key, object)
		}
	}

	return data, nil

}

// Set stores an object in cache for the given key
// Also, checks to see if the expiration is set, and starts the expiration
// ticker if it is not already running, and adds the new item to the expiration
// cache.
func (h *Hoard) Set(key string, object interface{}, expiration ...*Expiration) {
	var exp *Expiration

	if len(expiration) == 0 {
		exp = h.defaultExpiration
	} else {
		exp = expiration[0]
	}

	containerObject := container{key, object, time.Now(), time.Now(), exp}
	h.cacheSet(key, containerObject)

	if exp != ExpiresNever {
		h.expirationCacheSet(key, containerObject)
		h.StartFlushManager()
	}
}

// Has determines whether 
func (h *Hoard) Has(key string) bool {

	_, ok := h.cacheGet(key)
	return ok

}

// Remove removes the item with the specified key from the map.
func (h *Hoard) Remove(key string) {
	h.cacheDeadbolt.Lock()
	delete(h.cache, key)
	h.cacheDeadbolt.Unlock()
	h.expireInternal(key)
}

// expireInternal removes the item with the specified key from the expiration cache.
func (h *Hoard) expireInternal(key string) {
	h.expirationDeadbolt.Lock()
	delete(h.expirationCache, key)
	h.expirationDeadbolt.Unlock()
}

// SetExpires updates the expiration policy for the object of the
// specified key.
// 
// If the expiration is ExpiresNever, removes this item from the 
// expirationCache
func (h *Hoard) SetExpires(key string, expiration *Expiration) bool {

	object, ok := h.cacheGet(key)
	if !ok {
		return false
	}

	if expiration == nil {
		expiration = h.defaultExpiration
	}

	// update the expiration policy
	object.expiration = expiration

	// set the object back in the cache
	h.cacheSet(key, object)

	if expiration == ExpiresNever || expiration == nil {
		h.expireInternal(key)
	} else {
		h.expirationCacheSet(key, object)
	}

	return true

}

// cacheGet retrieves an object from the cache atomically
func (h *Hoard) cacheGet(key string) (container, bool) {

	h.cacheDeadbolt.RLock()
	object, ok := h.cache[key]
	h.cacheDeadbolt.RUnlock()
	return object, ok
}

// cacheSet sets an object in the cache atomically
func (h *Hoard) cacheSet(key string, object container) {

	h.cacheDeadbolt.Lock()
	h.cache[key] = object
	h.cacheDeadbolt.Unlock()

}

// expirationCacheGet retrieves an object from the cache atomically
func (h *Hoard) expirationCacheGet(key string) (container, bool) {

	h.expirationDeadbolt.RLock()
	object, ok := h.expirationCache[key]
	h.expirationDeadbolt.RUnlock()
	return object, ok
}

// expirationCacheSet sets an object in the cache atomically
func (h *Hoard) expirationCacheSet(key string, object container) {

	h.expirationDeadbolt.Lock()
	h.expirationCache[key] = object
	h.expirationDeadbolt.Unlock()

}

// getTickerRunning retrieves the ticker running status atomically
func (h *Hoard) getTickerRunning() bool {
	h.tickerDeadbolt.Lock()
	tickerRunning := h.tickerRunning
	h.tickerDeadbolt.Unlock()
	return tickerRunning
}

// setTickerRunning retrieves the ticker running status atomically
func (h *Hoard) setTickerRunning(tickerRunning bool) {
	h.tickerDeadbolt.Lock()
	h.tickerRunning = tickerRunning
	h.tickerDeadbolt.Unlock()
}
