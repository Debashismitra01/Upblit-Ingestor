package worker

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"New_folder/kafka"
	"New_folder/services"

	"go.mongodb.org/mongo-driver/mongo"
)

type Worker struct {
	consumer    *kafka.Consumer
	aggregator  *services.MetricsAggregator
	metricsCol  *mongo.Collection
	traceCol    *mongo.Collection
	flushTicker *time.Ticker
}

// Constructor
func NewWorker(
	broker string,
	topic string,
	groupID string,
	metricsCol *mongo.Collection,
	traceCol *mongo.Collection,
) *Worker {

	agg := services.NewMetricsAggregator()

	consumer := kafka.NewConsumer(
		broker,
		topic,
		groupID,
		agg,
		traceCol, // ✅ now properly passed
	)

	return &Worker{
		consumer:    consumer,
		aggregator:  agg,
		metricsCol:  metricsCol,
		traceCol:    traceCol,
		flushTicker: time.NewTicker(5 * time.Minute),
	}
}
func (w *Worker) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	go w.handleShutdown(cancel)

	// Start Kafka consumer
	go w.consumer.Start(ctx)

	log.Println("Worker started...")

	// Flush loop
	for {
		select {
		case <-ctx.Done():
			log.Println("Worker shutting down...")
			w.flush(ctx) // flush remaining data
			return

		case <-w.flushTicker.C:
			w.flush(ctx)
		}
	}
}
func (w *Worker) flush(ctx context.Context) {
	metrics := w.aggregator.Flush()

	if len(metrics) == 0 {
		log.Println("No metrics to flush")
		return
	}

	docs := make([]interface{}, len(metrics))
	for i, m := range metrics {
		docs[i] = m
	}

	_, err := w.metricsCol.InsertMany(ctx, docs)
	if err != nil {
		log.Println("Mongo insert error:", err)
		return
	}

	log.Printf("Flushed %d metrics\n", len(metrics))
}
func (w *Worker) handleShutdown(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-sigChan
	log.Println("Shutdown signal received...")
	cancel()
}
