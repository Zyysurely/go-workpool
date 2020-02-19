package GoWorkPool

import (
	"sync"
	"testing"
	"time"
	pool "project/pool"
	// "os/signal"
	// "os"
	"context"
	"log"
)

var (
	times = 1000000
	DefaultGoroutinePoolSize = int64(200000)
	DefaultBlockingTasks = int64(1000000)
	DefaultExpiredTime = 10 * time.Second
	DefaultOptions = &pool.OptionalPara {
		ExpiryDuration: time.Duration(DefaultExpiredTime),
		MaxBlockingTasks: int64(DefaultBlockingTasks),
		Nonblocking: false,
	}
	defaultGoroutinePool = pool.NewGoroutinePool(DefaultGoroutinePoolSize, DefaultGoroutinePoolSize, DefaultOptions)
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
func BenchmarkGoroutinePool(b *testing.B) {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defaultGoroutinePool.Ctx = ctx
	defer defaultGoroutinePool.Stop()

	// b.StartTimer()
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
	log.Printf("Pool exit signal recieved~~~")
	cancel()
	// b.StopTimer()
}
