package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"

	"github.com/segmentio/kafka-go"
)

type Order struct {
	ID     int     `json:"id"`
	Amount float64 `json:"amount"`
}

type OrderWithTax struct {
	ID     int     `json:"id"`
	Amount float64 `json:"amount"`
	Tax    float64 `json:"tax"`
}

func main() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{"kafka:9092"},
		GroupID:        "order-processor-group",
		Topic:          "orders",
		CommitInterval: 0,
	})

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"kafka:9092"},
		Topic:   "orders-taxed",
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		log.Println("🛑 Shutting down processor...")
		reader.Close()
		writer.Close()
		cancel()
	}()

	log.Println("🚀 Processor started.")
	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("❌ read error: %v", err)
			continue
		}

		var order Order
		if err := json.Unmarshal(m.Value, &order); err != nil {
			log.Printf("⚠️ invalid message: %s", m.Value)
			continue
		}

		enriched := OrderWithTax{
			ID:     order.ID,
			Amount: order.Amount,
			Tax:    order.Amount * 0.2,
		}

		value, _ := json.Marshal(enriched)

		if err := writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte("tax"),
			Value: value,
		}); err != nil {
			log.Printf("❌ write error: %v", err)
		} else {
			log.Printf("✅ processed order %d → taxed %.2f", order.ID, enriched.Tax)
		}

		if err := reader.CommitMessages(ctx, m); err != nil {
			log.Printf("⚠️ commit error: %v", err)
		}
	}
}
