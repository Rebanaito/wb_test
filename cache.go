package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// This function queries all entries from the 'orders' table and uses the pgx package to parse the
// result of the query into structs
func restoreCache(cache map[string]Order, conn *pgxpool.Pool, ctx context.Context, mu *sync.Mutex) {
	rows, err := conn.Query(ctx, "SELECT * FROM orders")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not query the database: %v\n", err)
		os.Exit(1)
	}
	orders, err := pgx.CollectRows[Order](rows, pgx.RowToStructByNameLax[Order])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse rows: %v\n", err)
		os.Exit(1)
	}
	rows.Close()
	var wg sync.WaitGroup
	for _, order := range orders {
		wg.Add(1)
		go queryOrder(cache, conn, ctx, order, &wg, mu)
	}
	wg.Wait()
}

// This function queries tables 'payments', 'deliveries', and 'items' using the order_uid for each order
func queryOrder(cache map[string]Order, conn *pgxpool.Pool, ctx context.Context, order Order, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()
	query := "SELECT * FROM deliveries WHERE order_uid = '" + order.Order_uid + "'"
	rows, err := conn.Query(ctx, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not query the database: %v\n", err)
		os.Exit(1)
	}
	delivery, err := pgx.CollectRows[Delivery](rows, pgx.RowToStructByNameLax[Delivery])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse rows: %v\n", err)
		os.Exit(1)
	}
	rows.Close()
	order.Delivery = delivery[0]
	query = "SELECT * FROM payments WHERE transaction = '" + order.Order_uid + "'"
	rows, err = conn.Query(ctx, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not query the database: %v\n", err)
		os.Exit(1)
	}
	payment, err := pgx.CollectRows[Payment](rows, pgx.RowToStructByNameLax[Payment])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse rows: %v\n", err)
		os.Exit(1)
	}
	rows.Close()
	order.Payment = payment[0]
	query = "SELECT * FROM items WHERE order_uid = '" + order.Order_uid + "'"
	rows, err = conn.Query(ctx, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not query the database: %v\n", err)
		os.Exit(1)
	}
	items, err := pgx.CollectRows[Item](rows, pgx.RowToStructByNameLax[Item])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse rows: %v\n", err)
		os.Exit(1)
	}
	rows.Close()
	order.Items = items
	sanitizeOrder(order)
	mu.Lock()
	cache[order.Order_uid] = order
	mu.Unlock()
}
