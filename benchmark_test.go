package GoWorkPool

import (
	"time"
	"testing"
	"sync"
)

var (
	times = 100
)

func demo() {
	time.Sleep(time.Duration(10) * time.Millisecond)
}

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