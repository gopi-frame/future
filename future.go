package future

import (
	"reflect"

	"github.com/gopi-frame/utils/catch"
)

// Future future
type Future[T any] struct {
	fn        func()
	value     T
	err       error
	completed chan struct{}
}

func (f *Future[T]) Await() T {
	<-f.completed
	if f.err != nil {
		panic(f.err)
	}
	return f.value
}

func (f *Future[T]) Then(onValue func(value T) T, onError func(err error)) *Future[T] {
	future := newFuture[T]()
	future.fn = func() {
		go catch.Try(func() {
			<-f.completed
			if f.err != nil {
				if onError != nil {
					onError(f.err)
				} else {
					panic(f.err)
				}
			} else if f.err == nil && onValue != nil {
				future.value = onValue(f.value)
			}
		}).CatchAll(func(err error) {
			future.err = err
		}).Finally(func() {
			close(future.completed)
		}).Run()
	}
	future.fn()
	return future
}

func (f *Future[T]) Catch(err error, handler func(err error)) *Future[T] {
	future := newFuture[T]()
	future.fn = func() {
		go catch.Try(func() {
			<-f.completed
			if f.err != nil && handler != nil {
				if reflect.TypeOf(err) == reflect.TypeOf(f.err) {
					handler(f.err)
				} else {
					panic(f.err)
				}
			} else {
				panic(f.err)
			}
		}).CatchAll(func(err error) {
			future.err = err
		}).Finally(func() {
			close(future.completed)
		}).Run()
	}
	future.fn()
	return future
}

func (f *Future[T]) CatchAll(handler func(err error)) *Future[T] {
	future := newFuture[T]()
	future.fn = func() {
		go catch.Try(func() {
			<-f.completed
			future.value = f.value
			if f.err != nil && handler != nil {
				handler(f.err)
			} else {
				panic(f.err)
			}
		}).CatchAll(func(err error) {
			future.err = err
		}).Finally(func() {
			close(future.completed)
		}).Run()
	}
	future.fn()
	return future
}

func (f *Future[T]) Complete(handler func()) *Future[T] {
	future := newFuture[T]()
	future.fn = func() {
		go catch.Try(func() {
			<-f.completed
			future.value = f.value
			future.err = f.err
			handler()
		}).CatchAll(func(err error) {
			future.err = err
		}).Finally(func() {
			close(future.completed)
		}).Run()
	}
	future.fn()
	return future
}
