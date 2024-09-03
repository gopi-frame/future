package future

import (
	"errors"
	"testing"
	"time"

	"github.com/gopi-frame/exception"
	"github.com/stretchr/testify/assert"
)

func TestAsync(t *testing.T) {
	start := time.Now()
	Async(func() int {
		time.Sleep(100 * time.Second)
		return 1
	})
	assert.Less(t, time.Since(start), 100*time.Second)
}

func TestVoid(t *testing.T) {
	start := time.Now()
	Void(func() {
		time.Sleep(100 * time.Second)
	})
	assert.Less(t, time.Since(start), 100*time.Second)
}

func TestValue(t *testing.T) {
	result := Value(1).Await()
	assert.EqualValues(t, 1, result)
}

func TestTimeout(t *testing.T) {
	t.Run("Timeout", func(t *testing.T) {
		assert.PanicsWithError(t, exception.NewTimeoutException().Error(), func() {
			Timeout(func() int {
				time.Sleep(100 * time.Second)
				return 1
			}, time.Second).Await()
		})
	})

	t.Run("Not timeout", func(t *testing.T) {
		assert.NotPanics(t, func() {
			result := Timeout(func() int {
				time.Sleep(100 * time.Millisecond)
				return 1
			}, time.Second).Await()
			assert.Equal(t, 1, result)
		})
	})
}

func TestDelay(t *testing.T) {
	start := time.Now()
	Delay(func() int { return 0 }, time.Second).Await()
	assert.GreaterOrEqual(t, time.Since(start), time.Second)
}

func TestForeach(t *testing.T) {
	var nums = []int{1, 2, 3, 4, 5}
	var result = []int{}
	Foreach(nums, func(num int) *Future[int] {
		return Async(func() int { result = append(result, num); return num })
	}).Await()
	assert.EqualValues(t, nums, result)
}

func TestWait(t *testing.T) {
	start := time.Now()
	Wait(
		Void(func() { time.Sleep(time.Millisecond * 100) }),
		Void(func() { time.Sleep(time.Millisecond * 200) }),
		Void(func() { time.Sleep(time.Millisecond * 300) }),
	)
	assert.GreaterOrEqual(t, time.Since(start), 300*time.Millisecond)
}

func TestFuture_Then(t *testing.T) {
	result := Async(func() int {
		return 1
	}).Then(func(result int) int {
		return result + 1
	}, nil).Then(func(result int) int {
		return result + 1
	}, nil).Await()
	assert.Equal(t, 3, result)
}

func TestFuture_Catch(t *testing.T) {
	t.Run("Exception on Async", func(t *testing.T) {
		assert.Panics(t, func() {
			Void(func() {
				panic(errors.New("exception on async"))
			}).Then(func(value any) any {
				return nil
			}, nil).Catch(exception.NewTimeoutException(), func(err error) {
				assert.EqualError(t, err, "exception on async")
			}).Await()
		})
	})

	t.Run("Exception on Then", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Void(func() {
			}).Then(func(value any) any {
				panic(errors.New("exception on then"))
			}, nil).Catch(errors.New("exception on then"), func(err error) {
				assert.EqualError(t, err, "exception on then")
			}).Await()
		})
	})
}

func TestFuture_Complete(t *testing.T) {
	t.Run("no exception", func(t *testing.T) {
		var i int
		Async(func() int {
			return 100
		}).Complete(func() {
			i = 1000
		}).Await()
		assert.Equal(t, 1000, i)
	})

	t.Run("has exception", func(t *testing.T) {
		var i int
		Async(func() int {
			panic("exception")
		}).Complete(func() {
			i = 1000
		}).CatchAll(func(err error) {

		}).Await()
		assert.Equal(t, 1000, i)
	})
}
