package pool


// not channel type worker queue
type workerQueue struct {
	workers	[]*Worker
	expiry  []*Worker
	size	int64
}

func NewWorkerQueue(size int64) *workerQueue {
	return &workerQueue{
		workers: make([]*Worker, 0, size),
		size: size,
	}
}

// add a worker
func (wq *workerQueue) add(worker *Worker) error {
	wq.workers = append(wq.workers, worker)
	return nil
 }

 // detach a worker
 func (wq *workerQueue) poll() *Worker {
	l := len(wq.workers)
	if l == 0 {
		return nil
	}
	res := wq.workers[l-1]
	wq.workers = wq.workers[:l-1]
	return res
 }
    