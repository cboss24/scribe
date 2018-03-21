package main

import (
	_ "database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"os"
)

var DbConn *sqlx.DB

func main() {
	DbConn = sqlx.MustConnect("postgres", fmt.Sprintf(
		"user=%s password=%s dbname=scribe host=db port=5432 sslmode=disable",
		os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD")))
	d := NewDispatcher(1)
	d.Start()
	Poll(os.Getenv("SQS_QUEUE_NAME"))
}
