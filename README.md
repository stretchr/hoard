# Hoard

A fast, smart caching package for Go.

##Get started
###Manual cache management
Hoard provides simple `Has`, `Get` and `Set` methods to enable you to work with objects by a key:

    func GetSomething() *Something {

      // do we have this in the cache?
      if !hoard.SharedHoard().Has("my-key") {
  
      	// get the object and store it
      	obj := SomeExpensiveMethodToGetTheObject()
      	hoard.SharedHoard().Set("my-key", obj)
  
      }
  
      // return the object from the cache
      return hoard.SharedHoard().Get("my-key").(*Something)

    }

###Hoard's special `Get` alternative
But Hoard's `Get` method also provides a nicer alternative to remove a lot of common code.  Passing a `func` as the second argument tells Hoard how to get the object if it doesn't have it in its cache. 

    func GetSomething() *Something {

      return hoard.SharedHoard().Get("my-key", func() (interface{}, *hoard.Expiration) {
    	
    	// get the object and return it
    	obj := SomeExpensiveMethodToGetTheObject()
    	
    	// return the object (and tell it to never expire)
    	return obj, hoard.ExpiresNever
    	
      }).(*Something)

    }

####With errors
For the common case of methods that return an optional error as the second argument, Hoard provides the `GetWithError` alternative that works as you might expect:

    func GetSomething() (*Something, error) {

      obj, err := hoard.SharedHoard().GetWithError("my-key", func() (interface{}, error, *hoard.Expiration) {
    	
      	// get the object and return it
      	obj, err := SomeExpensiveMethodToGetTheObject()
      	
      	// return the object (and tell it to never expire)
      	return obj, err, hoard.ExpiresNever
      	
      })
      
      // did it return an error?
      if err != nil {
      	return nil, err
      }
      
      // all is well
      return obj, nil

    }

If the `SomeExpensiveMethodToGetTheObject` method returns an error, nothing will be cached and next time the `GetSomething` func is called, it will try again.

##Expiring
Hoard can automatically expire objects depending on the expiration policy you assign when you `Set` the object in the cache.

###Expiring using `Set`
If you are using the `Set` method directly, the third argument is a `hoard.Expiration` object that describes the conditions under which the object should be removed from the cache.

To tell hoard that the `object` expires after 20 seconds, you can do:

    hoard.SharedHoard().Set(key, object, hoard.Expires().AfterSeconds(20))

###Expiring using `Get` alternative

If you are using the special `Get` alternative, then you return the `Expiration` object as the last argument.  To tell the object to expire after 20 seconds you might do:

    func GetSomething() *Something {

      obj, err := hoard.SharedHoard().GetWithError("my-key", func() (interface{}, error, *hoard.Expiration) {
    	
    	// get the object and return it
    	obj, err := SomeExpensiveMethodToGetTheObject()
    	
    	// return the object (and tell it to never expire)
    	return obj, err, hoard.ExpiresNever
    	
      })
      
      // did it return an error?
      if err != nil {
      	return nil, err, hoard.ExpiresNever
      }
      
      // all is well
      return obj, nil, hoard.Expires().AfterSeconds(20)

    }

##`SharedHoard` and your own Hoards

Hoard provides a common shared cache called `hoard.SharedHoard` that you can use from anywhere in your application, but you may decide to create your own storage (or multiple ones) which is as easy as calling `MakeHoard`:

    h1 := MakeHoard(hoard.ExpiresNever)
    h2 := MakeHoard(hoard.ExpiresNever)
    
Notice you can pass a default expiration policy object into the `MakeHoard` func that will be applied to all objects.

##Installation
To install hoard, just do:

    go get http://github.com/stretchrcom/hoard

Updating is as simple as:

    go get -u http://github.com/stretchrcom/hoard
