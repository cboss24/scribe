package main

import (
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
	"strings"
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
	Attempts      []byte   `db:"attempts"`
	Container     []byte   `db:"container"`
	CreatedAt     time.Time `db:"created_at"`
	DependsOn     []byte   `db:"depends_on"`
	JobDefinition *string   `db:"job_definition"`
	JobId         *string   `db:"job_id"`
	JobName       *string   `db:"job_name"`
	JobQueue      *string   `db:"job_queue"`
	LastChanged   time.Time `db:"last_changed"`
	Parameters    []byte   `db:"parameters"`
	RetryStrategy []byte   `db:"retry_strategy"`
	StartedAt     time.Time `db:"started_at"`
	Status        *string   `db:"status"`
	StatusReason  *string   `db:"status_reason"`
	StoppedAt     time.Time `db:"stopped_at"`
}

func NewQueryArg(e BatchEvent) QueryArg {
	return QueryArg{
		JobRecord: JobRecord{
			Attempts:      Marshal(e.Detail.Attempts),
			Container:     Marshal(e.Detail.Container),
			CreatedAt:     Unix(e.Detail.CreatedAt),
			DependsOn:     Marshal(e.Detail.DependsOn),
			JobDefinition: e.Detail.JobDefinition,
			JobId:         e.Detail.JobId,
			JobName:       e.Detail.JobName,
			JobQueue:      e.Detail.JobQueue,
			LastChanged:   e.Time,
			Parameters:    Marshal(e.Detail.Parameters),
			RetryStrategy: Marshal(e.Detail.RetryStrategy),
			StartedAt:     Unix(e.Detail.StartedAt),
			Status:        e.Detail.Status,
			StatusReason:  e.Detail.StatusReason,
			StoppedAt:     Unix(e.Detail.StoppedAt),
		},
		Overwrite: PreviousStates[*e.Detail.Status],
	}
}

type QueryArg struct {
	JobRecord
	Overwrite []string `db:"overwrite"`
}

func Marshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		log.Fatalln(err)
	}
	return b
}

func Unix(sec *int64) time.Time {
	if sec == nil {
		return time.Unix(0, 0)
	}
	return time.Unix(*sec, 0)
}

func updateJobRecord(db *sqlx.DB, e BatchEvent) {

	columns := []string{
		"attempts", "container", "created_at", "depends_on", "job_definition", "job_id",
		"job_name", "job_queue", "last_changed", "parameters", "retry_strategy", "status",
	}

	switch *e.Detail.Status {
	case "SUCCEEDED", "FAILED":
		columns = append(columns, []string{"started_at", "status_reason", "stopped_at"}...)
	case "RUNNING":
		columns = append(columns, []string{"started_at"}...)
	}

	params := make([]string, len(columns))
	for k, v := range columns {
		params[k] = ":" + v
	}

	update := make([]string, len(columns))
	for k, v := range columns {
		update[k] = v + " = :" + v
	}

	columnCsv := strings.Join(columns, ", ")
	paramCsv := strings.Join(params, ", ")
	updateCsv := strings.Join(update, ", ")

	statement := fmt.Sprintf(`
		INSERT INTO job (%s)
		VALUES (%s)
		ON CONFLICT (job_id) DO
			UPDATE SET %s
			WHERE job.last_changed < :last_changed
			OR (job.last_changed = :last_changed AND job.status IN (:overwrite))
		;`, columnCsv, paramCsv, updateCsv)
	input := NewQueryArg(e)

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
