package main

import (
	"encoding/json"
	"os"
	"strings"
	"time"
)

func jsonParser(data []byte) (OrderBlank, error) {
	var order OrderBlank
	err := json.Unmarshal(data, &order)
	return order, err
}

func datePretty(date time.Time) string {
	return strings.Join(strings.Split(date.UTC().String(), " ")[:2], " ")
}

func countItems() int {
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
	return count
}
