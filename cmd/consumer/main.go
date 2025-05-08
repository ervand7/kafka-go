package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/segmentio/kafka-go"
)

type Order struct {
	ID     int     `json:"id"`
	Amount float64 `json:"amount"`
}

func main() {
	// 1. Topic from env
	topic := os.Getenv("TOPIC")
	if topic == "" {
		topic = "orders"
	}

	groupID := "order-consumer-group"

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{"kafka:9092"},
		GroupID:     groupID,
		Topic:       topic,
		StartOffset: kafka.FirstOffset,
		MaxWait:     500 * time.Millisecond,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		log.Println("🛑 Interrupt received. Closing consumer...")
		_ = reader.Close()
		cancel()
	}()

	log.Println("📥 Kafka consumer started.")
	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("❌ read error: %v", err)
			break
		}

		var order Order
		if err := json.Unmarshal(m.Value, &order); err != nil {
			log.Printf("⚠️  failed to parse message: %s", m.Value)
			continue
		}

		log.Printf("✅ Read message at offset %d (partition %d): %+v", m.Offset, m.Partition, order)
	}
}
