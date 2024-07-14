package background

import (
	"log"
	"sync"
)

type Job struct {
	Identifier string
	Job        func(args ...interface{}) error
	Args       []interface{}
}

type Worker struct {
	queue         chan Job
	cancellations map[string]bool
	mu            sync.Mutex
}

func (w *Worker) Run() {
	go func() {
		for job := range w.queue {
			go func(job Job) {
				w.mu.Lock()
				if _, canceled := w.cancellations[job.Identifier]; canceled {
					delete(w.cancellations, job.Identifier)
					w.mu.Unlock()
					return
				}
				w.mu.Unlock()

				log.Println("Running job", job.Identifier)
				err := job.Job(job.Args...)
				if err != nil {
					log.Println("Error in job", job.Identifier, err)
				}
			}(job)
		}
	}()
}

func (w *Worker) AddJob(job Job) {
	log.Println("Add job", job.Identifier)
	w.queue <- job
}

func (w *Worker) RemoveJob(identifier string) {
	w.mu.Lock()
	log.Println("Remove job", identifier)
	w.cancellations[identifier] = true
	w.mu.Unlock()
}

func NewWorker() *Worker {
	return &Worker{
		queue:         make(chan Job),
		cancellations: make(map[string]bool),
	}
}
