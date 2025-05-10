package main

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"github.com/segmentio/kafka-go"
)

type OrderWithTax struct {
	ID     int     `json:"id"`
	Amount float64 `json:"amount"`
	Tax    float64 `json:"tax"`
}

func main() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{"kafka:9092"},
		GroupID:        "order-reporter-group",
		Topic:          "orders-taxed",
		CommitInterval: 0, // manual commit
	})

	var totalCents uint64 // atomic counter: tax in cents

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Ctrl-C shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		log.Println("🛑  Reporter stopping...")
		reader.Close()
		cancel()
	}()

	// periodic summary
	go func() {
		tick := time.NewTicker(10 * time.Second)
		for range tick.C {
			cents := atomic.LoadUint64(&totalCents)
			log.Printf("📊  total tax so far: %.2f", float64(cents)/100)
		}
	}()

	log.Println("📥  Reporter started.")
	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("❌ read error: %v", err)
			return
		}

		var o OrderWithTax
		if err := json.Unmarshal(m.Value, &o); err != nil {
			log.Printf("⚠️ bad msg: %s", m.Value)
			continue
		}

		// convert tax → cents, then atomically add
		cents := uint64(math.Round(o.Tax * 100))
		atomic.AddUint64(&totalCents, cents)

		if err := reader.CommitMessages(ctx, m); err != nil {
			log.Printf("⚠️ commit error: %v", err)
		}
	}
}
