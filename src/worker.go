package main

import (
	"fmt"
	"time"
)

type Worker struct {
	Id          int
	Input       chan BatchEvent
	Quit        chan bool
	WorkerQueue chan chan BatchEvent
}

func NewWorker(id int, workerQueue chan chan BatchEvent) Worker {
	return Worker{
		Id:          id,
		Input:       make(chan BatchEvent),
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
