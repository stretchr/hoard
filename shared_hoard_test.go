package hoard

import (
	"errors"
	"github.com/stretchrcom/testify/assert"
	"testing"
	"time"
)

func TestHoard_SharedHoard(t *testing.T) {

	h := SharedHoard()
	assert.NotNil(t, h)

	h2 := SharedHoard()
	assert.Equal(t, h, h2)

}

func TestSharedHoard_Get(t *testing.T) {

	firstCalled := false
	secondCalled := false

	result := Get("key", func() (interface{}, *Expiration) {
		firstCalled = true
		return "first", ExpiresNever
	})

	assert.Equal(t, result, "first")
	assert.True(t, firstCalled)

	result = Get("key", func() (interface{}, *Expiration) {
		secondCalled = true
		return "second", ExpiresNever
	})

	assert.NotEqual(t, result, "second")
	assert.False(t, secondCalled)

}

func TestSharedHoard_GetWithError(t *testing.T) {

	result, err := GetWithError("key", func() (interface{}, error, *Expiration) {
		return "first", nil, ExpiresNever
	})

	assert.Equal(t, result, "first")
	assert.Nil(t, err)

	result, err = GetWithError("key2", func() (interface{}, error, *Expiration) {
		return "second", errors.New("EXTERMINATE!!!"), ExpiresNever
	})

	assert.Equal(t, "second", result)
	assert.NotNil(t, err)

}

func TestSharedHoard_Remove(t *testing.T) {

	Get("something", func() (interface{}, *Expiration) {
		return 1, nil
	})

	assert.Equal(t, 1, Get("something"))

	Remove("something")
	assert.Equal(t, 2, Get("something", func() (interface{}, *Expiration) { return 2, nil }))

}

func TestSharedHoard_SetExpires(t *testing.T) {

	date := time.Now()

	Get("key", func() (interface{}, *Expiration) {
		return "first", ExpiresNever
	})

	assert.Equal(t, ExpiresNever, SharedHoard().cache["key"].expiration)

	SetExpires("key", Expires().OnDate(date))

	item, _ := SharedHoard().cacheGet("key")
	if assert.NotNil(t, &item) {
		if assert.NotNil(t, item.expiration, "Expiration should be set") {
			assert.Equal(t, date, item.expiration.absolute)
		}
	}

}

func TestSharedHoard_SetExpires_Panics(t *testing.T) {

	sharedHoard = MakeHoard(nil)
	assert.False(t, SetExpires("key", Expires().OnDate(time.Now())))

}

func TestSharedHoard_ExpirationSetting(t *testing.T) {

	result := Get("key2", func() (interface{}, *Expiration) {
		expiration := new(Expiration)
		expiration.AfterSecondsIdle(10)
		expiration.AfterSeconds(10)
		expiration.OnCondition(func() bool {
			return true
		})
		return "second", expiration
	})

	assert.Equal(t, result, "second")
	assert.NotEqual(t, 0, SharedHoard().cache["key2"].expiration.idle)
	assert.Condition(t, func() bool {
		return !SharedHoard().cache["key2"].expiration.absolute.IsZero()
	})
	assert.Condition(t, func() bool {
		return SharedHoard().cache["key2"].expiration.condition != nil
	})

}

func TestSharedHoard_Set(t *testing.T) {

	Set("key", 1)

	assert.Equal(t, 1, SharedHoard().Get("key"))
	assert.Equal(t, 1, Get("key"))

}

func TestSharedHoard_Has(t *testing.T) {

	_ = Get("key", func() (interface{}, *Expiration) {
		return "first", ExpiresNever
	})

	assert.True(t, Has("key"))
}
