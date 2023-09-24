package main

import (
	"encoding/json"
	"strings"
	"time"
)

func jsonParser(data []byte) (OrderBlank, error) {
	var order OrderBlank
	//err := json.Unmarshal(data, &order)
	reader := strings.NewReader(string(data))
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&order)
	return order, err
}

func datePretty(date time.Time) string {
	return strings.Join(strings.Split(date.UTC().String(), " ")[:2], " ")
}
