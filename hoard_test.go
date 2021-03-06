package hoard

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestHoard_Make(t *testing.T) {

	h := Make(ExpiresNever)

	if assert.NotNil(t, h) {
		assert.Equal(t, h.defaultExpiration, ExpiresNever)
	}

	h = Make(Expires().AfterSeconds(3))
	if assert.NotNil(t, h) {
		assert.Condition(t, func() bool {
			return h.defaultExpiration.duration == 3*time.Second
		})
	}

}

func TestHoard_Get(t *testing.T) {

	firstCalled := false
	secondCalled := false
	h := Make(ExpiresNever)

	result := h.Get("key", func() (interface{}, *Expiration) {
		firstCalled = true
		return "first", ExpiresNever
	})

	assert.Equal(t, result, "first")
	assert.True(t, firstCalled)

	result = h.Get("key", func() (interface{}, *Expiration) {
		secondCalled = true
		return "second", ExpiresNever
	})

	assert.NotEqual(t, result, "second")
	assert.False(t, secondCalled)

}

func TestHoard_GetWithError(t *testing.T) {

	h := Make(ExpiresNever)

	result, err := h.GetWithError("key", func() (interface{}, error, *Expiration) {
		return "first", nil, ExpiresNever
	})

	assert.Equal(t, result, "first")
	assert.Nil(t, err)

	result, err = h.GetWithError("key2", func() (interface{}, error, *Expiration) {
		return "second", errors.New("EXTERMINATE!!!"), ExpiresNever
	})

	assert.Equal(t, "second", result)
	assert.NotNil(t, err)

}

func TestHoard_Remove(t *testing.T) {

	h := Make(ExpiresNever)

	h.Get("something", func() (interface{}, *Expiration) {
		return 1, nil
	})

	assert.Equal(t, 1, h.Get("something"))

	h.Remove("something")
	assert.Equal(t, 2, h.Get("something", func() (interface{}, *Expiration) { return 2, nil }))

}

func TestHoard_SetExpires(t *testing.T) {

	date := time.Now()

	h := Make(ExpiresNever)
	h.Get("key", func() (interface{}, *Expiration) {
		return "first", ExpiresNever
	})

	assert.Equal(t, ExpiresNever, h.cache["key"].expiration)

	h.SetExpires("key", Expires().OnDate(date))

	item, _ := h.cacheGet("key")
	if assert.NotNil(t, &item) {
		if assert.NotNil(t, item.expiration, "Expiration should be set") {
			assert.Equal(t, date, item.expiration.date)
		}
	}

	// the expiratoin cache item should have its absolute expiration set to the date value as well
	expirationItem := h.expirationCache["key"]
	if assert.NotNil(t, &expirationItem) {
		if assert.NotNil(t, expirationItem.expiration, "Expiration should be set") {
			assert.Equal(t, date, expirationItem.expiration.absolute)
		}
	}

}

func TestHoard_SetExpires_Panics(t *testing.T) {

	h := Make(ExpiresNever)
	assert.False(t, h.SetExpires("key", Expires().OnDate(time.Now())))

}

func TestHoard_ExpirationSetting(t *testing.T) {

	h := Make(Expires().AfterSeconds(1))

	result := h.Get("key2", func() (interface{}, *Expiration) {
		expiration := Expires().AfterSecondsIdle(10).AfterSeconds(10).OnCondition(func() bool {
			return true
		})
		return "second", expiration
	})

	assert.Equal(t, result, "second")
	assert.NotEqual(t, 0, h.cache["key2"].expiration.idle)
	assert.NotEqual(t, 0, h.cache["key2"].expiration.duration)
	assert.Condition(t, func() bool {
		return h.cache["key2"].expiration.condition != nil
	})
	assert.Condition(t, func() bool {
		return !h.cache["key2"].expiration.absolute.IsZero()
	})

}

