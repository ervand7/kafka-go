package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/segmentio/kafka-go"

	"kafka-go-lab/internal/dlq"
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

	dlqWriter := dlq.NewWriter()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		log.Println("🛑 Processor stopping...")
		reader.Close()
		writer.Close()
		dlqWriter.Close()
		cancel()
	}()

	log.Println("🚀 Processor started.")
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("❌ read error: %v", err)
			break
		}

		var order Order
		if err := json.Unmarshal(msg.Value, &order); err != nil {
			dlq.Send(ctx, dlqWriter, dlq.FailedMessage{
				SourceTopic: "orders",
				Error:       "json_unmarshal",
				Payload:     msg.Value,
				Timestamp:   time.Now().UnixMilli(),
			})
			_ = reader.CommitMessages(ctx, msg)
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
			// produce failed → DLQ
			dlq.Send(ctx, dlqWriter, dlq.FailedMessage{
				SourceTopic: "orders",
				Error:       "produce_failed",
				Payload:     msg.Value,
				Timestamp:   time.Now().UnixMilli(),
			})
			_ = reader.CommitMessages(ctx, msg)
			continue
		}

		if err := reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("⚠️ commit error: %v", err)
		} else {
			log.Printf("✅ processed order %d (tax %.2f)", order.ID, enriched.Tax)
		}
	}
}
