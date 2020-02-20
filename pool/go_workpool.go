package pool


import (
	// "errors"
	"sync"
	"sync/atomic"
	"context"
	// "os/signal"
	"log"

	"project/internal"
	"runtime"
	"time"
)

const (
	RUNNING   = iota
	CLOSED
)


type GoroutinePool struct {
	sync.Mutex                         // lock
	workerLocker	sync.Locker	   // lock for workerQueue
	state           int32       	   // goroutine pool state
	Ctx				context.Context    // exit 
	maxPoolSize		int64              // max num of created goroutine
	corePoolSize    int64              // core num
	// workerQueue		chan *Worker       // running worker Queue with no task for dispach
	workerQueue		*workerQueue       // running worker Queue
	taskQueue		chan *Task         // task channel
	workCount		int64              // work routine num
	blockCount      int64              // limit the throughput
	once			sync.Once          // ensure close once
	option			*OptionalPara	   // including 

	closed          chan bool          // close signal
	cond            *sync.Cond         // wait for an idle worker
}

func NewGoroutinePool(cap int64, core int64, option *OptionalPara) *GoroutinePool {
	res := &GoroutinePool {
		maxPoolSize: cap,
		corePoolSize: core,
		// workerQueue: make(chan *Worker, cap),
		workerQueue: NewWorkerQueue(cap),
		taskQueue: make(chan *Task, option.MaxBlockingTasks),
		closed: make(chan bool, 1),
		option: option,
		workerLocker: internal.NewSpinLock(),
		state: RUNNING,
	}
	res.cond = sync.NewCond(res.workerLocker)

	// handle blocking task
	// go res.dispatch()

	// clear goroutine
	go res.clearExpired()
	return res
}

// submit task to the pool with error back
func (gp *GoroutinePool) Submit(task *Task) error{
	var w *Worker
	if w = gp.retrieveWorker(); w == nil {
		return ErrPoolOverload
	}
	w.TaskChannel <- task
	return nil
	// // if !gp.IsRunning() {
	// // 	return ErrPoolClosed
	// // }
	// // <= corePoolSize
	// if gp.running() < gp.corePoolSize {
	// 	gp.addWorker(true, task)
	// 	return nil
	// }
	// // add task to blockQueue
	// if !gp.option.Nonblocking && gp.option.MaxBlockingTasks != 0 && gp.blockCount < gp.option.MaxBlockingTasks {
	// 	gp.IncreBlocking();
	// 	w := gp.getWorker()
	// 	w.TaskChannel <- task
	// 	return nil
	// }
	// // blockQueue is full, max
	// if gp.running() < gp.maxPoolSize {
	// 	gp.addWorker(false, task)
	// 	return nil
	// }
	// // reject
	// return ErrPoolOverload
}

// freeWorker()
func (gp *GoroutinePool) FreeWorker(worker *Worker) error {
	if !gp.IsRunning() {
		return nil
	}
	gp.workerLocker.Lock()
	gp.workerQueue.add(worker)
	gp.cond.Signal()
	gp.workerLocker.Unlock()
	// gp.workerQueue <- worker
	return nil
}

// addWorker()
func (gp *GoroutinePool) addWorker(core bool, task *Task) {
	w := NewWorker(core, gp.Ctx, gp)
	// gp.Lock()
	// gp.workerQueue.add(w)
	// gp.Unlock()
	gp.increRunning()
	if task != nil {
		w.TaskChannel <- task
	}
	w.Run()
}

// Stop()
func (gp *GoroutinePool) Stop() {
	// atomic.StoreInt32(&gp.state, CLOSED)
	gp.once.Do(func(){
		close(gp.closed)
	})
	gp.workerLocker.Lock()
	gp.workerQueue.clear()
	gp.workerLocker.Unlock()
	log.Printf("#goroutines: %d\n", runtime.NumGoroutine())
}

// clears expired workers
func (gp *GoroutinePool) clearExpired() {
	heartbeat := time.NewTicker(gp.option.ExpiryDuration)
	defer heartbeat.Stop()

	for range heartbeat.C {
		if !gp.IsRunning() {
			break
		}

		gp.workerLocker.Lock()
		expiredWorkers := gp.workerQueue.retrieveExpiry(gp.option.ExpiryDuration)
		gp.workerLocker.Unlock()

	
		for i := range expiredWorkers {
			expiredWorkers[i].TaskChannel <- nil
		}

		if gp.running() == 0 {
			gp.cond.Broadcast()
		}
	}
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
func (gp *GoroutinePool) getWorker() *Worker{
	gp.workerLocker.Lock()
	w := gp.workerQueue.poll()
	if w != nil {
		// log.Printf("get")
		gp.workerLocker.Unlock()
		return w
	}
	Retry:
		// if gp.running() < gp.maxPoolSize {
		// 	gp.Unlock();
		// 	gp.addWorker(nil)
		// }
		// log.Printf("wait")
		gp.cond.Wait()
		// log.Printf("get")
		w = gp.workerQueue.poll()
		if w == nil {
			goto Retry
		}
	gp.workerLocker.Unlock()
	return w
}

// retrieveWorker returns a available worker to run the tasks.
func (gp *GoroutinePool) retrieveWorker() *Worker {
	var w *Worker
	spawnWorker := func() {
		w = NewWorker(false, gp.Ctx, gp)
		w.Run()
		gp.increRunning()
	}

	gp.workerLocker.Lock()

	w = gp.workerQueue.poll()
	if w != nil {
		gp.workerLocker.Unlock()
	} else if gp.running() < gp.corePoolSize {
		gp.workerLocker.Unlock()
		spawnWorker()
	} else {
		if gp.option.Nonblocking {
			gp.workerLocker.Unlock()
			return nil
		}
	Reentry:
		if gp.option.MaxBlockingTasks != 0 && gp.blockCount >= gp.option.MaxBlockingTasks {
			gp.workerLocker.Unlock()
			return nil
		}
		gp.blockCount++
		gp.cond.Wait()
		gp.blockCount--
		if gp.running() == 0 {
			gp.workerLocker.Unlock()
			spawnWorker()
			return w
		}

		w = gp.workerQueue.poll()
		if w == nil {
			goto Reentry
		}

		gp.workerLocker.Unlock()
	}
	return w
}

// instead of using one dispath
func (gp *GoroutinePool) dispatch() {
	for {
		select {
		case task := <-gp.taskQueue:
			// handle task blocking in queue
			gp.DecBlocking();
			w := gp.getWorker()
			w.TaskChannel <- task
			// gp.DecBlocking();
			// worker := <-gp.workerQueue
			// worker.TaskChannel <- task
		case <- gp.closed:
			log.Printf("Dispatch exits~~~")
			return
		}
	}
}
// ----------------------------------------------