func TestHoard_ConditionalExpiration(t *testing.T) {

	h := Make(Expires().AfterSeconds(1))

	result := h.Get("key", func() (interface{}, *Expiration) {
		expiration := Expires().OnCondition(func() bool {
			return true
		})
		return "first", expiration
	})

	assert.Equal(t, result, "first")

	result = h.Get("key", func() (interface{}, *Expiration) {
		expiration := Expires().OnCondition(func() bool {
			return true
		})
		return "second", expiration
	})

	assert.Equal(t, result, "second")
}

func TestHoard_Set(t *testing.T) {

	h := Make(ExpiresNever)

	h.Set("key", 1)

	assert.Equal(t, 1, h.Get("key"))

}

func TestHoard_Has(t *testing.T) {
	h := Make(ExpiresNever)

	_ = h.Get("key", func() (interface{}, *Expiration) {
		return "first", ExpiresNever
	})

	assert.True(t, h.Has("key"))
}

func TestHoard_OverrideDefault(t *testing.T) {

	h := Make(Expires().AfterSeconds(1))

	_ = h.Get("key", func() (interface{}, *Expiration) {
		return "first", ExpiresNever
	})

	assert.Equal(t, ExpiresNever, h.cache["key"].expiration)

	h = Make(ExpiresNever)

	_ = h.Get("key", func() (interface{}, *Expiration) {
		return "first", Expires().AfterSecondsIdle(1)
	})

	assert.Equal(t, 1, h.cache["key"].expiration.idle.Seconds())

}

func TestHoard_UseDefault(t *testing.T) {

	h := Make(Expires().AfterSecondsIdle(1))

	_ = h.Get("key", func() (interface{}, *Expiration) {
		return "first", ExpiresDefault
	})

	assert.Equal(t, 1, h.cache["key"].expiration.idle.Seconds())

	h = Make(ExpiresNever)

	_ = h.Get("key", func() (interface{}, *Expiration) {
		return "first", ExpiresDefault
	})

	assert.Equal(t, ExpiresNever, h.cache["key"].expiration)

}

var multiThread = Make(ExpiresDefault)
var cachedInt = 0
var cachedIntError = 0

func TestHoard_GetSafety(t *testing.T) {

	// We should be able to call the special get function from multiple threads
	// and have it only execute once

	go GetInt(t)
	go GetInt(t)
	go GetInt(t)
	go GetInt(t)
	GetInt(t)
	time.Sleep(time.Millisecond * 10)

}

func GetInt(t *testing.T) {

	retrievedInt := multiThread.Get("MultiThreadGetInt", func() (interface{}, *Expiration) {
		cachedInt++
		time.Sleep(10 * time.Millisecond)
		return cachedInt, ExpiresNever
	}).(int)

	assert.Equal(t, retrievedInt, 1)

	retrievedIntError, _ := multiThread.GetWithError("MultiThreadGetIntError", func() (interface{}, error, *Expiration) {
		cachedIntError++
		time.Sleep(10 * time.Millisecond)
		return cachedInt, nil, ExpiresNever
	})

	assert.Equal(t, retrievedIntError.(int), 1)

}

func TestHoard_ReEntry(t *testing.T) {

	h := Make(ExpiresDefault)

	result := h.Get("reone", func() (interface{}, *Expiration) {
		return h.Get("retwo", func() (interface{}, *Expiration) {
			return "retwo", ExpiresNever
		}), ExpiresNever
	}).(string)

	assert.Equal(t, result, "retwo")

}

