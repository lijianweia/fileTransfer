package transfer

type Dispatcher struct {
	WokerPool chan chan Job
	Num       int
}

func NewDispatcher(maxWorkers int) *Dispatcher {
	pool := make(chan chan Job, maxWorkers)

	return &Dispatcher{pool, maxWorkers}
}

func (d *Dispatcher) Run() {
	for i := 0; i < d.Num; i++ {
		worker := NewWorker(d.WokerPool)
		worker.Start()
	}
	go d.dispatch()
}
func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-JobQueue:
			go func(job Job) {
				jobChannel := <-d.WokerPool
				jobChannel <- job
			}(job)
		}
	}

}
