package pool

import (
	"context"
	"time"

	"log"
)

type Worker struct {
	isCore		   bool
	ctx			   context.Context
    TaskChannel    chan Task        // handle task
	Pool		   *GoroutinePool   // Pool owner
	recycleTime    time.Time
	completedTasks int64
}

func NewWorker(isCore bool, ctx context.Context, pool *GoroutinePool) *Worker {
	return &Worker {
		isCore: isCore,
		ctx: ctx,
		TaskChannel: make(chan Task, 1),
		Pool: pool,	
	}
}

func (w *Worker) completeTask(task *Task) {
	err := task.Run()
	if err != nil {
		log.Errorf("task complete with error: %+v\n", err)
	}
	w.recycleTime = time.Now()
	w.compeletedTasks++
}

// runWorker() method
func (w *Worker) Run() {
    go func() {
		// exit handler
		defer func() {
			w.Pool.DecRunning()
			if p := recover(); p != nil {
				if ph := w.pool.option.PanicHandler; ph != nil {
					ph(p)
				} else {
					log.Printf("worker exits due to panic: %+v\n", p)
				}
			}
		}()

		// getTask()
        for {
            select {
				case <-w.ctx.Done():
					log.Printf("worker closed~~~ bye~\n")
					return
				default:
				select {
					case task := <-w.TaskChannel:
						completeTask(task)
					case task := <-w.pool.taskQueue:
						w.Pool.DecBlocking()
						completeTask(task)
				}
            }
        }
    }()
}