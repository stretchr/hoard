package hoard

import (
	"sync"
	"time"
)

// container contains the cached data as well as metadata for the caching engine.
type container struct {

	// data is the actual cached data.
	data interface{}

	// accessed is the time this entry was last accessed.
	accessed time.Time

	// created is the time this entry was added to the cache
	created time.Time

	// expiration holds the expiration properties for this object.
	expiration *Expiration
}

// expirationContainer only contains the metadata for the caching engine
type expirationContainer struct {

	// accessed is the time this entry was last accessed.
	accessed time.Time

	// created is the time this entry was added to the cache
	created time.Time

	// expiration holds the expiration properties for this object.
	expiration *Expiration
}

// cloneExpirationContainer returns a copy of the container without the data payload
func (c *container) cloneExpirationContainer() expirationContainer {
	return expirationContainer{
		accessed:   c.accessed,
		created:    c.created,
		expiration: c.expiration,
	}
}

// Hoard is the object through which all caching happens.
//
// Hoard manages caching data by key, as well as managing the expiration
// of said data based on the expiration policy you provide.
//
// The flushing system will be started on demand, and will be terminated when
// there is no more work to do.
type Hoard struct {
	// cache is a map containing the container objects.
	cache map[string]container

	// expirationCache is a map containing container objects.
	expirationCache map[string]expirationContainer

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

	// keyDeadbolts hold a mutex for each key to provide thread safety for
	// multiple thread access and reentrant calls
	keyDeadbolts map[string]*sync.Mutex

	// keyDeadbolt provides thread safety for the keyDeadbolts map
	keyDeadbolt sync.Mutex

	// interval between expiration checks performed by startFlushManager()
	expirationCheckInterval time.Duration
}

