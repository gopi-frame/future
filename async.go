package future

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gopi-frame/exception"
	"github.com/gopi-frame/support/lists"
	"github.com/gopi-frame/utils/catch"
)

func newFuture[T any]() *Future[T] {
	return &Future[T]{
		completed: make(chan struct{}),
	}
}

// Void future without return
func Void(callback func()) *Future[any] {
	return Async(func() any {
		callback()
		return nil
	})
}

// Async async
func Async[T any](callback func() T) *Future[T] {
	future := newFuture[T]()
	future.fn = func() {
		go catch.Try(func() {
			future.value = callback()
		}).CatchAll(func(err error) {
			future.err = err
		}).Finally(func() {
			close(future.completed)
		}).Run()
	}
	future.fn()
	return future
}

// Value value
func Value[T any](value T) *Future[T] {
	return Async(func() T {
		return value
	})
}

// Timeout with timeout
func Timeout[T any](callback func() T, timeout time.Duration) *Future[T] {
	future := newFuture[T]()
	future.fn = func() {
		ctx, cancel := context.WithTimeoutCause(context.Background(), timeout,
			exception.NewTimeoutException())
		defer cancel()
		done := make(chan struct{}, 1)
		go catch.Try(func() {
			future.value = callback()
		}).CatchAll(func(err error) {
			future.err = err
		}).Finally(func() {
			close(done)
		}).Run()
		select {
		case <-ctx.Done():
			future.err = context.Cause(ctx)
			close(future.completed)
			return
		case <-done:
			close(future.completed)
			return
		}
	}
	future.fn()
	return future
}

// Delay delay
func Delay[T any](callback func() T, delay time.Duration) *Future[T] {
	future := newFuture[T]()
	future.fn = func() {
		go catch.Try(func() {
			time.Sleep(delay)
			future.value = callback()
		}).CatchAll(func(err error) {
			future.err = err
		}).Finally(func() {
			close(future.completed)
		}).Run()
	}
	future.fn()
	return future
}

// Foreach foreach
func Foreach[T any, R any](elements []T, callback func(element T) *Future[R]) *Future[any] {
	wg := sync.WaitGroup{}
	wg.Add(len(elements))
	future := Void(func() {
		for _, element := range elements {
			callback(element).Await()
		}
	})
	return future
}

// Wait wait
func Wait[T any](futures ...*Future[T]) *Future[*lists.List[T]] {
	values := lists.NewList(make([]T, len(futures))...)
	errs := []error{}
	future := newFuture[*lists.List[T]]()
	future.fn = func() {
		wg := sync.WaitGroup{}
		wg.Add(len(futures))
		for index, f := range futures {
			go func(index int, f *Future[T]) {
				values.Lock()
				catch.Try(func() {
					<-f.completed
					values.Set(index, f.value)
				}).CatchAll(func(err error) {
					f.err = err
					errs = append(errs, f.err)
				}).Finally(func() {
					wg.Done()
					values.Unlock()
				}).Run()
			}(index, f)
		}
		wg.Wait()
		future.err = errors.Join(errs...)
		future.value = values
		close(future.completed)
	}
	future.fn()
	return future
}
