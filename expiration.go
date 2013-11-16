package hoard

import (
	"time"
)

// ExpirationCondition is a type for the function signature of a custom expiration
// function.
type ExpirationCondition func() bool

// ExpiresNever is an Expiration that indicates the object never expires.
var ExpiresNever *Expiration = &Expiration{}

// ExpiresDefault is used to indicate that the cache system should use the
// default expiration policy.
var ExpiresDefault *Expiration = nil

// Expiration describes when an object will expire.
type Expiration struct {
	// idle is the sliding window duration for expiration.
	idle time.Duration

	// duration is the expiration time to pass after the object has been added to the cache
	duration time.Duration

	// date is an specific point in time to expire at
	date time.Time

	// absolute is the absolute point in time, used for fast comparasion.
	// its the earliest resulting time from idle, duration or date.
	absolute time.Time

	// condition is a function provided by the creator which is called to
	// determine if an object is expired.
	condition ExpirationCondition
}

// Expires creates a new empty Expiration object.
//
// Example
//
//     hoard.Expires().AfterSeconds(2)
func Expires() *Expiration {
	return new(Expiration)
}

// updateAbsoluteTime sets the internal absolute field to the earliest point
// in time resulting from idle, duration or date.
func (e *Expiration) updateAbsoluteTime(lastAccess, created time.Time) *Expiration {
	abs := e.date
	if e.idle != 0 {
		if t := lastAccess.Add(e.idle); t.Before(abs) || abs.IsZero() {
			abs = t
		}
	}
	if e.duration != 0 {
		if t := created.Add(e.duration); t.Before(abs) || abs.IsZero() {
			abs = t
		}
	}
	e.absolute = abs
	return e
}

// isExpiredAbsolute is a quicker variant of IsExpired, but only works if the internal absolute time has been set,
// i.e. updateAbsoluteTime() must have been called on it before
func (e *Expiration) isExpiredAbsolute(currentTime time.Time) bool {

	if !e.absolute.IsZero() && currentTime.After(e.absolute) {
		return true
	}
	if e.condition != nil && e.condition() {
		return true
	}
	return false
}

// IsExpired determines if an expiration object has expired due to the
// lastAccess time, the creation time, an absolute point in time or an expiration condition.
func (e *Expiration) IsExpired(lastAccess, created time.Time) bool {
	currentTime := time.Now()

	if e.duration != 0 && currentTime.Sub(created) > e.duration {
		return true
	}
	if e.idle != 0 && currentTime.Sub(lastAccess) > e.idle {
		return true
	}
	if !e.date.IsZero() && currentTime.After(e.date) {
		return true
	}
	if e.condition != nil && e.condition() {
		return true
	}
	return false
}

// IsExpiredByTime determines if an expiration object has expired due to the
// lastAccess time and the current time.
//
// Depracted, only there for downward compatibility reasons, because it fails to deal with pure durations
// set with AfterSeconds(), AfterMinutes() etc.
// Use IsExpired instead
func (e *Expiration) IsExpiredByTime(lastAccess, currentTime time.Time) bool {
	if e.idle != 0 {
		if currentTime.Sub(lastAccess) >
			e.idle {
			return true
		}
	}
	if !e.date.IsZero() {
		if currentTime.After(e.date) {
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

// IsExpiredByCondition determines if an expiration object has expired by calling
// the provided ExpirationCondition function.
func (e *Expiration) IsExpiredByCondition() bool {
	if e.condition != nil {
		if e.condition() {
			return true
		}
	}
	return false
}

// after creates a time.Time from the duration and multiplier provided.
func (e *Expiration) after(duration int64, multiplier time.Duration) time.Duration {
	return time.Duration(duration) * multiplier
}

// AfterSeconds expires the item after "seconds" seconds have passed.
func (e *Expiration) AfterSeconds(seconds int64) *Expiration {
	e.duration = e.after(seconds, time.Second)
	return e
}

// AfterMinutes expires the item after "minutes" minutes have passed.
func (e *Expiration) AfterMinutes(minutes int64) *Expiration {
	e.duration = e.after(minutes, time.Minute)
	return e
}

// AfterHours expires the item after "hours" hours have passed.
func (e *Expiration) AfterHours(hours int64) *Expiration {
	e.duration = e.after(hours, time.Hour)
	return e
}

// AfterDays expires the item after "days" days have passed.
func (e *Expiration) AfterDays(days int64) *Expiration {
	e.duration = e.after(days, time.Hour*24)
	return e
}

// AfterDuration expires the item after "duration" duration has passed.
func (e *Expiration) AfterDuration(duration time.Duration) *Expiration {
	e.duration = duration
	return e
}

// AfterSecondsIdle expires the item if it hasn't been accessed for
// "seconds" seconds.
func (e *Expiration) AfterSecondsIdle(seconds int64) *Expiration {
	e.idle = e.after(seconds, time.Second)
	return e
}

// AfterMinutesIdle expires the item if it hasn't been accessed for
// "minutes" minutes.
func (e *Expiration) AfterMinutesIdle(minutes int64) *Expiration {
	e.idle = e.after(minutes, time.Minute)
	return e
}

// AfterHoursIdle expires the item if it hasn't been accessed for
// "hours" hours.
func (e *Expiration) AfterHoursIdle(hours int64) *Expiration {
	e.idle = e.after(hours, time.Hour)
	return e
}

// AfterDaysIdle expires the item if it hasn't been accessed for
// "days" days.
func (e *Expiration) AfterDaysIdle(days int64) *Expiration {
	e.idle = e.after(days, time.Hour*24)
	return e
}

// AfterDurationIdle expires the item if it hasn't been accessed for
// "duration" duration.
func (e *Expiration) AfterDurationIdle(duration time.Duration) *Expiration {
	e.idle = duration
	return e
}

// OnDate expires the item once "date" date has passed.
func (e *Expiration) OnDate(date time.Time) *Expiration {
	e.date = date
	return e
}

// OnCondition expires the item if the "condition" func returns true.
//
// This condition is checked before retrieving the item from cache. If the
// condition returns true, the item is deleted and a new item will be fetched
// from the DataGetter.
func (e *Expiration) OnCondition(condition ExpirationCondition) *Expiration {
	e.condition = condition
	return e
}
