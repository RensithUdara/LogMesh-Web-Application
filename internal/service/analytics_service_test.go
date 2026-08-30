package service

import (
	"context"
	"testing"

	"logmesh/internal/model"
)

func TestAnalyticsSummaryCountsLogs(t *testing.T) {
	logs := NewInMemoryLogService(10)
	analytics := NewAnalyticsService(logs)

	_, _ = logs.Ingest(context.Background(), model.IngestLogRequest{
		Service:     "payments",
		Environment: "production",
		Level:       model.LevelError,
		Message:     "payment failed",
		Host:        "server-01",
	})
	_, _ = logs.Ingest(context.Background(), model.IngestLogRequest{
		Service:     "payments",
		Environment: "production",
		Level:       model.LevelWarn,
		Message:     "slow gateway",
		Host:        "server-02",
	})

	summary, err := analytics.Summary(context.Background())
	if err != nil {
		t.Fatalf("summary failed: %v", err)
	}

	if summary.Total != 2 || summary.Errors != 1 || summary.Warnings != 1 {
		t.Fatalf("unexpected summary counts: %#v", summary)
	}
	if summary.ServiceCount != 1 {
		t.Fatalf("expected one service, got %d", summary.ServiceCount)
	}

	sources, err := analytics.Sources(context.Background())
	if err != nil {
		t.Fatalf("sources failed: %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("expected one source, got %d", len(sources))
	}
	if sources[0].HostCount != 2 {
		t.Fatalf("expected two hosts, got %d", sources[0].HostCount)
	}
}
