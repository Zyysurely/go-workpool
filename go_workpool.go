package pool


import (
	"sync"
	"fmt"
	"error"
)



type GoroutinePool struct {
	// sync.Mutex                    // 嵌入锁机制

	Cap				int32         // 最大容量
	WorkerQueue		chan *Worker  // 池中的woker队列
	TaskSum			int32         // 池中可承载的task
	Closed			chan bool     // 关闭信号
	Once			sync.Once     // 确定只会关闭一次
}

func NewGoroutinePool(cap int32, expireTime int64) *GoroutinePool {
	return &GoroutinePool {
		Cap: cap,
		ExpireTime: expireTime,
		WorkerQueue: make(chan *Worker, cap),
		Closed: make(chan int32),
	}
}

func (gp *GoroutinePool) AddSingleTask(task *Task) error {
	w := gp.getWorker()
	w.set(task);
}

func (gp *GoroutinePool) getWorker() *Worker {
	select {
		case w := <- gp.WorkerQueue: 
	}
}