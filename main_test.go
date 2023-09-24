package main

import (
	"context"
	"fmt"
	"os"
	"path"
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
	var cache = make(map[string]Order)
	var wg sync.WaitGroup
	sc.Subscribe("model", func(m *stan.Msg) {
		wg.Add(1)
		go func() {
			processMessage(m, conn, ch, logChan)
			wg.Done()
		}()
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
	wg.Wait()
	data, _ := os.ReadFile("messages/invalid_field.txt")
	sc.Publish("model", data)
	wg.Wait()
	invalid, _ := os.ReadDir("logs/invalid/")
	if len(invalid) != 1 {
		t.Fatalf("Invalid messages should have been logged. Need: 1, have: %d", len(invalid))
	}
	for _, file := range invalid {
		os.RemoveAll(path.Join([]string{"logs/invalid/", file.Name()}...))
	}
	count := 0
	for key := range cache {
		_ = key
		count++
	}
	if count != 1000 {
		t.Fatalf("cache is supposed to be storing 1000 orders, have: %d", count)
	}
	var newCache = make(map[string]Order)
	restoreCache(newCache, conn, context.Background(), mu)
	count = 0
	for key := range newCache {
		_ = key
		count++
	}
	if count != 1000 {
		t.Fatalf("database is supposed to be storing 1000 orders, have: %d", count)
	}
	files, _ := os.ReadDir("logs/failed/")
	failed := len(files)
	for _, file := range files {
		os.RemoveAll(path.Join([]string{"logs/failed/", file.Name()}...))
	}
	if failed != 1000 {
		t.Fatalf("not all errors were logged, must have 1000 invalid orders in log/invalid, have %d", failed)
	}
}
