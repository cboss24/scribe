package main

import (
	_ "database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
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

var PreviousStates map[string][]string = map[string][]string{
	"SUBMITTED": {},
	"PENDING":   {"SUBMITTED"},
	"RUNNABLE":  {"PENDING", "RUNNING"},
	"STARTING":  {"RUNNABLE"},
	"RUNNING":   {"STARTING"},
	"SUCCEEDED": {"RUNNING"},
	"FAILED":    {"RUNNING"},
}

type JobRecord struct {
	BatchId     string    `db:"batch_id"`
	Status      string    `db:"status"`
	LastChanged time.Time `db:"last_changed"`
}

type QueryArg struct {
	JobRecord
	Overwrite []string `db:"overwrite"`
}

func updateJobRecord(db *sqlx.DB, e BatchEvent) {
	statement := `
	INSERT INTO job (batch_id, status, last_changed)
	VALUES (:batch_id, :status, :last_changed)
	ON CONFLICT (batch_id) DO
		UPDATE SET status = :status, last_changed = :last_changed
		WHERE job.last_changed < :last_changed OR (job.last_changed = :last_changed AND job.status IN (:overwrite))
	;`
	input := QueryArg{
		JobRecord: JobRecord{BatchId: *e.Detail.JobId, Status: *e.Detail.Status, LastChanged: e.Time},
		Overwrite: PreviousStates[*e.Detail.Status],
	}
	query, args, err := sqlx.Named(statement, input)
	if err != nil {
		log.Fatalln(err)
	}
	sql, args, err := sqlx.In(query, args...)
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
