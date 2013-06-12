// A smart, fast caching system for  Go
//
// Hoard allows you to cache any kind of data and retrieve it at any time.
// Hoard also allows you to specify an expiration policy for any data you cache.
// This makes it simple to control the amount of data you cache, as well as how
// long your data stays in cache before needing to be refreshed.
//
// Example Usage
//
// The following is a complete example of loading data into the cache using the
// special Get approach and the global shared Hoard instance:
//    func GetSomething() *Something {
//
//      return hoard.Get("my-key", func() (interface{}, *hoard.Expiration) {
//
//      // get the object and return it
//      obj := SomeExpensiveMethodToGetTheObject()
//
//      // return the object (and tell it to never expire)
//      return obj, hoard.ExpiresNever
//
//      }).(*Something)
//
//    }
//
// There are alternative ways to place data in the cache, such as a standard
// "Set" method. Please see the documentation of the individual functions for
// more details.
//
// Also, more examples and information can be found at https://github.com/stretchr/hoard
package hoard
