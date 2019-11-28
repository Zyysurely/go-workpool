package pool


import (
	"fmt"
	// "errors"
	"sync/atomic"
	"context"
	// "os/signal"
)



type GoroutinePool struct {
	// sync.Mutex                      // 嵌入锁机制

	ctx				context.Context // ctx控制所有goroutine退出信号
	Cap				int32           // 最大容量
	WorkerQueue		chan *Worker    // 池中的woker队列
	Closed			chan bool       // 关闭信号
	Running			int32           // 已经启动的worker数量
	freeRunning		int32		    // 可分配的worker数量
	// Once			sync.Once       // 确定只会关闭一次
}

func NewGoroutinePool(cap int32) *GoroutinePool {
	return &GoroutinePool {
		Cap: cap,
		WorkerQueue: make(chan *Worker, cap),
		Closed: make(chan bool, 1),
	}
}

func (gp *GoroutinePool) AddSingleTask(task *Task) error {
	if atomic.LoadInt32(&gp.freeRunning) == 0 && atomic.LoadInt32(&gp.Running) < gp.Cap {
		w := NewWorker(gp, gp.ctx)
		w.TaskChannel <- *task
		w.Run()
		atomic.AddInt32(&gp.Running, 1)
	} else {
		select {
			// 不能新建了，只能等待空闲的
			case w := <- gp.WorkerQueue:
				atomic.AddInt32(&gp.freeRunning, -1)
				w.TaskChannel <- *task
		}
		// 不能新建了，只能等待空闲的
		w := <- gp.WorkerQueue
		w.TaskChannel <- *task
	}
	return nil
}


// goroutine池启动
func (gp *GoroutinePool) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	gp.ctx = ctx
	// exitChan := make(chan os.Signal, 1)
	// signal.Notify()
	<-gp.Closed
	fmt.Println("Pool need to exit~~~")
	cancel()
}


// goroutine池关闭
func (gp *GoroutinePool) Stop() {
	gp.Closed <- true
}