package main

import (
	"context"
	_ "database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"os"
	"os/signal"
	"sync"
)

var DbConn *sqlx.DB

// TODO make these flags
const (
	concurrency = 1
	buffer      = 100
)

func main() {
	DbConn = sqlx.MustConnect("postgres", fmt.Sprintf(
		"user=%s password=%s dbname=scribe host=db port=5432 sslmode=disable",
		os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD")))

	messages := make(chan BatchMessage, buffer)
	queue := os.Getenv("SQS_QUEUE_NAME")

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			cancel()

		case <-ctx.Done():
		}
	}()

	go Poll(ctx, queue, messages)

	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go Worker(&wg, messages)
	}
	wg.Wait()
}
