package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"logmesh/internal/config"
	"logmesh/internal/kafka"
	"logmesh/internal/processor"
	"logmesh/internal/storage"
)

func main() {
	config.LoadDotEnv(".env")
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))

	consumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaLogsTopic, "logmesh-processors")
	if consumer == nil {
		logger.Error("LOGMESH_KAFKA_BROKERS is required for processor")
		os.Exit(1)
	}
	defer consumer.Close()

	store, err := storage.NewOpenSearchStore(cfg.OpenSearchURL)
	if err != nil {
		logger.Error("opensearch setup failed", "error", err)
		os.Exit(1)
	}
	if store == nil {
		logger.Error("LOGMESH_OPENSEARCH_URL is required for processor")
		os.Exit(1)
	}

	dlq := kafka.NewProducer(cfg.KafkaBrokers, cfg.KafkaDLQTopic)
	defer dlq.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	runner := processor.New(consumer, dlq, store, logger, processor.Config{
		Workers:      cfg.ProcessorWorkers,
		BatchSize:    cfg.ProcessorBatchSize,
		BatchTimeout: time.Duration(cfg.ProcessorBatchMS) * time.Millisecond,
		MaxRetries:   3,
	})

	logger.Info("starting processor", "workers", cfg.ProcessorWorkers, "batch_size", cfg.ProcessorBatchSize)
	if err := runner.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("processor stopped with error", "error", err)
		os.Exit(1)
	}
	logger.Info("processor stopped")
}
