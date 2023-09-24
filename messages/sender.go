package main

import (
	"fmt"
	"os"
	"time"

	"github.com/nats-io/stan.go"
)

func main() {
	sc, err := stan.Connect("test", "simple-pub")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to the NATS streaming server: %v\n", err)
		os.Exit(1)
	}
	defer sc.Close()
	for {
		for i := 1; i <= 1000; i++ {
			name := fmt.Sprintf("../mock/orders/mock_data_%d.json", i)
			data, _ := os.ReadFile(name)
			sc.Publish("model", data)
			time.Sleep(2 * time.Millisecond)
		}
	}
}
