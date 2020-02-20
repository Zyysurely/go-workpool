package pool


import (
	// "errors"
	"sync"
	"sync/atomic"
	"context"
	// "os/signal"
	"log"
)

const (
	RUNNING   = iota
	CLOSED
)


type GoroutinePool struct {
	sync.Mutex                         // lock

	state           int32       	   // goroutine pool state
	Ctx				context.Context    // exit 
	maxPoolSize		int64              // max num of created goroutine
	corePoolSize    int64              // core num
	workerQueue		chan *Worker       // running worker Queue with no task for dispach
	// workerQueue		*workerQueue   // running worker Queue
	taskQueue		chan *Task         // task channel
	workCount		int64              // work routine num
	blockCount      int64              // limit the throughput
	// once			sync.Once          // ensure close once
	option			*OptionalPara	   // including 

	closed          chan bool          // close signal
	cond            *sync.Cond         // wait for an idle worker
}

func NewGoroutinePool(cap int64, core int64, option *OptionalPara) *GoroutinePool {
	res := &GoroutinePool {
		maxPoolSize: cap,
		corePoolSize: core,
		workerQueue: make(chan *Worker, cap),
		// workerQueue: NewWorkerQueue(cap),
		taskQueue: make(chan *Task, option.MaxBlockingTasks),
		closed: make(chan bool, 1),
		option: option,
	}
	// handle blocking task
	go res.dispatch()
	return res
}

// submit task to the pool with error back
func (gp *GoroutinePool) Submit(task *Task) error{
	if !gp.IsRunning() {
		return ErrPoolClosed
	}
	// <= corePoolSize
	if gp.running() < gp.corePoolSize {
		gp.addWorker(true, task)
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
		gp.addWorker(false, task)
		return nil
	}
	// reject
	return ErrPoolOverload
}

// freeWorker()
func (gp *GoroutinePool) FreeWorker(worker *Worker) error {
	// gp.Lock()
	// gp.workerQueue.add(worker)
	// // gp.cond.Signal()
	// gp.Unlock()
	gp.workerQueue <- worker
	return nil
}

// addWorker()
func (gp *GoroutinePool) addWorker(core bool, task *Task) {
	w := NewWorker(core, gp.closed, gp)
	// gp.Lock()
	// gp.workerQueue.add(w)
	// gp.Unlock()
	gp.increRunning()
	if task != nil {
		w.TaskChannel <- task
	}
	w.Run()
}

// Stop(), ensure to close once
func (gp *GoroutinePool) Stop() {
	atomic.StoreInt32(&gp.state, CLOSED)
	gp.once.Do(
		func() {
			close(gp.closed)
		})
}

// TODO: clears expired workers
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

func (gp *GoroutinePool) IsRunning() bool{
	if atomic.LoadInt32(&gp.state) == CLOSED {
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

func (gp *GoroutinePool) Blocking() int64{
	return atomic.LoadInt64(&gp.blockCount);
}

func (gp *GoroutinePool) IncreBlocking() {
	atomic.AddInt64(&gp.blockCount, 1)
}

func (gp *GoroutinePool) DecBlocking() {
	atomic.AddInt64(&gp.blockCount, -1)
}
// ----------------------------------------------


// ----------------------------------------------
// channel with dispach 
// ----------------------------------------------
// need not think about the limit
// func (gp *GoroutinePool) getWorker() *Worker{
// 	gp.Lock()
// 	w := gp.workerQueue.poll()
// 	if w != nil {
// 		gp.Unlock();
// 		return w
// 	}
// 	// if gp.running() < gp.maxPoolSize {
// 	// 	gp.Unlock();
// 	// 	gp.addWorker(nil)
// 	// }
// 	gp.cond.Wait()
// 	w = gp.workerQueue.poll()
// 	gp.Unlock()
// 	return w
// }

func (gp *GoroutinePool) clearWorker() {
	
}

func (gp *GoroutinePool) dispatch() {
	for {
		select {
		case <- gp.closed:
			log.Printf("Dispatch exits~~~")
			return
		default:
			select {
			case task := <-gp.taskQueue:
				// handle task blocking in queue
				// gp.DecBlocking();
				// w := gp.getWorker()
				// w.TaskChannel <- task
				gp.DecBlocking();
				worker := <-gp.workerQueue
				worker.TaskChannel <- task
			case <- gp.closed:
				log.Printf("Dispatch exits~~~")
				return
			}
		}
	}
}
// ----------------------------------------------
