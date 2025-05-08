package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type Order struct {
	ID     int     `json:"id"`
	Amount float64 `json:"amount"`
}

func main() {
	topic := os.Getenv("TOPIC")
	if topic == "" {
		topic = "orders"
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{"kafka:9092"},
		Topic:    topic,
		Balancer: &kafka.Hash{},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		log.Println("⏹  Stopping producer...")
		_ = writer.Close()
		cancel()
	}()

	rand.Seed(time.Now().UnixNano())
	id := 1

	log.Println("🚀 Producer started. Ctrl-C to stop.")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			order := Order{
				ID:     id,
				Amount: float64(rand.Intn(10_000)) / 100.0, // 0.00–99.99
			}
			data, _ := json.Marshal(order)

			msg := kafka.Message{
				Key:   []byte(strconv.Itoa(order.ID)),
				Value: data,
			}

			if err := writer.WriteMessages(ctx, msg); err != nil {
				log.Printf("❌ write error: %v", err)
			} else {
				log.Printf("✅ sent: %+v", order)
			}

			id++
			time.Sleep(2 * time.Second)
		}
	}
}
