package pool


import (
	// "errors"
	"sync"
	"sync/atomic"
	"context"
	// "os/signal"
	"log"
	"time"
)

type PoolState int
const (
	RUNNING PoolState = iota
	SHUTDOWN
	STOP 
	TERMINATED
)


type GoroutinePool struct {
	sync.Mutex                         // lock

	state           int       		   // goroutine pool state
	ctx				context.Context    // exit 
	maxPoolSize		int64              // max num of created goroutine
	corePoolSize    int64              // core num   
	workerQueue		chan *Worker       // work routine channel
	taskQueue		chan *Task         // task channel
	workCount		int64              // work routine num
	blockCount      int64              // limit the throughput
	once			sync.Once          // ensure close once
	option			*OptionalPara	   // including 

	closed          chan bool          // close signal
}

func NewGoroutinePool(cap int64, core int64, option *OptionalPara) *GoroutinePool {
	res := &GoroutinePool {
		maxPoolSize: cap,
		corePoolSize: core,
		workerQueue: make(chan *Worker, cap),
		taskQueue: make(chan *Task, option.MaxBlockingTasks),
		closed: make(chan bool, 1),
		option: option,
	}
	// go res.clearExpired()
	return res
}

// submit task to the pool with error back
func (gp *GoroutinePool) Submit(task *Task) error{
	if !gp.isRunning() {
		return ErrPoolClosed
	}
	// <= corePoolSize
	if gp.running() <= gp.corePoolSize {
		gp.addWorker(task)
		return nil
	}
	// add task to blockQueue
	if !gp.option.Nonblocking && gp.option.MaxBlockingTasks != 0 && gp.blockCount < gp.option.MaxBlockingTasks {
		gp.taskQueue <- task
		gp.IncreBlocking();
		return nil
	}
	// blockQueue is full, max
	if gp.running() < gp.maxPoolSize {
		gp.addWorker(task)
		return nil
	}
	// reject
	return ErrPoolOverload
}

// addWorker()
func (gp *GoroutinePool) addWorker(task *Task) {
	w := NewWorker(true, gp.ctx, gp)
	gp.workerQueue <- w
	gp.increRunning()
	w.TaskChannel <- *task
	w.Run()
}

// start goroutine pool
func (gp *GoroutinePool) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	gp.ctx = ctx
	// exitChan := make(chan os.Signal, 1)
	// signal.Notify()
	<-gp.closed
	log.Printf("Pool exit signal recieved~~~")
	cancel()
}


// Stop()
func (gp *GoroutinePool) Stop() {
	gp.closed <- true
}

// Terminate()
func (gp *GoroutinePool) Terminate() {
	gp.closed <- true
}

// clears expired workers
func (gp *GoroutinePool) clearExpired() {
	// heartBeat := time.NewTicker()
	// defer heartBeat.Stop()
	
	// for range heartBeat.C {

	// }
}

// ----------------------------------------------
// running state 、num 、 blocking related api
// ----------------------------------------------
func (gp *GoroutinePool) running() int64{
	return atomic.LoadInt64(&gp.workCount);
}

func (gp *GoroutinePool) isRunning() bool{
	if atomic.LoadInt64(&gp.state) == SHUTDOWN || atomic.LoadInt32(&gp.state) == STOP {
		return false
	}
	return true
}

func (gp *GoroutinePool) increRunning() {
	atomic.AddInt64(&gp.workCount, 1)
}

func (gp *GoroutinePool) DecRunning() {
	atomic.AddInt64(&gp.workCount, -1)
}

func (gp *GoroutinePool) IncreBlocking() {
	atomic.AddInt64(&gp.blockCount, 1)
}

func (gp *GoroutinePool) DecBlocking() {
	atomic.AddInt64(&gp.blockCount, -1)
}
// ----------------------------------------------