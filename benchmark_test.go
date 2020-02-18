package GoWorkPool

import (
	"sync"
	"testing"
	"time"
)

var (
	times = 100
)

func demo() {
	time.Sleep(time.Duration(10) * time.Millisecond)
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

// with newly
func BenchmarkWorkPool(b *testing.B) {
	t := gp.NewTask(
		func() error {
			// fmt.Println(time.Now())
			// 耗时代表
			time.Sleep(5000)
			return nil
		})

	p := gp.NewGoroutinePool(4)
	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt, os.Kill)
	exitFlag := make(chan bool, 1)
	go func() {
		for {
			select {
			case <-exitFlag:
				break
			default:
				p.AddSingleTask(t)
			}
		}
		fmt.Println("sendStop")
	}()
	p.Run()
	<-exitChan
	fmt.Println("os want to stop")
	exitFlag <- true
	p.Stop()
}
