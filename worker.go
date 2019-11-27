package pool


type Worker struct {
    TaskChannel  chan Task
    Closed        chan bool
}

func NewWorker() *Worker {
	return &Worker {
		TaskChannel: make(chan Task),
		Closed: make(chan bool),
	}
}

// 启动一个worker池中的goroutine
func (w *Worker) Start() {
    go func() {
        for {
            w.WorkerPool <- w.TaskChannel
            select {
            case task := <-w.TaskChannel:
                if err := job.Payload.UploadToS3(); err != nil {
                    log.Errorf("Error uploading to S3: %s", err.Error())
                }

            case <-w.quit:
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