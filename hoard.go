package hoard

import (
	"sync"
	"time"
)

// container contains the cached data as well as metadata for the caching engine.
type container struct {
	// key is the string key at which this object is cached.
	key string

	// data is the actual cached data.
	data interface{}

	// accessed is the time this entry was last accessed.
	accessed time.Time

	// expiration holds the expiration properties for this object.
	expiration *Expiration
}

// Hoard is the object through which all caching happens.
type Hoard struct {
	// cache is a map containing the container objects.
	cache map[string]container

	// expirationCache is a map containing container objects.
	expirationCache map[string]container

	// defaultExpiration is an expiration object applied to all objects that
	// do not explicitly provide an expiration.
	defaultExpiration *Expiration

	// ticker controls how often the flush check is run.
	ticker *time.Ticker

	// tickerRunning stores whether the ticker is running or not.
	tickerRunning bool

	// cacheDeadbolt is used to lock the cache object.
	cacheDeadbolt sync.RWMutex

	// expirationDeadbolt is used to lock the expirationCache object.
	expirationDeadbolt sync.RWMutex

	// tickerRunningDeadbolt is used to lock the ticker object.
	tickerRunningDeadbolt sync.Mutex
}

// startFlushManager starts the ticker to check for expired objects and
// flushes those that are expired.
func (h *Hoard) startFlushManager() {

	if !h.getTickerRunning() {
		h.setTickerRunning(true)

		h.ticker = time.NewTicker(1 * time.Second)

		go func() {
			for currentTime := range h.ticker.C {

				var expirations []string

				if len(h.expirationCache) != 0 {

					h.expirationDeadbolt.RLock()

					for key, value := range h.expirationCache {

						if value.expiration != nil {
							if value.expiration.IsExpiredByTime(value.accessed, currentTime) {
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

// expireInternal removes the item with the specified key from the expiration cache.
func (h *Hoard) expireInternal(key string) {
	h.expirationDeadbolt.Lock()
	delete(h.expirationCache, key)
	h.expirationDeadbolt.Unlock()
}

// DataGetter is a type for the function signature used to place data into the
// caching system from the "Get" method.
type DataGetter func() (interface{}, *Expiration)

// DataGetterWithError is a type for the function signature used to place data
// into the caching system (and handling an error) from the "Get" method.
type DataGetterWithError func() (interface{}, error, *Expiration)

// Make creates a new *Hoard object. This function must be used to create
// a hoard object as it readies various internal fields.
func Make(defaultExpiration *Expiration) *Hoard {

	h := new(Hoard)

	h.cache = make(map[string]container)
	h.expirationCache = make(map[string]container)
	h.defaultExpiration = defaultExpiration

	return h

}

/* Get is the most concise way to put data into the cache. However, it only
* works for single return values.
* This method performs several functions:
* 1. If the key is in cache, it returns the object immediately and updates the last accessed time.
* 2. If the key is not in the cache:
*	a. It retrieves the object and expiration properties from the dataGetter.
*	b. It creates a container object in which to store the data and metadata.
*	c. It stores the new container in the cache.
*	d. It stores the container in the expirationCache if necessary.
*
* If no dataGetter is passed and the key is not in the cache, returns nil. */
func (h *Hoard) Get(key string, dataGetter ...DataGetter) interface{} {

	var data interface{}
	object, ok := h.cacheGet(key)

	if ok {
		// The object exists, but may be expired
		if object.expiration != nil {
			if object.expiration.IsExpiredByCondition() {
				Remove(object.key)
				ok = false
			}
		}
	}

	if !ok {
		if len(dataGetter) == 0 {
			// The object wasn't in cache and there is no dataGetter
			return nil
		}

		var expiration *Expiration

		data, expiration = dataGetter[0]()

		if expiration == ExpiresDefault {
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
func (h *Hoard) GetWithError(key string, dataGetterWithError ...DataGetterWithError) (interface{}, error) {

	var data interface{}
	object, ok := h.cacheGet(key)

	if ok {
		// The object exists, but may be expired
		if object.expiration.IsExpiredByCondition() {
			Remove(object.key)
			ok = false
		}
	}

	if !ok {
		if len(dataGetterWithError) == 0 {
			return nil, nil
		}

		var expiration *Expiration
		var err error

		data, err, expiration = dataGetterWithError[0]()

		if err != nil {
			return data, err
		}

		if expiration == ExpiresDefault {
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

// Set stores an object in cache for the given key and starts the flush manager
// if it isn't already running.
func (h *Hoard) Set(key string, object interface{}, expiration ...*Expiration) {
	var exp *Expiration

	if len(expiration) == 0 {
		exp = h.defaultExpiration
	} else {
		exp = expiration[0]
	}

	containerObject := container{key, object, time.Now(), exp}
	h.cacheSet(key, containerObject)

	if exp != ExpiresNever {
		h.expirationCacheSet(key, containerObject)
		h.startFlushManager()
	}
}

// Has returns whether the key exists in the cache or not.
func (h *Hoard) Has(key string) bool {

	_, ok := h.cacheGet(key)
	return ok

}

// Remove removes the item with the specified key from the cache.
func (h *Hoard) Remove(key string) {
	h.cacheDeadbolt.Lock()
	delete(h.cache, key)
	h.cacheDeadbolt.Unlock()
	h.expireInternal(key)
}

// SetExpires updates the expiration policy for the object of the
// specified key.
//
// If the expiration is ExpiresNever, removes this item from the
// expirationCache
func (h *Hoard) SetExpires(key string, expiration *Expiration) bool {

	object, ok := h.cacheGet(key)
	if !ok {
		// not ok - we don't have this object
		return false
	}

	if expiration == ExpiresDefault {
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

// cacheGet retrieves an object from the cache atomically.
func (h *Hoard) cacheGet(key string) (container, bool) {

	h.cacheDeadbolt.RLock()
	object, ok := h.cache[key]
	h.cacheDeadbolt.RUnlock()
	return object, ok
}

// cacheSet sets an object in the cache atomically.
func (h *Hoard) cacheSet(key string, object container) {

	h.cacheDeadbolt.Lock()
	h.cache[key] = object
	h.cacheDeadbolt.Unlock()

}

// expirationCacheGet retrieves an object from the expirationCache atomically.
func (h *Hoard) expirationCacheGet(key string) (container, bool) {

	h.expirationDeadbolt.RLock()
	object, ok := h.expirationCache[key]
	h.expirationDeadbolt.RUnlock()
	return object, ok
}

// expirationCacheSet sets an object in the expirationCache atomically.
func (h *Hoard) expirationCacheSet(key string, object container) {

	h.expirationDeadbolt.Lock()
	h.expirationCache[key] = object
	h.expirationDeadbolt.Unlock()

}

// getTickerRunning retrieves the ticker running status atomically.
func (h *Hoard) getTickerRunning() bool {
	h.tickerRunningDeadbolt.Lock()
	tickerRunning := h.tickerRunning
	h.tickerRunningDeadbolt.Unlock()
	return tickerRunning
}

// setTickerRunning retrieves the ticker running status atomically.
func (h *Hoard) setTickerRunning(tickerRunning bool) {
	h.tickerRunningDeadbolt.Lock()
	h.tickerRunning = tickerRunning
	h.tickerRunningDeadbolt.Unlock()
}
