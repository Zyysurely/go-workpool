package pool

import (
	"context"
	"sync/atomic"
	"fmt"
)


type Worker struct {
	ctx			 context.Context
    TaskChannel  chan Task
	Pool		 *GoroutinePool   // owner
}

func NewWorker(poolFa *GoroutinePool, ctx context.Context) *Worker {
	return &Worker {
		ctx: ctx,
		TaskChannel: make(chan Task, 1),
		Pool: poolFa,
	}
}

// 启动一个worker池中的goroutine
func (w *Worker) Run() {
	// 启动goroutine
    go func() {
		// 一直等待执行
        for {
            select {
				case task := <-w.TaskChannel:
					task.Run()
					w.Pool.WorkerQueue <- w
					atomic.AddInt32(&w.Pool.freeRunning, 1)
				case <-w.ctx.Done():
					fmt.Println("task closed~~~")
					// we have received a signal to stop
					return
            }
        }
    }()
}