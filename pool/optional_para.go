package pool

import (
	"time"
	"errors"
)

var (
	// Error types for GoroutinePool API
	ErrInvalidPoolSize = errors.New("invalid size for pool")
	ErrInvalidPoolExpiry = errors.New("invalid expiry for pool")
	ErrPoolClosed = errors.New("this pool can't accept new task")
	ErrPoolOverload = errors.New("too many goroutines blocked on submit or Nonblocking is set")
)

type OptionalPara struct {
	// not core free routine clear duration
	ExpiryDuration		time.Duration
	// max tasks in blockingQueue
	MaxBlockingTasks	int64
	// enable when disabled blocking
	Nonblocking 		bool
	// panic handler
	PanicHandler 		func(interface{})
}