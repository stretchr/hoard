# Hoard

A fast, smart caching package for Go.

##Overview
Hoard provides easy-to-use, high performant cache management capabilities to your Go programs.

###API Documentation

  * Jump right into the [API documentation](http://go.pkgdoc.org/github.com/stretchrcom/hoard)

###When should I use Hoard?
Caching is useful if:

  * you have objects that are expensive or slow to create or initialise such as those read from a database or downloaded from the web etc.
  * you have a software architecture where lazy loading of resources is beneficial to your program
  * you care about, and intend to manage the amount of memory being used by your code
  
###How does Hoard work?

The first time you need an object, Hoard will ask you to create it.  It will then store the object in memory until it expires.  If your code needs it again (and it hasn't expired), it will be returned.  If it has already expired, Hoard will ask you to create it again the next time your code needs it.

Internally, Hoard manages the expiration of objects in a number of high-performant ways, and you are able to specify when objects should expire.

###What kind of expiration does Hoard support?

When you ask Hoard to cache an object, you can specify how long it should be kept before it is deleted.

  * `hoard.ExpiresNever` - the object will _never_ expire
  
You can tell Hoard to expire an object a specific amount of time after it is cached, using the `After*` funcs:
  
  * `hoard.Expires().AfterSeconds(s)` - the object will expire after `s` seconds
  * `hoard.Expires().AfterMinutes(m)` - the object will expire after `m` minutes
  * `hoard.Expires().AfterHours(h)` - the object will expire after `h` hours
  * `hoard.Expires().AfterDays(d)` - the object will expire after `d` days
  
You can tell Hoard to expire an object a specific amount of time after the last time an object is retrieved, using the `After*Idle` funcs:
  
  * `hoard.Expires().AfterSecondsIdle(s)` - the object will expire `s` seconds after it is last used
  * `hoard.Expires().AfterMinutesIdle(m)` - the object will expire `m` minutes after it is last used
  * `hoard.Expires().AfterHoursIdle(h)` - the object will expire `h` hours after it is last used
  * `hoard.Expires().AfterDaysIdle(d)` - the object will expire `d` days after it is last used
  
You can specify an actual `time.Time` date at which the object should expire:

  * `hoard.Expires().OnDate(t)` - the `time.Time` when the object will expire
  
Or you can write your own expiry func using the `hoard.Expires().OnCondition(f)` func.
  
##Get started
Hoard offers a few different ways to manage caching in your Go programs.  
###Manual cache management
Hoard provides simple `Has`, `Get` and `Set` methods to enable you to work with objects by a key:

    func GetSomething() *Something {

      // do we have this in the cache?
      if !hoard.Has("my-key") {
  
        // get the object and store it
        obj := SomeExpensiveMethodToGetTheObject()
        hoard.Set("my-key", obj)
  
      }
  
      // return the object from the cache
      return hoard.Get("my-key").(*Something)

    }

###Hoard's special `Get` alternative
Hoard's `Get` method also provides a much simpler alternative that removes a lot of common code.  Passing a `func` (of type `HoardFunc`) as the second argument tells Hoard how to get the object if it doesn't have it in its cache. 

    func GetSomething() *Something {

      return hoard.Get("my-key", func() (interface{}, *hoard.Expiration) {
      
      // get the object and return it
      obj := SomeExpensiveMethodToGetTheObject()
      
      // return the object (and tell it to never expire)
      return obj, hoard.ExpiresNever
      
      }).(*Something)

    }

Remember, because the func is declared inline, variables defined around it will be available (via closures) making it easy to do other initialisation work in the `HoardFunc`.

####HoardFunc type
The `HoardFunc` type is defined as:

    type HoardFunc func() (interface{}, *Expiration)

The func takes no arguments, but will return an object (the object being defined), and an `Expiration` instance describing when the object should expire (see Expiring below).  For indefinite expiration (i.e. once it's created it should never expire) you can use the handy `hoard.ExpiresNever` object.

####With errors
For the common case of methods that return an optional error as the second argument, Hoard provides the `GetWithError` alternative that works as you might expect:

    func GetSomething() (*Something, error) {

      obj, err := hoard.GetWithError("my-key", func() (interface{}, error, *hoard.Expiration) {
      
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
      return obj.(*Something), nil

    }

If the `SomeExpensiveMethodToGetTheObject` method returns an error, nothing will be cached and next time the `GetSomething` func is called, it will try again.

##Expiring
Hoard can automatically expire objects depending on the expiration policy you assign when you `Set` the object in the cache.

###Expiring using `Set`
If you are using the `Set` method directly, the third argument is a `hoard.Expiration` object that describes the conditions under which the object should be removed from the cache.

To tell hoard that the `object` expires after 20 seconds, you can do:

    hoard.Set(key, object, hoard.Expires().AfterSeconds(20))

###Expiring using `Get` alternative

If you are using the special `Get` alternative, then you return the `Expiration` object as the last argument.  To tell the object to expire after 20 seconds you might do:

    func GetSomething() *Something {

      return hoard.Get("my-key", func() (interface{}, *hoard.Expiration) {
      
        // get the object and return it
        obj, err := SomeExpensiveMethodToGetTheObject()
      
        // return the object (and tell it to never expire)
        return obj, err, hoard.ExpiresNever
      
      }).(*Something)
    
    }

###Complex expiration policies

It is possible to combine certain expiration settings to form a complex policy, and this is done by chaining the method calls on an `Expiration` object.

For example, if we want our tweets data to expire after twenty minutes of idle use, or after an hour regardless we could write this:

    return obj, hoard.Expires().AfterMinutesIdle(20).AfterHours(1)

##Design patterns

We recommend that you write a wrapper `struct` that manages your hoards and provides strongly-typed interfaces to access your objects.  This not only improves your own APIs (even if you never intend on sharing your code) but also means all of your caching code will be in one place, instead of peppered throughout your code.

For example, if we wanted to read a page of Tweets in our Go program, and wanted to only update it every twenty minutes, we could write a `TweetService` struct.

    // TweetService is responsible for loading tweets.
    type TweetService struct{}
    
    // GetTweets loads tweets via the twitter API
    func (t *TweetService) GetTweets() (tweets *Tweets, err error) {
    
      // wrap it in a Hoard call for easy caching
      tweets, err = hoard.GetWithError("tweets", func() (interface{}, error, *hoard.Expiration) {
      
        // load the tweets
        loadedTweets, loadingErr := TweetLoadingAPICallThatTakesAWhile()
        
        // just return the result - and include our expiration policy
        return loadedTweets, loadingErr, hoard.Expires().AfterMinutes(20)
      
      })
    
    }

When our program calls `GetTweets()` the first time, Hoard will run the code that loads the tweets, and store the response in memory for twenty minutes.  During that time, all calls to `GetTweets()` will be lightening fast, since the object will be returned directly from memory, instead of via the API again.  After 20 minutes, the object will be deleted, and the next call to `GetTweets()` will load the updated data from the API.

##`SharedHoard` and your own Hoards

Hoard provides a common shared cache called `hoard.SharedHoard` that you can use from anywhere in your application.  All global funcs (`hoard.Get`, `hoard.Set`, `hoard.Has` etc.) will interact with the shared hoard instance.

You may decide to create your own storage (or multiple ones) which is as easy as calling `MakeHoard`:

    h1 := MakeHoard(hoard.ExpiresNever)
    h2 := MakeHoard(hoard.ExpiresNever)

When you have an instance of the hoard, the same funcs are available as methods which will exclusively interact with the relevant `Hoard`.

Also, notice you can pass a default expiration policy object into the `MakeHoard` func that will be applied to all objects by default.

##Installation
To install hoard, just do:

    go get http://github.com/stretchrcom/hoard

Updating is as simple as:

    go get -u http://github.com/stretchrcom/hoard
