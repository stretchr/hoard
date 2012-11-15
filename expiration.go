package hoard

import (
	"time"
)

// ExpirationFunc is a type for the function signature of a custom expiration
// function.
type ExpirationFunc func() bool

// ExpiresNever is an Expiration that indicates the objects never expire.
var ExpiresNever *Expiration = nil

// Expiration describes when an object or objects will expire.
type Expiration struct {
	// idle is the sliding window for expiration in nanoseconds
	idle time.Duration

	// absolute is the absolute expiration time
	absolute time.Time

	// condition is the function to call to determine if an expiration
	// condition is met
	condition ExpirationFunc
}

// Expires creates a new empty Expiration object.
//
// Example
//
//     hoard.Expires().AfterSeconds(2)
func Expires() *Expiration {
	return new(Expiration)
}

// AfterSeconds expires the item after "seconds" seconds have passed.
func (e *Expiration) AfterSeconds(seconds int64) *Expiration {
	e.absolute = time.Now().Add(time.Duration(seconds) * time.Second)
	return e
}

// ExpiresAfterHours expires the item after "hours" hours have passed
func (e *Expiration) AfterHours(hours int64) *Expiration {
	e.absolute = time.Now().Add(time.Duration(hours) * time.Hour)
	return e
}

// ExpiresAfterDays expires the item after "days" days have passed
func (e *Expiration) AfterDays(days int64) *Expiration {
	e.absolute = time.Now().Add(time.Duration(days) * time.Hour * 24)
	return e
}

// ExpiresAfterSecondsIdle expires the item if it hasn't been accessed for
// "seconds" seconds
func (e *Expiration) AfterSecondsIdle(seconds int64) *Expiration {
	e.idle = time.Duration(seconds) * time.Second
	return e
}

// ExpiresAfterHoursIdle expires the item if it hasn't been accessed for 
// "hours" hours
func (e *Expiration) AfterHoursIdle(hours int64) *Expiration {
	e.idle = time.Duration(hours) * time.Hour
	return e
}

// ExpiresAfterDaysIdle expires the item if it hasn't been accessed for
// "days" days
func (e *Expiration) AfterDaysIdle(days int64) *Expiration {
	e.idle = time.Duration(days) * time.Hour * 24
	return e
}

// ExpiresOnDate expires the item once "date" date has passed
func (e *Expiration) OnDate(date time.Time) *Expiration {
	e.absolute = date
	return e
}

// ExpiresOnCondition expires the item if the "condition" func returns true
// TODO: describe WHEN this is checked (i.e. after every Get?)
func (e *Expiration) OnCondition(condition ExpirationFunc) *Expiration {
	e.condition = condition
	return e
}
