package pool

type Task struct {
	// handle task
	f func() error
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
