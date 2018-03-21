package main

import (
	_ "database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
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
			case m := <-w.Input:
				fmt.Printf("Worker %d Received Input!\n", w.Id)
				updateJobRecord(DbConn, m.Event)
				deleteMessage()
			case <-w.Quit:
				fmt.Printf("Worker %d Quitting!\n", w.Id)
				return
			}

		}
	}()
}

var Previous map[string][]string = map[string][]string{
	"SUBMITTED": {},
	"PENDING":   {"SUBMITTED"},
	"RUNNABLE":  {"PENDING", "RUNNING"},
	"STARTING":  {"RUNNABLE"},
	"RUNNING":   {"STARTING"},
	"SUCCEEDED": {"RUNNING"},
	"FAILED":    {"RUNNING"},
}

func updateJobRecord(db *sqlx.DB, e BatchEvent) {
	statement := `
	INSERT INTO job (batch_id, status, last_changed)
	VALUES (?, ?, ?)
	ON CONFLICT (batch_id) DO
		UPDATE SET status = ?, last_changed = ?
		WHERE job.last_changed < ? OR (job.last_changed = ? AND job.status IN (?))
	;`
	sql, args, err := sqlx.In(statement, *e.Detail.JobId, *e.Detail.Status, e.Time, *e.Detail.Status, e.Time, e.Time, e.Time, Previous[*e.Detail.Status])
	if err != nil {
		log.Fatalln(err)
	}
	sql = db.Rebind(sql)

	_, err = db.Exec(sql, args...)
	if err != nil {
		log.Fatalln(err)
	}
}

func deleteMessage() {

}

func (w *Worker) Stop() {
	go func() {
		w.Quit <- true
	}()
}
