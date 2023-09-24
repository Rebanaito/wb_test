package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/stan.go"
)

type Order struct {
	Order_uid          string
	Track_number       string
	Entry              string
	Delivery           Delivery
	Payment            Payment
	Items              []Item
	Locale             string
	Internal_signature string
	Customer_id        string
	Delivery_service   string
	Shardkey           string
	Sm_id              int
	Date_created       time.Time
	Oof_shard          string
}

type OrderBlank struct {
	Order_uid          *string
	Track_number       *string
	Entry              *string
	Delivery           DeliveryBlank
	Payment            PaymentBlank
	Items              []ItemBlank
	Locale             *string
	Internal_signature *string
	Customer_id        *string
	Delivery_service   *string
	Shardkey           *string
	Sm_id              *int
	Date_created       time.Time
	Oof_shard          *string
}

type Delivery struct {
	Order_uid string
	Name      string
	Phone     string
	Zip       string
	City      string
	Address   string
	Region    string
	Email     string
}

type DeliveryBlank struct {
	Name    *string
	Phone   *string
	Zip     *string
	City    *string
	Address *string
	Region  *string
	Email   *string
}

type Payment struct {
	Transaction   string
	Request_id    string
	Currency      string
	Provider      string
	Amount        int
	Payment_dt    int
	Bank          string
	Delivery_cost int
	Goods_total   int
	Custom_fee    int
}

type PaymentBlank struct {
	Transaction   *string
	Request_id    *string
	Currency      *string
	Provider      *string
	Amount        *int
	Payment_dt    *int
	Bank          *string
	Delivery_cost *int
	Goods_total   *int
	Custom_fee    *int
}

type Item struct {
	Order_uid    string
	Chrt_id      int
	Track_number string
	Price        int
	Rid          string
	Name         string
	Sale         int
	Size         string
	Total_price  int
	Nm_id        int
	Brand        string
	Status       int
}

type ItemBlank struct {
	Chrt_id      *int
	Track_number *string
	Price        *int
	Rid          *string
	Name         *string
	Sale         *int
	Size         *string
	Total_price  *int
	Nm_id        *int
	Brand        *string
	Status       *int
}

type badMessage struct {
	date  time.Time
	data  []byte
	valid bool
}

var ErrMissingOrderData = errors.New("incoming order has missing data")
var ErrMissingItemData = errors.New("incoming order has missing item data")
var ErrInvalidOrderUid = errors.New("invalid 'order_uid'")
var ErrInvalidTrackNumber = errors.New("invalid 'track_number'")
var ErrInvalidEntry = errors.New("invalid 'entry'")
var ErrInvalidLocale = errors.New("invalid 'locale'")
var ErrInvalidCustomerId = errors.New("invalid 'customer_id'")
var ErrInvalidDeliveryService = errors.New("invalid 'delivery_service'")
var ErrInvalidShardKey = errors.New("invalid 'shardkey'")
var ErrInvalidSmId = errors.New("invalid 'sm_id'")
var ErrInvalidDate = errors.New("invalid 'date_created'")
var ErrInvalidOofShard = errors.New("invalid 'oof_shard'")
var ErrInvalidDeliveryName = errors.New("invalid 'delivery.name'")
var ErrInvalidDeliveryPhone = errors.New("invalid 'delivery.phone'")
var ErrInvalidDeliveryZip = errors.New("invalid 'delivery.zip'")
var ErrInvalidDeliveryCity = errors.New("invalid 'delivery.city'")
var ErrInvalidDeliveryAddress = errors.New("invalid 'delivery.address'")
var ErrInvalidDeliveryRegion = errors.New("invalid 'delivery.region'")
var ErrInvalidDeliveryEmail = errors.New("invalid 'delivery.email'")
var ErrInvalidPaymentTransaction = errors.New("invalid 'payment.transaction'")
var ErrInvalidPaymentCurrency = errors.New("invalid 'payment.currency'")
var ErrInvalidPaymentProvider = errors.New("invalid 'payment.provider'")
var ErrInvalidPaymentAmount = errors.New("invalid 'payment.amount'")
var ErrInvalidPaymentDt = errors.New("invalid 'payment.payment_dt'")
var ErrInvalidPaymentBank = errors.New("invalid 'payment.bank'")
var ErrInvalidPaymentDeliveryCost = errors.New("invalid 'payment.delivery_cost'")
var ErrInvalidPaymentGoodsTotal = errors.New("invalid 'payment.goods_total'")
var ErrInvalidPaymentCustomFee = errors.New("invalid 'payment.custom_fee'")
var ErrInvalidItemChrtId = errors.New("invalid 'item.chrt_id'")
var ErrInvalidItemTrackNumber = errors.New("invalid 'item.track_number'")
var ErrInvalidItemPrice = errors.New("invalid 'item.price'")
var ErrInvalidItemRid = errors.New("invalid 'item.rid'")
var ErrInvalidItemName = errors.New("invalid 'item.name'")
var ErrInvalidItemSale = errors.New("invalid 'item.sale'")
var ErrInvalidItemSize = errors.New("invalid 'item.size'")
var ErrInvalidItemTotal = errors.New("invalid 'item.total_price'")
var ErrInvalidItemNmId = errors.New("invalid 'item.nm_id'")
var ErrInvalidItemBrand = errors.New("invalid 'item.brand'")
var ErrInvalidItemStatus = errors.New("invalid 'item.status'")
var ErrInvalidMath = errors.New("invalid discount calculation")
var ErrInvalidTotal = errors.New("item prices do not add up to total")

func main() {
	// Establishing connection to the NATS Streaming server
	sc, err := stan.Connect("test", "sub-1")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Counld not connect to the NATS Streaming server.")
		os.Exit(1)
	}
	defer sc.Close()

	// Connecting to the PostgreSQL database
	conn, err := pgxpool.New(context.Background(), "postgres://revanite:password@localhost:5432/wb")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating a pgx pool: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Initializing main variables:
	// 'cache' stores all orders in the database currently
	// 'mu' is a mutex pointer that we use to safely write to cache from concurrent goroutines
	// 'ch' is the channel we use to pass on the processed orders to be stored in cache
	// 'logChan' is the channel we use to pass information about invalid messages to be logged
	var cache = make(map[string]Order)
	var mu = &sync.Mutex{}
	var ch = make(chan *Order)
	var logChan = make(chan badMessage)

	// Restoring cache from the database
	restoreCache(cache, conn, context.Background(), mu)

	// Two routines - first one writes to cache, second one logs invalid messages
	go writeToCache(mu, cache, ch)
	go logBadMessages(logChan)

	// Specifying what function should handle incoming HTTP requests
	fileName := "search.html"
	tmpl, _ := template.ParseFiles(fileName)
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		handler(rw, r, cache, mu, tmpl)
	})

	// Subscription to the NATS Streaming channel, specifying the function to handle incoming messages
	sc.Subscribe("model", func(m *stan.Msg) {
		go processMessage(m, conn, ch, logChan)
	})
	// Starting the HTTP server to handle search requests
	log.Fatal(http.ListenAndServe(":8080", nil))
}
