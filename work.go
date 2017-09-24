package transfer

import (
	"log"
)

type Job struct {
	offset        int
	blockSize     uint64
	fileName      string
	saveFileBlock string
}

var JobQueue chan Job
var isJobFinish chan bool

type Worker struct {
	WorkerPool  chan chan Job
	JobChannel  chan Job
	isJobFinish chan bool
	quit        chan bool
	c           *Client
}

func NewWorker(workerPool chan chan Job) Worker {
	w := Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		quit:       make(chan bool),
	}
	w.c = NewClient(AddressFileC)
	w.c.Dial()
	return w
}

func (w Worker) Start() {
	go func() {
		for {
			w.WorkerPool <- w.JobChannel
			select {
			case job := <-w.JobChannel:
				//res, err := w.c.Stat(job.fileName)
				err := w.c.DownLoadBlock(job.fileName, job.saveFileBlock, job.offset)
				if err != nil {
					log.Println(err)
					continue
				}
				isJobFinish <- true
			case <-w.quit:
				return
			}
		}
	}()
}

func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()

}
