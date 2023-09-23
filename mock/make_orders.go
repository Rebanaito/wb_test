package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type OrderBlank struct {
	Order_uid          *string
	Track_number       *string
	Entry              *string
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

func main() {
	data, _ := os.ReadFile("MOCK_DATA.json")
	lines := strings.Split(string(data), "\n")
	count := 0
	for i := range lines {
		var order OrderBlank
		json.Unmarshal([]byte(lines[i]), &order)
		count += len(order.Items)
		// name := fmt.Sprintf("orders/mock_data_%d.json", i+1)
		// os.WriteFile(name, []byte(lines[i]), 0644)
	}
	fmt.Println(count)
}