// The below functions take forever to run as they wait for expirations to tick
// They are commented out to speed up development. They should be run before any
// commit to ensure they still pass.
/*
func TestHoard_TickerStartStop(t *testing.T) {

	h := Make(ExpiresNever)

	_ = h.Get("key", func() (interface{}, *Expiration) {
		return "first", ExpiresNever
	})

	assert.False(t, h.getTickerRunning())
	assert.Equal(t, 1, len(h.cache))

	_ = h.Get("key2", func() (interface{}, *Expiration) {
		return "first", Expires().AfterSeconds(1)
	})

	assert.True(t, h.getTickerRunning())
	assert.Equal(t, 2, len(h.cache))

	time.Sleep(3 * time.Second)

	assert.False(t, h.getTickerRunning())

	_ = h.Get("key3", func() (interface{}, *Expiration) {
		return "first", Expires().AfterSeconds(1)
	})
	_ = h.Get("key4", func() (interface{}, *Expiration) {
		return "first", Expires().AfterSeconds(2)
	})

	assert.True(t, h.getTickerRunning())
	assert.Equal(t, 3, len(h.cache))

	time.Sleep(4 * time.Second)

	assert.False(t, h.getTickerRunning())
	assert.Equal(t, 1, len(h.cache))

}

func TestHoard_IdleExpiration(t *testing.T) {

	h := Make(ExpiresNever)

	result := h.Get("key3", func() (interface{}, *Expiration) {
		return "first", Expires().AfterSecondsIdle(2)
	})
	assert.Equal(t, result, "first")
	time.Sleep(1 * time.Second)

	// test the sliding window
	result = h.Get("key3", func() (interface{}, *Expiration) {
		return "second", ExpiresNever
	})

	assert.Equal(t, result, "first")

	time.Sleep(3 * time.Second)

	result = h.Get("key3", func() (interface{}, *Expiration) {
		return "second", ExpiresNever
	})

	assert.Equal(t, result, "second")

}

func TestHoard_AbsoluteExpiration(t *testing.T) {

	h := Make(ExpiresNever)

	result := h.Get("key4", func() (interface{}, *Expiration) {
		return "first", Expires().AfterSeconds(1)
	})
	assert.Equal(t, result, "first")
	time.Sleep(2 * time.Second)

	result = h.Get("key4", func() (interface{}, *Expiration) {
		return "second", ExpiresNever
	})

	assert.Equal(t, result, "second")

}

func TestHoard_ConditionalExpiration(t *testing.T) {

	h := Make(ExpiresNever)

	result := h.Get("key5", func() (interface{}, *Expiration) {
		return "first", Expires().OnCondition(func() bool {
			return true
		})
	})

	assert.Equal(t, result, "first")

	time.Sleep(2 * time.Second)

	result = h.Get("key5", func() (interface{}, *Expiration) {
		return "second", ExpiresNever
	})

	assert.Equal(t, result, "second")
}

// Tests thread safety between adding items and flushing items
// This is an expensive, slow test, so it is commented out for now
/*
func TestHoard_ThreadSafety(t *testing.T) {

	h := Make(ExpiresNever)

	iterations := 100000

	var wait sync.WaitGroup
	wait.Add(4)

	go func() {
		for i := 0; i < iterations; i++ {
			_ = h.Get(fmt.Sprintf("stresstest%d", i), func() (interface{}, *Expiration) {
				return "first", Expires().AfterSeconds(int64(rand.Int() % 10))
			})
		}
		wait.Done()
	}()

	go func() {
		for i := 0; i < iterations; i++ {
			_ = h.Get(fmt.Sprintf("stresstest-1-%d", i), func() (interface{}, *Expiration) {
				return "first", Expires().AfterSeconds(int64(rand.Int() % 10))
			})
		}
		wait.Done()
	}()

	go func() {
		for i := 0; i < iterations; i++ {
			_ = h.Get(fmt.Sprintf("stresstest-2-%d", i), func() (interface{}, *Expiration) {
				return "first", Expires().AfterSeconds(int64(rand.Int() % 10))
			})
		}
		wait.Done()
	}()

	go func() {
		for i := 0; i < iterations; i++ {
			_ = h.Get(fmt.Sprintf("stresstest-3-%d", i), func() (interface{}, *Expiration) {
				return "first", Expires().AfterSeconds(int64(rand.Int() % 10))
			})
		}
		wait.Done()
	}()

	wait.Wait()

	for len(h.cache) > 0 {
		time.Sleep(1 * time.Second)
	}

}*/

func BenchmarkHoard_AddingExpiring(b *testing.B) {

	b.StopTimer()

	h := Make(ExpiresNever)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = h.Get(string(i), func() (interface{}, *Expiration) {
			return 1, Expires().AfterSeconds(int64(rand.Int() % 2))
		})
	}

}
