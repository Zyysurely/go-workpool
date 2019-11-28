package pool

type Task struct {
	// task中的任务函数，这里可自定义，如果对应复杂的功能，可以直接对应一个interface
	f func() error;
}

func NewTask(f func() error) *Task {
	return &Task {
		f: f,
	}
}

func (t *Task) Run() error{
	err := t.f()
	return err
}
