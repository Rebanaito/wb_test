package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/stan.go"
)

// Function for message processing
func processMessage(m *stan.Msg, conn *pgxpool.Pool, ch chan *Order, logChan chan badMessage) {
	// Datetime and raw byte data are stored for potential logging
	time := time.Now()
	message := m.Data

	// Parsing raw byte data into OrderBlank struct using json package functions.
	// Using Blank struct made of pointers allows us to check which data the json
	// parser was able to find
	orderBlank, err := jsonParser(m.Data)

	if err == nil {
		// Validating fields - all must be present according to the provided model,
		// values should be in valid ranges and add up.
		valid, errors := validOrder(orderBlank)

		if valid {
			// Creating a new struct to now store the values
			order := createOrder(orderBlank)

			// Sanitizing string values using the pgx package function
			sanitizeOrder(order)

			// Inserting the validated and sanitized order into the database
			insert(&order, conn, context.Background(), message, time, ch, logChan)
		} else {
			message = append([]byte(fmt.Sprintf("Invalid order: %v\n", errors)), message...)
			logChan <- badMessage{time, message, false}
		}
	} else {
		message = append([]byte(fmt.Sprintf("JSON parsing error: %v\n", err)), message...)
		logChan <- badMessage{time, message, false}
	}
}

// Function that adds new entries to the database. In case of an error occuring midway (on one of the tables),
// the function rolls back and cleans up the incomplete data it just added and logs the error
func insert(order *Order, conn *pgxpool.Pool, ctx context.Context, message []byte, time time.Time, ch chan *Order, logChan chan badMessage) {
	var err error
	_, err = conn.Exec(ctx, `INSERT INTO orders (order_uid, track_number, entry, locale,
								internal_signature, customer_id, delivery_service, shardkey,
								sm_id, date_created, oof_shard) VALUES ($1, $2, $3, $4, $5,
								$6, $7, $8, $9, $10, $11)`, order.Order_uid, order.Track_number,
		order.Entry, order.Locale, order.Internal_signature, order.Customer_id,
		order.Delivery_service, order.Shardkey, order.Sm_id, order.Date_created, order.Oof_shard)
	if err == nil {
		_, err = conn.Exec(ctx, `INSERT INTO deliveries (order_uid, name, phone, zip, city,
								address, region, email) VALUES ($1, $2, $3, $4, $5, $6, $7,
								$8)`, order.Delivery.Order_uid, order.Delivery.Name, order.Delivery.Phone,
			order.Delivery.Zip, order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
		if err == nil {
			_, err = conn.Exec(ctx, `INSERT INTO payments (transaction, request_id, currency,
									provider, amount, payment_dt, bank, delivery_cost, goods_total,
									custom_fee) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
				order.Payment.Transaction, order.Payment.Request_id, order.Payment.Currency,
				order.Payment.Provider, order.Payment.Amount, order.Payment.Payment_dt,
				order.Payment.Bank, order.Payment.Delivery_cost, order.Payment.Goods_total, order.Payment.Custom_fee)
			if err == nil {
				for _, item := range order.Items {
					_, err = conn.Exec(ctx, `INSERT INTO items (order_uid, chrt_id, track_number,
											price, rid, name, sale, size, total_price, nm_id, brand,
											status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
											$11, $12)`, item.Order_uid, item.Chrt_id, item.Track_number,
						item.Price, item.Rid, item.Name, item.Sale, item.Size,
						item.Total_price, item.Nm_id, item.Brand, item.Status)
					if err != nil {
						revertInserts(order.Order_uid, conn, ctx, 3)
						break
					}
				}
			} else {
				revertInserts(order.Order_uid, conn, ctx, 2)
			}
		} else {
			revertInserts(order.Order_uid, conn, ctx, 1)
		}
	}
	if err != nil {
		logChan <- badMessage{time, message, true}
		ch <- nil
	} else {
		ch <- order
	}
}

// Cleanup function for insert
func revertInserts(order_uid string, conn *pgxpool.Pool, ctx context.Context, level int) error {
	var err error
	switch level {
	case 3:
		query := fmt.Sprintf(`DELETE FROM items WHERE order_uid = "%s"`, order_uid)
		conn.Exec(ctx, query)
		query = fmt.Sprintf(`DELETE FROM payments WHERE transaction = "%s"`, order_uid)
		conn.Exec(ctx, query)
		fallthrough
	case 2:
		query := fmt.Sprintf(`DELETE FROM deliveries WHERE order_uid = "%s"`, order_uid)
		conn.Exec(ctx, query)
		fallthrough
	case 1:
		query := fmt.Sprintf(`DELETE FROM orders WHERE order_uid = "%s"`, order_uid)
		_, err = conn.Exec(ctx, query)
	}
	return err
}

// Function that logs invalid messages
func logBadMessages(logChan chan badMessage) {
	for log := range logChan {
		year, month, day := log.date.Date()
		hour, minute, seconds := log.date.Clock()
		nano := log.date.Nanosecond()
		var name string
		if log.valid {
			name = fmt.Sprintf("./logs/failed/%s-%s-%s_%s:%s:%s:%s", strconv.Itoa(year), strconv.Itoa(int(month)),
				strconv.Itoa(day), strconv.Itoa(hour), strconv.Itoa(minute), strconv.Itoa(seconds), strconv.Itoa(nano))
		} else {
			name = fmt.Sprintf("./logs/invalid/%s-%s-%s_%s:%s:%s:%s", strconv.Itoa(year), strconv.Itoa(int(month)),
				strconv.Itoa(day), strconv.Itoa(hour), strconv.Itoa(minute), strconv.Itoa(seconds), strconv.Itoa(nano))
		}
		err := os.WriteFile(name, log.data, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Log error: %v\n", err)
		}
	}
}

// Function that adds entries to cache
func writeToCache(mu *sync.Mutex, cache map[string]Order, ch chan *Order) {
	for result := range ch {
		if result != nil {
			mu.Lock()
			cache[result.Order_uid] = *result
			mu.Unlock()
		}
	}
}
