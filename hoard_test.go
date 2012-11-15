package hoard

import (
	"errors"
	"github.com/stretchrcom/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestHoard_MakeHoard(t *testing.T) {

	h := MakeHoard(ExpiresNever)

	if assert.NotNil(t, h) {
		assert.Nil(t, h.defaultExpiration)
	}

	e := new(Expiration)
	e.AfterSeconds(1)

	h = MakeHoard(e)
	if assert.NotNil(t, h) {
		assert.Condition(t, func() bool {
			return !h.defaultExpiration.absolute.IsZero()
		})
	}

}

func TestHoard_SharedHoard(t *testing.T) {

	h := SharedHoard()
	assert.NotNil(t, h)

	h2 := SharedHoard()
	assert.Equal(t, h, h2)

}

func TestHoard_Get(t *testing.T) {

	firstCalled := false
	secondCalled := false
	h := MakeHoard(ExpiresNever)

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

	h := MakeHoard(ExpiresNever)

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

func TestHoard_Expire(t *testing.T) {

	h := MakeHoard(ExpiresNever)

	h.Get("something", func() (interface{}, *Expiration) {
		return 1, nil
	})

	assert.Equal(t, 1, h.Get("something"))

	h.Expire("something")
	assert.Equal(t, 2, h.Get("something", func() (interface{}, *Expiration) { return 2, nil }))

}

func TestHoard_SetExpires(t *testing.T) {

	date := time.Now()

	h := MakeHoard(ExpiresNever)
	h.Get("key", func() (interface{}, *Expiration) {
		return "first", ExpiresNever
	})

	assert.Equal(t, ExpiresNever, h.cache["key"].expiration)

	h.SetExpires("key", Expires().OnDate(date))

	item, _ := h.cacheGet("key")
	if assert.NotNil(t, &item) {
		if assert.NotNil(t, item.expiration, "Expiration should be set") {
			assert.Equal(t, date, item.expiration.absolute)
		}
	}

}

func TestHoard_SetExpires_Panics(t *testing.T) {

	h := MakeHoard(ExpiresNever)
	assert.False(t, h.SetExpires("key", Expires().OnDate(time.Now())))

}

func TestHoard_ExpirationSetting(t *testing.T) {

	h := MakeHoard(ExpiresNever)

	result := h.Get("key2", func() (interface{}, *Expiration) {
		expiration := new(Expiration)
		expiration.AfterSecondsIdle(10)
		expiration.AfterSeconds(10)
		expiration.OnCondition(func() bool {
			return true
		})
		return "second", expiration
	})

	assert.Equal(t, result, "second")
	assert.NotEqual(t, 0, h.cache["key2"].expiration.idle)
	assert.Condition(t, func() bool {
		return !h.cache["key2"].expiration.absolute.IsZero()
	})
	assert.Condition(t, func() bool {
		return h.cache["key2"].expiration.condition != nil
	})

}

func TestHoard_Set(t *testing.T) {

	h := MakeHoard(ExpiresNever)

	h.Set("key", 1)

	assert.Equal(t, 1, h.Get("key"))

}

func TestHoard_Has(t *testing.T) {
	h := MakeHoard(ExpiresNever)

	_ = h.Get("key", func() (interface{}, *Expiration) {
		return "first", ExpiresNever
	})

	assert.True(t, h.Has("key"))
}

// The below functions take forever to run as they wait for expirations to tick
// They are commented out to speed up development
/*
func TestHoard_TickerStartStop(t *testing.T) {

	h := SharedHoard()

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

	h := MakeHoard(ExpiresNever)

	result := h.Get("key3", func() (interface{}, *Expiration) {
		expiration := new(Expiration)
		expiration.ExpiresAfterSecondsIdle(2)
		return "first", expiration
	})
	assert.Equal(t, result, "first")
	time.Sleep(1 * time.Second)

	// test the sliding window
	result = h.Get("key3", func() (interface{}, *Expiration) {
		return "second", ExpiresNever
	})

	assert.Equal(t, result, "first")

	time.Sleep(2 * time.Second)

	result = h.Get("key3", func() (interface{}, *Expiration) {
		return "second", ExpiresNever
	})

	assert.Equal(t, result, "second")

}

func TestHoard_AbsoluteExpiration(t *testing.T) {

	h := MakeHoard(ExpiresNever)

	result := h.Get("key4", func() (interface{}, *Expiration) {
		expiration := new(Expiration)
		expiration.ExpiresAfterSeconds(1)
		return "first", expiration
	})
	assert.Equal(t, result, "first")
	time.Sleep(2 * time.Second)

	result = h.Get("key4", func() (interface{}, *Expiration) {
		return "second", ExpiresNever
	})

	assert.Equal(t, result, "second")

}

func TestHoard_ConditionalExpiration(t *testing.T) {

	h := MakeHoard(ExpiresNever)

	result := h.Get("key5", func() (interface{}, *Expiration) {
		expiration := new(Expiration)
		expiration.ExpiresOnCondition(func() bool {
			return true
		})
		return "first", expiration
	})

	assert.Equal(t, result, "first")

	time.Sleep(2 * time.Second)

	result = h.Get("key5", func() (interface{}, *Expiration) {
		return "second", ExpiresNever
	})

	assert.Equal(t, result, "second")
}*/

/*
// Tests thread safety between adding items and flushing items
// This is an expensive, slow test, so it is commented out for now
func TestHoard_ThreadSafety(t *testing.T) {

	h := MakeHoard(ExpiresNever)

	iterations := 100000

	var wait sync.WaitGroup
	wait.Add(4)

	go func() {
		for i := 0; i < iterations; i++ {
			_ = h.Get(fmt.Sprintf("stresstest%d", i), func() (interface{}, *Expiration) {
				expiration := new(Expiration)
				expiration.ExpiresAfterSeconds(int64(rand.Int() % 10))
				return "first", expiration
			})
		}
		wait.Done()
	}()

	go func() {
		for i := 0; i < iterations; i++ {
			_ = h.Get(fmt.Sprintf("stresstest-1-%d", i), func() (interface{}, *Expiration) {
				expiration := new(Expiration)
				expiration.ExpiresAfterSeconds(int64(rand.Int() % 10))
				return "first", expiration
			})
		}
		wait.Done()
	}()

	go func() {
		for i := 0; i < iterations; i++ {
			_ = h.Get(fmt.Sprintf("stresstest-2-%d", i), func() (interface{}, *Expiration) {
				expiration := new(Expiration)
				expiration.ExpiresAfterSeconds(int64(rand.Int() % 10))
				return "first", expiration
			})
		}
		wait.Done()
	}()

	go func() {
		for i := 0; i < iterations; i++ {
			_ = h.Get(fmt.Sprintf("stresstest-3-%d", i), func() (interface{}, *Expiration) {
				expiration := new(Expiration)
				expiration.ExpiresAfterSeconds(int64(rand.Int() % 10))
				return "first", expiration
			})
		}
		wait.Done()
	}()

	wait.Wait()

	for len(h.cache) > 0 {
		time.Sleep(1 * time.Second)
	}

}
*/

func BenchmarkHoard_AddingExpiring(b *testing.B) {

	b.StopTimer()

	h := SharedHoard()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = h.Get(string(i), func() (interface{}, *Expiration) {
			expiration := new(Expiration)
			expiration.AfterSeconds(int64(rand.Int() % 2))
			return 1, ExpiresNever
		})
	}

}
