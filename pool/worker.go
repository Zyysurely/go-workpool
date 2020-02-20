package pool

import (
	"context"
	"time"

	"log"
)

type Worker struct {
	isCore		   bool
	ctx			   context.Context   // or use this to close it
    TaskChannel    chan *Task        // handle task
	Pool		   *GoroutinePool    // Pool owner
	recycleTime    time.Time        
	completedTasks int64
	timer		   time.Timer
	quit           <-chan bool       // read only channel
}

func NewWorker(isCore bool, quit <-chan bool, pool *GoroutinePool) *Worker {
	return &Worker {
		isCore: isCore,
		TaskChannel: make(chan *Task, 1),
		Pool: pool,
		quit: quit,
	}
}

func (w *Worker) completeTask(task *Task) error{
	err := task.Run()
	if err != nil {
		log.Printf("task complete with error: %+v\n", err)
	}
	if !w.Pool.IsRunning() {
		return ErrPoolClosed
	}
	w.recycleTime = time.Now()
	w.completedTasks++
	return nil
}

// runWorker() method
func (w *Worker) Run() {
    go func() {
		// exit handler
		defer func() {
			w.Pool.DecRunning()
			if p := recover(); p != nil {
				if ph := w.Pool.option.PanicHandler; ph != nil {
					ph(p)
				} else {
					log.Printf("worker exits due to panic: %+v\n", p)
				}
			}
		}()
		
		// // 清理定时器
		// if !w.isCore {
		// 	timer := time.NewTimer(2 * time.Second)
		// }

		for task := range w.TaskChannel {
			if task == nil {
				// log.Println("worker exits")
				w.Pool.DecRunning()
				return
			}
			err := w.completeTask(task)
			if err!= nil {
				log.Println("worker exits")
				w.Pool.DecRunning()
				return
			}
			w.Pool.FreeWorker(w)
		}

		// for {
		// 	select {
		// 	case task := <-w.TaskChannel:
		// 		if task == nil {
		// 			// log.Println("worker exits")
		// 			w.Pool.DecRunning()
		// 			return
		// 		}
		// 		err := w.completeTask(task)
		// 		if err!= nil {
		// 			log.Println("worker exits")
		// 			w.Pool.DecRunning()
		// 			return
		// 		}
		// 		w.Pool.FreeWorker(w)
		// 	case <-w.quit:
		// 		return
		// 	}
		// }
		// // 弃用channel传递的方法，因为channel吞吐和select的操作效率并不高，并且一直监听taskqueue的效也是比较低的，改为放池中的状态
		// // getTask()
        // for {
        //     select {
		// 	case <-w.ctx.Done():
		// 		// log.Printf("worker closed~~~ bye~\n")
		// 		return
		// 	default:
		// 		select {
		// 		case task := <-w.TaskChannel:					w.completeTask(task)
		// 		case task := <-w.Pool.taskQueue:
		// 			w.Pool.DecBlocking()
		// 			w.completeTask(task)
		// 		case <-w.ctx.Done():
		// 			// log.Printf("worker closed~~~ bye~\n")
		// 			return
		// 		}
        //     }
        // }
    }()
}