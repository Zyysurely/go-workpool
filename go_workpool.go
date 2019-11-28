package pool


import (
	"sync"
	"fmt"
	"error"
	"atomic"
	"context"
)



type GoroutinePool struct {
	sync.Mutex                 // 嵌入锁机制

	Cap				int32         // 最大容量
	WorkerQueue		chan *Worker  // 池中的woker队列
	Closed			chan bool     // 关闭信号
	Running			int32         // 已经启动的worker数量
	freeRunning		int32		  // 可分配的worker数量
	Once			sync.Once     // 确定只会关闭一次
}

func NewGoroutinePool(cap int32, expireTime int64) *GoroutinePool {
	return &GoroutinePool {
		Cap: cap,
		WorkerQueue: make(chan *Worker, cap),
		Closed: make(chan int32),
	}
}

func (gp *GoroutinePool) AddSingleTask(ctx *context.Context, task *Task) error {
	if(gp.freeRunning == 0 && gp.Running < gp.Cap) {
		w := NewWorker(gp, dp.Closed)
			w.TaskChannel <- task
			worker.Run()
			atomic.AddInt32(&gp.Running, 1)
	} else {
		// 只能等待空闲的
		w := <- gp.WorkerQueue
		w.TaskChannel <- task
	}
}


// goroutine池启动
func (gp *GoroutinePool) Run() {
	ctx, cancel = context.WithCancel(context.Background())
	<-gp.Closed
	cancel()
}


// goroutine池关闭
func (gp *GoroutinePool) stop(ctx *context.Context) {
	gp.Closed <- true
}