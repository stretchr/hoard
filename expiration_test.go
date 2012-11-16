package hoard

import (
	"github.com/stretchrcom/testify/assert"
	"testing"
	"time"
)

func TestExpires(t *testing.T) {

	e := Expires()
	assert.NotNil(t, e)

}

func TestComplexExpires(t *testing.T) {

	date := time.Now()
	condition := ExpirationFunc(func() bool { return true })
	e := Expires().AfterHoursIdle(4).OnDate(date).OnCondition(condition)

	assert.Equal(t, e.condition, condition)
	assert.Equal(t, e.absolute, date)
	assert.Equal(t, e.idle.Hours(), 4)

}

func TestAfterSeconds(t *testing.T) {

	e := Expires().AfterSeconds(2)
	assert.NotNil(t, e)
	assert.False(t, e.absolute.IsZero())

}

func TestAfterMinutes(t *testing.T) {

	e := Expires().AfterMinutes(2)
	assert.NotNil(t, e)
	assert.False(t, e.absolute.IsZero())

}

func TestAfterHours(t *testing.T) {

	e := Expires().AfterHours(2)
	assert.NotNil(t, e)
	assert.False(t, e.absolute.IsZero())

}

func TestAfterDays(t *testing.T) {

	e := Expires().AfterDays(2)
	assert.NotNil(t, e)
	assert.False(t, e.absolute.IsZero())

}

func TestAfterDuration(t *testing.T) {

	e := Expires().AfterDuration(1)
	assert.NotNil(t, e)
	assert.False(t, e.absolute.IsZero())

}

func TestAfterSecondsIdle(t *testing.T) {

	e := Expires().AfterSecondsIdle(2)
	assert.NotNil(t, e)
	assert.Equal(t, 2, e.idle.Seconds())

}

func TestAfterMinutesIdle(t *testing.T) {

	e := Expires().AfterMinutesIdle(2)
	assert.NotNil(t, e)
	assert.Equal(t, 2, e.idle.Minutes())

}

func TestAfterDaysIdle(t *testing.T) {

	e := Expires().AfterDaysIdle(2)
	assert.NotNil(t, e)
	assert.Equal(t, 2*24, e.idle.Hours())

}

func TestAfterHoursIdle(t *testing.T) {

	e := Expires().AfterHoursIdle(2)
	assert.NotNil(t, e)
	assert.Equal(t, 2, e.idle.Hours())

}

func TestAfterDurationIdle(t *testing.T) {

	e := Expires().AfterDurationIdle(2 * time.Hour)
	assert.NotNil(t, e)
	assert.Equal(t, 2, e.idle.Hours())

}

func TestOnDate(t *testing.T) {

	date := time.Now()
	e := Expires().OnDate(date)
	assert.NotNil(t, e)
	assert.Equal(t, date, e.absolute)

}

func TestOnCondition(t *testing.T) {

	condition := ExpirationFunc(func() bool { return true })
	e := Expires().OnCondition(condition)
	assert.NotNil(t, e)
	assert.Equal(t, condition, e.condition)

}
