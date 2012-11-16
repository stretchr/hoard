package hoard

import (
	"sync"
)

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

/*
	Global shortcut methods that just access the SharedHoard
*/

// Get gets a value from the shared hoard.
//
// This is a shortcut func, see the Hoard funcs for more details.
func Get(key string, hoardFunc ...HoardFunc) interface{} {
	return SharedHoard().Get(key, hoardFunc...)
}

// GetWithError gets a value (or error) from the shared hoard.
//
// This is a shortcut func, see the Hoard funcs for more details.
func GetWithError(key string, hoardFuncWithError ...HoardFuncWithError) (interface{}, error) {
	return SharedHoard().GetWithError(key, hoardFuncWithError...)
}

// Remove removes an object by key form the shared hoard.
//
// This is a shortcut func, see the Hoard funcs for more details.
func Remove(key string) {
	SharedHoard().Remove(key)
}

// SetExpires updates the expiration policy for the item with the specified key in
// the shared hoard.
//
// This is a shortcut func, see the Hoard funcs for more details.
func SetExpires(key string, expiration *Expiration) bool {
	return SharedHoard().SetExpires(key, expiration)
}

// Set adds (or overwrites) an object in the shared hoard.
//
// This is a shortcut func, see the Hoard funcs for more details.
func Set(key string, object interface{}, expiration ...*Expiration) {
	SharedHoard().Set(key, object, expiration...)
}

// Has gets whether the object exists in the shared hoard.
//
// This is a shortcut func, see the Hoard funcs for more details.
func Has(key string) bool {
	return SharedHoard().Has(key)
}
