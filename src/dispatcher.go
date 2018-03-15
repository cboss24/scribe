package main

import "fmt"

type Dispatcher struct {
	WorkerQueue chan chan BatchMessage
	WorkerCount int
}

func NewDispatcher(workerCount int) Dispatcher {
	return Dispatcher{
		WorkerQueue: make(chan chan BatchMessage, workerCount),
		WorkerCount: workerCount,
	}
}

func (d *Dispatcher) Start() {
	for i := 0; i < d.WorkerCount; i++ {
		fmt.Printf("Starting Worker %d\n", i)
		worker := NewWorker(i, d.WorkerQueue)
		worker.Start()
	}

	go func() {
		for {
			select {
			case event := <-EventQueue:
				worker := <-d.WorkerQueue
				fmt.Println("Sending event to worker.")
				worker <- event
			}
		}
	}()
}

func (d *Dispatcher) Stop() {
	for i := 0; i < d.WorkerCount; i++ {
		fmt.Printf("Stopped Worker %d\n", i)
		worker := NewWorker(i, d.WorkerQueue)
		worker.Start()
	}
}