// startFlushManager starts the ticker to check for expired objects and
// flushes those that are expired.
func (h *Hoard) startFlushManager() {

	if !h.getTickerRunning() {
		h.setTickerRunning(true)

		h.ticker = time.NewTicker(h.expirationCheckInterval)

		go func() {
			for currentTime := range h.ticker.C {
				var expirations []string

				if len(h.expirationCache) != 0 {

					h.expirationDeadbolt.RLock()

					for key, value := range h.expirationCache {

						if value.expiration != nil {
							if value.expiration.isExpiredAbsolute(currentTime) {
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

// expirationCacheSet sets an object in the expirationCache atomically.
func (h *Hoard) expirationCacheSet(key string, object container) {

	// get expiratíonConatiner without data payload
	expirationContainer := object.cloneExpirationContainer()

	// make sure the expiration has set its absolute time correctly.
	// Because expiration is a pointer to an expiration shared with the object in normal cache, both will be updated
	expirationContainer.expiration.updateAbsoluteTime(object.accessed, object.created)

	h.expirationDeadbolt.Lock()
	h.expirationCache[key] = expirationContainer
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

// DataGetter is a type for the function signature used to place data into the
// caching system from the "Get" method.
type DataGetter func() (interface{}, *Expiration)

// DataGetterWithError is a type for the function signature used to place data
// into the caching system (and handling an error) from the "Get" method.
type DataGetterWithError func() (interface{}, error, *Expiration)

// Make creates a new *Hoard object. This function must be used to create
// a hoard object as it readies various internal fields.
//
// If a Hoard object is created using new(), it will panic as soon as you
// attempt to use it.
func Make(defaultExpiration *Expiration) *Hoard {

	h := new(Hoard)

	h.cache = make(map[string]container)
	h.expirationCache = make(map[string]expirationContainer)
	h.defaultExpiration = defaultExpiration
	h.keyDeadbolts = make(map[string]*sync.Mutex)
	h.expirationCheckInterval = time.Second

	return h

}

// SetExpirationCheckInterval sets the time interval to wait between checking
// all expirable objects in the cache  and flushing expired ones.
//
// Default is one second.
//
// This function will not change an already running ticker, it should therefore
// preferably be called right after Make()
func (h *Hoard) SetExpirationCheckInterval(d time.Duration) *Hoard {
	h.expirationCheckInterval = d
	return h
}

// Get retrieves data from the cache using the key provided.
//
// If a dataGetter func is passed as the second argument, the Get method uses
// it to ask the calling code to provide data to be cached. This is the most
// concise and idomatic way of placing data in the cache.
//
// A DataGetter calling Get with the same key as the key for which the
// DataGetter is called, the system will deadlock. It is best to avoid calling
// Get from within a DataGetter unless you make sure the same key is not used
// twice.
//
// The dataGetter function only works for methods that return a single value.
// If your code needs to return a value and an error, use the GetWithError
// method.
//
// If no dataGetter is passed and the key is not in the cache, Get returns nil.
func (h *Hoard) Get(key string, dataGetter ...DataGetter) interface{} {

	var data interface{}
	object, ok := h.cacheGet(key)
	expired := false

	if ok {
		// The object exists, but may be expired
		if object.expiration != nil {
			if object.expiration.IsExpired(object.accessed, object.created) { // need to check for expiration by time and condition, because h.expirationCheckInterval could be relatively large compared to objects expire time
				Remove(key)
				expired = true
			}
		}
	}

	// Short circuit for quick retrieval
	if ok && !expired {
		data = object.data
		object.accessed = time.Now()
		h.cacheSet(key, object)

		if object.expiration != nil && object.expiration != ExpiresNever {
			h.expirationCacheSet(key, object)
		}

		return data
	}

	// We need to make a deadbolt for this key if one doesn't exist
	h.keyDeadbolt.Lock()
	if _, keyDeadboltExists := h.keyDeadbolts[key]; !keyDeadboltExists {
		if _, exists := h.keyDeadbolts[key]; !exists {
			h.keyDeadbolts[key] = &sync.Mutex{}
		}
	}

	keyDeadbolt := h.keyDeadbolts[key]
	h.keyDeadbolt.Unlock()

	// defer the unlock to account for early exits.
	defer func() {
		keyDeadbolt.Unlock()

		// delete key specific deadbolt to avoid mutexes piling up
		h.keyDeadbolt.Lock()
		delete(h.keyDeadbolts, key)
		h.keyDeadbolt.Unlock()
	}()

	// We need to lock this section to prevent multiple threads from calling
	// the getter method more than once
	keyDeadbolt.Lock()

	// Now we need to make sure that the data we are seeking wasn't retrieved
	// by another thread, and that it hasn't been expired in that time

	object, ok = h.cacheGet(key)
	if ok {
		// The object exists, but may be expired
		if object.expiration != nil {
			if object.expiration.IsExpired(object.accessed, object.created) { // need to check for expiration by time and condition, because h.expirationCheckInterval could be relatively large compared to objects expire time
				Remove(key)
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
	}

	return data

}

// GetWithError operates the same way as Get, but handles error cases.
//
// Please refer to the documentation for the Get method for more information on
// usage and unsupported behavior.
//
// If an error is encountered, the data and error are returned directly and
// no caching is done.
func (h *Hoard) GetWithError(key string, dataGetterWithError ...DataGetterWithError) (interface{}, error) {

	var data interface{}
	object, ok := h.cacheGet(key)
	expired := false

	if ok {
		// The object exists, but may be expired
		if object.expiration != nil {
			if object.expiration.IsExpired(object.accessed, object.created) { // need to check for expiration by time and condition, because h.expirationCheckInterval could be relatively large compared to objects expire time
				Remove(key)
				expired = true
			}
		}
	}

	// Short circuit for quick retrieval
	if ok && !expired {
		data = object.data
		object.accessed = time.Now()
		h.cacheSet(key, object)

		if object.expiration != nil && object.expiration != ExpiresNever {
			h.expirationCacheSet(key, object)
		}
		return data, nil
	}

	// We need to make a deadbolt for this key if one doesn't exist
	h.keyDeadbolt.Lock()
	if _, keyDeadboltExists := h.keyDeadbolts[key]; !keyDeadboltExists {
		if _, exists := h.keyDeadbolts[key]; !exists {
			h.keyDeadbolts[key] = &sync.Mutex{}
		}
	}
	keyDeadbolt := h.keyDeadbolts[key]
	h.keyDeadbolt.Unlock()

	// defer the unlock to account for early exits.
	defer func() {
		keyDeadbolt.Unlock()

		// delete key specific deadbolt to avoid mutexes piling up
		h.keyDeadbolt.Lock()
		delete(h.keyDeadbolts, key)
		h.keyDeadbolt.Unlock()
	}()

	// We need to lock this section to prevent multiple threads from calling
	// the getter method more than once
	keyDeadbolt.Lock()

	// Now we need to make sure that the data we are seeking wasn't retrieved
	// by another thread, and that it hasn't been expired in that time

	object, ok = h.cacheGet(key)
	if ok {
		// The object exists, but may be expired
		if object.expiration != nil {
			if object.expiration.IsExpired(object.accessed, object.created) { // need to check for expiration by time and condition, because h.expirationCheckInterval could be relatively large compared to objects expire time
				Remove(key)
				ok = false
			}
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
	}

	return data, nil

}

// Set stores an object in cache for the given key.
//
// The third argument, expiration, is optional. If it is not provided, the
// default expiration policy for this instance will be used.
func (h *Hoard) Set(key string, object interface{}, expiration ...*Expiration) {
	var exp *Expiration

	if len(expiration) == 0 {
		exp = h.defaultExpiration
	} else {
		exp = expiration[0]
	}

	containerObject := container{object, time.Now(), time.Now(), exp}
	h.cacheSet(key, containerObject)

	if exp != nil && exp != ExpiresNever {
		h.expirationCacheSet(key, containerObject)
		h.startFlushManager()
	}
}

// Has returns whether or not the key exists in the cache.
func (h *Hoard) Has(key string) bool {

	_, ok := h.cacheGet(key)
	return ok

}

// Remove removes an object by key from the cache.
func (h *Hoard) Remove(key string) {
	h.cacheDeadbolt.Lock()
	delete(h.cache, key)
	h.cacheDeadbolt.Unlock()
	h.expireInternal(key)
}

// SetExpires updates the expiration policy for the object of the
// specified key.
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
