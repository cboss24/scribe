package main

import (
	"fmt"
	"time"
)

type Worker struct {
	Id          int
	Input       chan BatchMessage
	Quit        chan bool
	WorkerQueue chan chan BatchMessage
}

func NewWorker(id int, workerQueue chan chan BatchMessage) Worker {
	return Worker{
		Id:          id,
		Input:       make(chan BatchMessage),
		Quit:        make(chan bool),
		WorkerQueue: workerQueue,
	}
}

func (w *Worker) Start() {
	go func() {
		for {
			w.WorkerQueue <- w.Input
			select {
			case <-w.Input:
				fmt.Printf("Worker %d Received Input!\n", w.Id)
				time.Sleep(10 * time.Second)
			case <-w.Quit:
				fmt.Printf("Worker %d Quitting!\n", w.Id)
				return
			}

		}
	}()
}

func (w *Worker) Stop() {
	go func() {
		w.Quit <- true
	}()
}
