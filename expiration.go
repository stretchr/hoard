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

// IsExpired determines if an expiration object has expired
func (e *Expiration) IsExpired(lastAccess, currentTime time.Time) bool {
	if e.idle != 0 {
		if currentTime.Sub(lastAccess) >
			e.idle {
			return true
		}
	}
	if !e.absolute.IsZero() {
		if currentTime.After(e.absolute) {
			return true
		}
	}
	if e.condition != nil {
		if e.condition() {
			return true
		}
	}
	return false
}

// after does the work for each After* function
func (e *Expiration) after(duration int64, multiplier time.Duration) time.Time {
	return time.Now().Add(time.Duration(duration) * multiplier)
}

// AfterSeconds expires the item after "seconds" seconds have passed.
func (e *Expiration) AfterSeconds(seconds int64) *Expiration {
	e.absolute = e.after(seconds, time.Second)
	return e
}

// AfterMinutes expires the item after "minutes" minutes have passed.
func (e *Expiration) AfterMinutes(minutes int64) *Expiration {
	e.absolute = e.after(minutes, time.Minute)
	return e
}

// AfterHours expires the item after "hours" hours have passed
func (e *Expiration) AfterHours(hours int64) *Expiration {
	e.absolute = e.after(hours, time.Hour)
	return e
}

// AfterDays expires the item after "days" days have passed
func (e *Expiration) AfterDays(days int64) *Expiration {
	e.absolute = e.after(days, time.Hour*24)
	return e
}

// afterIdle does the work for each After*Idle function
func (e *Expiration) afterIdle(duration int64, multiplier time.Duration) time.Duration {
	return time.Duration(duration) * multiplier
}

// AfterSecondsIdle expires the item if it hasn't been accessed for
// "seconds" seconds
func (e *Expiration) AfterSecondsIdle(seconds int64) *Expiration {
	e.idle = e.afterIdle(seconds, time.Second)
	return e
}

// AfterMinutesIdle expires the item if it hasn't been accessed for
// "seconds" seconds
func (e *Expiration) AfterMinutesIdle(minutes int64) *Expiration {
	e.idle = e.afterIdle(minutes, time.Minute)
	return e
}

// AfterHoursIdle expires the item if it hasn't been accessed for 
// "hours" hours
func (e *Expiration) AfterHoursIdle(hours int64) *Expiration {
	e.idle = e.afterIdle(hours, time.Hour)
	return e
}

// AfterDaysIdle expires the item if it hasn't been accessed for
// "days" days
func (e *Expiration) AfterDaysIdle(days int64) *Expiration {
	e.idle = e.afterIdle(days, time.Hour*24)
	return e
}

// OnDate expires the item once "date" date has passed
func (e *Expiration) OnDate(date time.Time) *Expiration {
	e.absolute = date
	return e
}

// OnCondition expires the item if the "condition" func returns true
//
// This condition is checked, once per second, alongside the other expirations.
// Do not put any expensive code inside this condition. This check must be
// as quick as possible.
func (e *Expiration) OnCondition(condition ExpirationFunc) *Expiration {
	e.condition = condition
	return e
}
