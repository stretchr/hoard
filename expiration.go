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

// ExpiresAfterSeconds expires the item after "seconds" seconds have passed
func (e *Expiration) ExpiresAfterSeconds(seconds int64) {
	e.absolute = time.Now().Add(time.Duration(seconds) * time.Second)
}

// ExpiresAfterHours expires the item after "hours" hours have passed
func (e *Expiration) ExpiresAfterHours(hours int64) {
	e.absolute = time.Now().Add(time.Duration(hours) * time.Hour)
}

// ExpiresAfterDays expires the item after "days" days have passed
func (e *Expiration) ExpiresAfterDays(days int64) {
	e.absolute = time.Now().Add(time.Duration(days) * time.Hour * 24)
}

// ExpiresAfterSecondsIdle expires the item if it hasn't been accessed for
// "seconds" seconds
func (e *Expiration) ExpiresAfterSecondsIdle(seconds int64) {
	e.idle = time.Duration(seconds) * time.Second
}

// ExpiresAfterHoursIdle expires the item if it hasn't been accessed for 
// "hours" hours
func (e *Expiration) ExpiresAfterHoursIdle(hours int64) {
	e.idle = time.Duration(hours) * time.Hour
}

// ExpiresAfterDaysIdle expires the item if it hasn't been accessed for
// "days" days
func (e *Expiration) ExpiresAfterDaysIdle(days int64) {
	e.idle = time.Duration(days) * time.Hour * 24
}

// ExpiresOnDate expires the item once "date" date has passed
func (e *Expiration) ExpiresOnDate(date time.Time) {
	e.absolute = date
}

// ExpiresOnCondition expires the item if the "condition" func returns true
func (e *Expiration) ExpiresOnCondition(condition ExpirationFunc) {
	e.condition = condition
}
