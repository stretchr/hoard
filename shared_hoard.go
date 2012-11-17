package hoard

import (
	"sync"
)

// sharedHoard stores the singleton *Hoard instance.
var sharedHoard *Hoard

// initOnce is used to guarantee that the sharedHoard is initialized only once.
var initOnce sync.Once

// Shared returns the global, shared Hoard object.
//
// Using the shared Hoard object and the associated global functions is the
// easiest way to get data into the cache. It is recommended you always use
// the shared Hoard instance unless you know you need your own discrete
// instance.
//
// The shared Hoard object has a default expiration policy of ExpiresNever.
// If you wish your data to expire, you must provide your own policy to
// override the ExpiresNever default.
func Shared() *Hoard {

	initOnce.Do(func() {
		sharedHoard = Make(ExpiresNever)
	})

	return sharedHoard

}

/*
	Global shortcut functions for accessing the shared Hoard object
*/

// Get gets a value from the shared hoard.
//
// This is a shortcut function, see the Hoard methods for more details.
func Get(key string, dataGetter ...DataGetter) interface{} {
	return Shared().Get(key, dataGetter...)
}

// GetWithError gets a value (with error) from the shared hoard.
//
// This is a shortcut function, see the Hoard methods for more details.
func GetWithError(key string, dataGetterWithError ...DataGetterWithError) (interface{}, error) {
	return Shared().GetWithError(key, dataGetterWithError...)
}

// Remove removes an object by key from the shared hoard.
//
// This is a shortcut function, see the Hoard methods for more details.
func Remove(key string) {
	Shared().Remove(key)
}

// SetExpires updates the expiration policy for the item with the specified key in
// the shared hoard.
//
// This is a shortcut function, see the Hoard methods for more details.
func SetExpires(key string, expiration *Expiration) bool {
	return Shared().SetExpires(key, expiration)
}

// Set adds (or overwrites) an object to the shared hoard.
//
// This is a shortcut function, see the Hoard methods for more details.
func Set(key string, object interface{}, expiration ...*Expiration) {
	Shared().Set(key, object, expiration...)
}

// Has gets whether the object exists in the shared hoard.
//
// This is a shortcut function, see the Hoard methods for more details.
func Has(key string) bool {
	return Shared().Has(key)
}
