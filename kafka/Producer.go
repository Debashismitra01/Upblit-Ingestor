package kafka

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"UpblitIngestor/models"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(broker string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(broker),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			BatchSize:    100,
			BatchTimeout: 10 * time.Millisecond,
		},
	}
}

func (p *Producer) SendTraces(docs []models.Trace) error {
	msgs := make([]kafka.Message, 0, len(docs))

	for _, doc := range docs {
		data, err := json.Marshal(doc)
		if err != nil {
			return err
		}

		msgs = append(msgs, kafka.Message{
			Key:   []byte(strconv.FormatInt(doc.ApplicationID, 10)), // important for partitioning
			Value: data,
		})
	}

	return p.writer.WriteMessages(context.Background(), msgs...)
}

func (p *Producer) Close() {
	if err := p.writer.Close(); err != nil {
		log.Println("error closing producer:", err)
	}
}
