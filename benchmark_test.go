package GoWorkPool

import (
	// "sync"
	"testing"
	"time"
	pool "project/pool"
	// "os/signal"
	// "os"
)

var (
	times = 10000
	DefaultGoroutinePoolSize = 100
	DefaultBlockingTasks = 100
	DefaultExpiredTime = 10 * time.Second
	DefaultOptions = &OptionalPara {
		ExpiryDuration: time.Duration(DefaultExpiredTime),
		MaxBlockingTasks: DefaultBlockingTasks,
		Nonblocking: false,
	}
	defaultGoroutinePool = NewGoroutinePool(DefaultAntsPoolSize, 0.5*DefaultAntsPoolSize, DefaultOptions)
)

func demo() error{
	time.Sleep(time.Duration(10) * time.Millisecond)
	return nil
}

// with origin goroutine schedule
func BenchmarkGoroutines(b *testing.B) {
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(times)
		for j := 0; j < times; j++ {
			go func() {
				demo()
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

// with newly goroutinePool
func BenchmarkWorkPool(b *testing.B) {
	t := pool.NewTask(demo)
	for i := 0; i < b.N; i++ {
		wg.Add(times)
		for j := 0; j < times; j++ {
			defaultGoroutinePool.Submit(
				pool.NewTask(
				func() error {
					err := demo()
					wg.Done()
					return err
				}))
		}
		wg.Wait()
	}
	p.Stop()
}
