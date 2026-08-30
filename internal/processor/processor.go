package processor

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"logmesh/internal/kafka"
	"logmesh/internal/model"
	"logmesh/internal/storage"
)

type Config struct {
	Workers      int
	BatchSize    int
	BatchTimeout time.Duration
	MaxRetries   int
}

type Processor struct {
	consumer *kafka.Consumer
	dlq      kafka.Producer
	store    *storage.OpenSearchStore
	logger   *slog.Logger
	cfg      Config
}

func New(consumer *kafka.Consumer, dlq kafka.Producer, store *storage.OpenSearchStore, logger *slog.Logger, cfg Config) *Processor {
	if cfg.Workers <= 0 {
		cfg.Workers = 4
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 500
	}
	if cfg.BatchTimeout <= 0 {
		cfg.BatchTimeout = time.Second
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}
	return &Processor{consumer: consumer, dlq: dlq, store: store, logger: logger, cfg: cfg}
}

func (p *Processor) Run(ctx context.Context) error {
	jobs := make(chan model.LogEvent, p.cfg.BatchSize*p.cfg.Workers)
	var workers sync.WaitGroup

	for i := 0; i < p.cfg.Workers; i++ {
		workers.Add(1)
		go func(workerID int) {
			defer workers.Done()
			p.worker(ctx, workerID, jobs)
		}(i + 1)
	}

	for {
		select {
		case <-ctx.Done():
			close(jobs)
			workers.Wait()
			return ctx.Err()
		default:
			event, err := p.consumer.Fetch(ctx)
			if err != nil {
				if ctx.Err() != nil {
					close(jobs)
					workers.Wait()
					return ctx.Err()
				}
				p.logger.Error("kafka fetch failed", "error", err)
				time.Sleep(500 * time.Millisecond)
				continue
			}
			jobs <- event
		}
	}
}

func (p *Processor) worker(ctx context.Context, workerID int, jobs <-chan model.LogEvent) {
	batch := make([]model.LogEvent, 0, p.cfg.BatchSize)
	timer := time.NewTimer(p.cfg.BatchTimeout)
	defer timer.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := p.writeWithRetry(ctx, batch); err != nil {
			p.logger.Error("batch failed, publishing dlq", "worker", workerID, "error", err, "count", len(batch))
			for _, event := range batch {
				_ = p.dlq.Publish(ctx, event)
			}
		}
		batch = batch[:0]
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case event, ok := <-jobs:
			if !ok {
				flush()
				return
			}
			batch = append(batch, event)
			if len(batch) >= p.cfg.BatchSize {
				flush()
				resetTimer(timer, p.cfg.BatchTimeout)
			}
		case <-timer.C:
			flush()
			resetTimer(timer, p.cfg.BatchTimeout)
		}
	}
}

func (p *Processor) writeWithRetry(ctx context.Context, batch []model.LogEvent) error {
	var lastErr error
	for attempt := 1; attempt <= p.cfg.MaxRetries; attempt++ {
		if err := p.store.BulkIndex(ctx, batch); err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt) * 200 * time.Millisecond)
			continue
		}
		return nil
	}
	return lastErr
}

func resetTimer(timer *time.Timer, timeout time.Duration) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(timeout)
}
