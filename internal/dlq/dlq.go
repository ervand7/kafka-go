package dlq

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

type FailedMessage struct {
	SourceTopic string `json:"source_topic"`
	Error       string `json:"error"`
	Payload     []byte `json:"payload"`
	Timestamp   int64  `json:"ts"`
}

func NewWriter() *kafka.Writer {
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{"kafka:9092"},
		Topic:        "orders-dlq",
		BatchTimeout: 20 * time.Millisecond,
	})
}

func Send(ctx context.Context, w *kafka.Writer, fm FailedMessage) {
	bytes, _ := json.Marshal(fm)
	_ = w.WriteMessages(ctx, kafka.Message{
		Key:   []byte(fm.SourceTopic),
		Value: bytes,
	})
}
