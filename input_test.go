package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/stan.go"
)

func TestMain(t *testing.T) {
	sc, err := stan.Connect("test", "sub-1")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Counld not connect to the NATS Streaming server.")
		os.Exit(1)
	}
	defer sc.Close()
	conn, err := pgxpool.New(context.Background(), "postgres://revanite:password@localhost:5432/wb")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating a pgx pool: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	conn.Exec(context.Background(), "DELETE FROM items")
	conn.Exec(context.Background(), "DELETE FROM deliveries")
	conn.Exec(context.Background(), "DELETE FROM payments")
	conn.Exec(context.Background(), "DELETE FROM orders")
	var mu = &sync.Mutex{}
	var ch = make(chan *Order)
	var logChan = make(chan badMessage)
	var wg sync.WaitGroup
	var cache = make(map[string]Order)
	sc.Subscribe("model", func(m *stan.Msg) {
		wg.Add(1)
		go processMessage(m, conn, &wg, ch, logChan)
	})
	go writeToCache(mu, cache, ch)
	go logBadMessages(logChan)
	for i := 0; i < 2; i++ {
		for i := 1; i <= 1000; i++ {
			name := fmt.Sprintf("mock/orders/mock_data_%d.json", i)
			data, _ := os.ReadFile(name)
			sc.Publish("model", data)
			time.Sleep(2 * time.Millisecond)
		}
	}
	var newCache = make(map[string]Order)
	restoreCache(newCache, conn, context.Background(), mu)
	count := 0
	for key := range cache {
		_ = key
		count++
	}
	if count != 1000 {
		t.Fatalf("cache is missing some orders after receiving messages")
	}
	count = 0
	for key := range newCache {
		_ = key
		count++
	}
	if count != 1000 {
		t.Fatalf("cache is missing some orders after restoring from the DB")
	}

}
