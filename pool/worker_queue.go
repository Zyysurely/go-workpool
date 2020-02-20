package pool

import (
	// "log"
	"time"
)

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

 // clear worker routine
 func (wq *workerQueue) clear() {
	for i := 0; i < len(wq.workers); i++ {
		// log.Printf("%d", len(wq.workers))
		wq.workers[i].TaskChannel <- nil
		// close(wq.workers[i].TaskChannel)
	}
	wq.workers = wq.workers[:0]
 }

 // stop timeout free worker
 func (wq *workerQueue) retrieveExpiry(duration time.Duration) []*Worker {
	n := len(wq.workers)
	if n == 0 {
		return nil
	}

	expiryTime := time.Now().Add(-duration)
	index := wq.binarySearch(0, n-1, expiryTime)

	wq.expiry = wq.expiry[:0]
	if index != -1 {
		wq.expiry = append(wq.expiry, wq.workers[:index+1]...)
		m := copy(wq.workers, wq.workers[index+1:])
		wq.workers = wq.workers[:m]
	}
	return wq.expiry
}

func (wq *workerQueue) binarySearch(l, r int, expiryTime time.Time) int {
	var mid int
	for l <= r {
		mid = (l + r) / 2
		if expiryTime.Before(wq.workers[mid].recycleTime) {
			r = mid - 1
		} else {
			l = mid + 1
		}
	}
	return r
}

    