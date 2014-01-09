# Hoard - Cache in your Chips

A fast, smart caching package for Go.

##Overview
Hoard provides easy-to-use, high performance cache management capabilities to your Go programs.

###API Documentation

  * Jump right into the [API documentation](http://go.pkgdoc.org/github.com/stretchr/hoard)

###When should I use Hoard?
Caching is useful if:

  * You have objects that are expensive/slow to create, such as reading from a database or the web.
  * Your program could benefit from lazy loading of resources.
  
###How does Hoard work?

The first time you need an object, Hoard will ask you to create it.  It will then store the object you provide in memory until it expires.  If your code needs it again, it will be returned from the cache.  If it has already expired, Hoard will ask you to create it again and store the result in the cache.

Internally, Hoard manages the expiration of objects in a performant manner, and allows you to specify specific policies for when an object should expire.

###What kind of expiration does Hoard support?

When you ask Hoard to cache an object, you can specify how long it should be kept before it is deleted.

  * `hoard.ExpiresNever` - the object will _never_ expire
  * `hoard.ExpiresDefault` - the object will inherit the default expiration policy
  
You can tell Hoard to expire an object after a specific amount of time has passed by using the `After*` funcs:
  
  * `hoard.Expires().AfterSeconds(s)` - the object will expire after `s` seconds
  * `hoard.Expires().AfterMinutes(m)` - the object will expire after `m` minutes
  * `hoard.Expires().AfterHours(h)` - the object will expire after `h` hours
  * `hoard.Expires().AfterDays(d)` - the object will expire after `d` days
  
You can tell Hoard to expire an object after a specific amount of time has passed since it was last accessed by using the `After*Idle` funcs:
  
  * `hoard.Expires().AfterSecondsIdle(s)` - the object will expire `s` seconds after it was last accessed
  * `hoard.Expires().AfterMinutesIdle(m)` - the object will expire `m` minutes after it was last accessed
  * `hoard.Expires().AfterHoursIdle(h)` - the object will expire `h` hours after it was last accessed
  * `hoard.Expires().AfterDaysIdle(d)` - the object will expire `d` days after it was last accessed
  
You can specify an actual `time.Time` date at which the object should expire:

  * `hoard.Expires().OnDate(t)` - the `time.Time` when the object will expire
  
Or you can write your own expiry function using the `hoard.Expires().OnCondition(f)` func.
  
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
Hoard's `Get` method also provides a much simpler alternative that removes a lot of common code.  Passing a `func` (of type `DataGetter`) as the second argument tells Hoard how to get the object if it doesn't have it in its cache. 

    func GetSomething() *Something {

      return hoard.Get("my-key", func() (interface{}, *hoard.Expiration) {
      
        // get the object and return it
        obj := SomeExpensiveMethodToGetTheObject()
      
        // return the object (and tell it to never expire)
        return obj, hoard.ExpiresNever
      
      }).(*Something)

    }

Remember, because the function is declared inline, variables defined around it will be available (via closures) making it easy to do other initialisation work in the `DataGetter`.

####DataGetter and DataGetterWithError
The `DataGetter` type is defined as:

    type DataGetter func() (interface{}, *Expiration)

The function takes no arguments, but must return an object (the object you intend to cache), and a `hoard.Expiration` instance describing when the object should expire (see [Expiring](#expiring) below).  For indefinite expiration (i.e. once it's created it should never expire) you can use the handy `hoard.ExpiresNever` object.

The `DataGetterWithError` type is defined as:

	type DataGetterWithError func() (interface{}, error, *Expiration)

It serves the same purpose as the `DataGetter` type, except that it allows you to return an error as seen in the next section.

####Returning Data and Error
For the common case of methods that return an error as the second argument, Hoard provides the `GetWithError` alternative that works as you might expect:

    func GetSomething() (*Something, error) {

      obj, err := hoard.GetWithError("my-key", func() (interface{}, error, *hoard.Expiration) {
      
        // get the object and return it
        obj, err := SomeExpensiveFunctionToGetTheObject()
        
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

If the `SomeExpensiveFunctionToGetTheObject` function returns an error, nothing will be cached and next time the `GetSomething` function is called, it will try again.

##Expiring
Hoard can automatically expire objects depending on the expiration policy you provide when placing the object in the cache.

###Expiring by default
If you opt to create your own Hoard instance using `MakeHoard`, you can pass a default expiration policy to it.

When you use the `Set` method to add an object to your hoard and your omit the Expiration argument, the default policy will be applied.

When you use the `Get` method and define a function to provide the data, you must return `hoard.ExpiresDefault` to instruct Hoard to use the default expiration policy you initially set.

To override the default, simply pass an expiration as normal.

###Expiring using `Set`
If you are using the `Set` method directly, the third argument is a `hoard.Expiration` object that describes the conditions under which the object should be removed from the cache.

To tell hoard that the `object` expires after 20 seconds, you can do:

    hoard.Set(key, object, hoard.Expires().AfterSeconds(20))

###Expiring using `Get` alternative

If you are using the special `Get` alternative, then you return the `Expiration` object as the last argument.  To tell the object to expire after 20 seconds you might do:

    func GetSomething() *Something {

      return hoard.Get("my-key", func() (interface{}, *hoard.Expiration) {
      
        // get the object and return it
        obj, err := SomeExpensiveFunctionToGetTheObject()
      
        // return the object (and tell it to expire after 20 seconds)
        return obj, err, hoard.Expires().AfterSeconds(20)
      
      }).(*Something)
    
    }

###Complex expiration policies

It is possible to combine certain expiration settings to form a complex policy, and this is done by chaining the method calls on an `Expiration` object.

For example, if we want our data to expire after being idle for twenty minutes, or after an hour regardless, we can do:

    return obj, hoard.Expires().AfterMinutesIdle(20).AfterHours(1)

##Design patterns

We recommend that you write a wrapper `struct` that manages your hoards and provides strongly-typed interfaces to access your objects.  This not only improves your own APIs (even if you never intend on sharing your code) but also means all of your caching code will be in one place, instead of peppered throughout.

For example, if we wanted to read a page of Tweets in our Go program, and only wanted to update it every twenty minutes, we could write a `TweetService` struct.

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

When our program calls `GetTweets()` the first time, Hoard will run the code that loads the tweets and store the response in memory for twenty minutes.  During that time, all calls to `GetTweets()` will be instantaneous, as the object will be returned directly from memory.  After 20 minutes, the object will be deleted, and the next call to `GetTweets()` will load the updated data from Twitter.

##The shared Hoard and your own Hoards

Hoard provides a common shared cache called `hoard.Shared` that you can use from anywhere in your application.  All global funcs (`hoard.Get`, `hoard.Set`, `hoard.Has` etc.) will interact with the shared hoard instance. The shared instance has a default policy of `ExpiresNever`, so you must always use your own expiration policy if you wish the data to expire.

If you'd like to create discrete `Hoard` objects, simply use `Make`:

    h1 := Make(hoard.ExpiresNever) // Never expire by default
    h2 := Make(hoard.Expires().AfterDays(1)) // Expire after one day by default

When you create a `Hoard` instance, all the same functions are available and they will operate on the instance explicitly.

##Installation
To install hoard, just do:

    go get github.com/stretchr/hoard

Updating is as simple as:

    go get -u github.com/stretchr/hoard
    
    
##Contributing

Please feel free to submit issues, fork the repository and send pull requests!

When submitting an issue, we ask that you please include a complete test function that demonstrates the issue.

------

Licence
=======
Copyright (c) 2012 - 2013 Mat Ryer and Tyler Bunnell

Please consider promoting this project if you find it useful.

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
