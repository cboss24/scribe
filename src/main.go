package main

import (
	_ "database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"os"
)

func main() {
	_, err := sqlx.Connect("postgres", fmt.Sprintf(
		"user=%s password=%s dbname=scribe host=db port=5432 sslmode=disable",
		os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD")))
	if err != nil {
		log.Fatalln(err)
	}

	n := 10
	d := NewDispatcher(n)
	d.Start()
	Poll(os.Getenv("SQS_QUEUE_NAME"))
}
