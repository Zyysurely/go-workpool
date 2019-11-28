package pool


type Worker struct {
    TaskChannel  chan Task
	Closed       chan bool
	Pool		 *GoroutinePool   // owner
}

func NewWorker(poolFa *GoroutinePool, closed chan bool) *Worker {
	return &Worker {
		TaskChannel: make(chan Task),
		Closed: closed,
		Pool: poolFa,
	}
}

// 启动一个worker池中的goroutine
func (w *Worker) Run() {
	// 启动goroutine
    go func() {
		// 一直等待执行
        for {
            w.WorkerPool <- w.TaskChannel
            select {
			case task := <-w.TaskChannel:
				task.Run()
				w.Pool.WorkerQueue <- w
			case <-w.Closed:
				fmt.Println("task closed~~~")
                // we have received a signal to stop
                return
            }
        }
    }()
}

// Stop signals the worker to stop listening for work requests.
func (w *Worker) Stop() {
	w.Closed <- true
}