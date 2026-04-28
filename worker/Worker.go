package worker

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"UpblitIngestor/kafka"
	"UpblitIngestor/services"

	"go.mongodb.org/mongo-driver/mongo"
)

type Worker struct {
	traceConsumer *kafka.TraceConsumer
	logConsumer   *kafka.LogConsumer

	aggregator  *services.MetricsAggregator
	metricsCol  *mongo.Collection
	flushTicker *time.Ticker
}

// Constructor
func NewWorker(
	broker string,
	traceTopic string,
	logTopic string,
	traceGroupID string,
	logGroupID string,
	metricsCol *mongo.Collection,
	traceCol *mongo.Collection,
	logCol *mongo.Collection,
) *Worker {

	agg := services.NewMetricsAggregator()

	traceConsumer := kafka.NewTraceConsumer(
		broker,
		traceTopic,
		traceGroupID,
		agg,
		traceCol,
	)

	logConsumer := kafka.NewLogConsumer(
		broker,
		logTopic,
		logGroupID,
		logCol,
	)

	return &Worker{
		traceConsumer: traceConsumer,
		logConsumer:   logConsumer,
		aggregator:    agg,
		metricsCol:    metricsCol,
		flushTicker:   time.NewTicker(5 * time.Minute),
	}
}

func (w *Worker) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	go w.handleShutdown(cancel)

	// Start both consumers
	go w.traceConsumer.Start(ctx)
	go w.logConsumer.Start(ctx)

	log.Println("Worker started (trace + log consumers)...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker shutting down...")
			w.flush(ctx)
			return

		case <-w.flushTicker.C:
			w.flush(ctx)
		}
	}
}

// Flush aggregated metrics to Mongo
func (w *Worker) flush(ctx context.Context) {
	metrics := w.aggregator.Flush()

	if len(metrics) == 0 {
		log.Println("No metrics to flush")
		return
	}

	dbCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	docs := make([]interface{}, len(metrics))
	for i, m := range metrics {
		docs[i] = m
	}

	_, err := w.metricsCol.InsertMany(dbCtx, docs)
	if err != nil {
		log.Println("Mongo metrics insert error:", err)
		return
	}

	log.Printf("Flushed %d metrics\n", len(metrics))
}

// Handle OS shutdown signals
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
