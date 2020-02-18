package pool


import (
	"fmt"
	// "errors"
	"sync/atomic"
	"context"
	// "os/signal"
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

	state           PoolState		   // goroutine pool state
	ctx				context.Context    // exit 
	maxPoolSize		int64              // max num of created goroutine
	corePoolSize    int64              // core num   
	workQueue		chan *Worker       // work routine channel
	taskQueue		chan *Task         // task channel
	workCount		int64              // work routine num
	blockCount      int64              // limit the throughput
	once			sync.Once          // ensure close once
	option			*OptionPara		   // including 
}

func NewGoroutinePool(cap int32, option *OptionPara) (*GoroutinePool, error) {
	
	res := &GoroutinePool {
		Cap: cap,
		WorkerQueue: make(chan *Worker, cap),
		Closed: make(chan bool, 1),
		Option: option,
	}
	go res.clearExpired()
	return res
}

// submit task to the pool with error back
func (gp *GoroutinePool) Submit(task Task) error{
	if !isRunning() {
		return ErrPoolClosed
	}
	// <= corePoolSize
	if p.running() <= p.corePoolSize {
		w := NewWorker(gp, gp.ctx)
		w.TaskChannel <- *task
		w.Run()
		return null
	}
	// add task to blockQueue
	if gp.option.MaxBlockingTasks != 0 && gp.blockCount < gp.option.MaxBlockingTasks {
		gp.taskQueue <- *task
		gp.IncreBlocking();
		return null
	}
	// blockQueue is full, max
	if p.running() < p.maxPoolSize {
		w := NewWorker(gp, gp.ctx)
		w.TaskChannel <- *task
		w.Run()
		return null
	}
	// reject
	return ErrPoolOverload
}

// goroutine池启动
func (gp *GoroutinePool) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	gp.ctx = ctx
	// exitChan := make(chan os.Signal, 1)
	// signal.Notify()
	<-gp.Closed
	fmt.Println("Pool exit signal recieved~~~")
	cancel()
}


// Stop()
func (gp *GoroutinePool) Stop() {
	gp.Closed <- true
}

// Terminate()
func (gp *GoroutinePool) Terminate() {
	gp.Closed <- true
}

// clears expired workers
func (gp *GoroutinePool) clearExpired() {
	heartBeat := time.NewTicker()
	defer heartBeat.Stop()
	
	for range heartBeat.C {

	}
}

func (gp *GoroutinePool) IncreBlocking() {
	atomic.AddInt64(&gp.blockCount, 1)
}

func (gp *GoroutinePool) DecreBlocking() {
	atomic.AddInt64(&gp.blockCount, -1)
}

// ----------------------------------------------
// running state and num related private api
// ----------------------------------------------
func (gp *GoroutinePool) running() int64{
	return atomic.LoadInt64(&gp.workCount);
}

func (gp *GoroutinePool) increRunning() {
	atomic.AddInt64(&gp.workCount, 1)
}

func (gp *GoroutinePool) decreRunning() {
	atomic.AddInt64(&gp.workCount, -1)
}

func (gp *GoroutinePool) isRunning() bool{
	if atomic.LoadInt64(&gp.state) == SHUTDOWN || atomic.LoadInt32(&gp.state) == STOP {
		return false
	}
	return true
}
// ----------------------------------------------