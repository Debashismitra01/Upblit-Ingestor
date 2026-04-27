package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"UpblitIngestor/models"
	"UpblitIngestor/services"

	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
)

type Consumer struct {
	reader          *kafka.Reader
	aggregator      *services.MetricsAggregator
	traceCollection *mongo.Collection
}

// Constructor
func NewConsumer(broker, topic, groupID string, agg *services.MetricsAggregator, traceCol *mongo.Collection) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     []string{broker},
			Topic:       topic,
			GroupID:     groupID,
			MinBytes:    10e3, // 10KB
			MaxBytes:    10e6, // 10MB
			MaxWait:     1 * time.Second,
			StartOffset: kafka.LastOffset, // change to FirstOffset if needed
		}),
		aggregator:      agg,
		traceCollection: traceCol,
	}
}

// Start consuming messages
func (c *Consumer) Start(ctx context.Context) {
	log.Println("Kafka consumer started...")

	batch := make([]models.Trace, 0, 100)

	for {
		select {
		case <-ctx.Done():
			log.Println("Kafka consumer shutting down...")
			c.flushBatch(batch)
			_ = c.reader.Close()
			return

		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Println("Kafka read error:", err)
				continue
			}

			var trace models.Trace
			if err := json.Unmarshal(msg.Value, &trace); err != nil {
				log.Println("Unmarshal error:", err)
				continue
			}

			batch = append(batch, trace)

			// Flush batch if size reached
			if len(batch) >= 100 {
				c.flushBatch(batch)
				// 2. Store in MongoDB
				docs := make([]interface{}, len(batch))
				for i, t := range batch {
					docs[i] = t
				}

				_, err := c.traceCollection.InsertMany(context.Background(), docs)
				if err != nil {
					log.Println("Mongo trace insert error:", err)
				}
				batch = batch[:0]
			}
		}
	}
}

// Push batch to aggregator
func (c *Consumer) flushBatch(batch []models.Trace) {
	if len(batch) == 0 {
		return
	}

	c.aggregator.AddTraces(batch)
}
