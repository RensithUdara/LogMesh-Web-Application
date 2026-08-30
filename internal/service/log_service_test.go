package service

import (
	"context"
	"testing"

	"logmesh/internal/model"
)

func TestIngestDefaultsTimestampEnvironmentAndMasksSensitiveMetadata(t *testing.T) {
	svc := NewInMemoryLogService(10)

	event, err := svc.Ingest(context.Background(), model.IngestLogRequest{
		Service: "api",
		Level:   model.LevelInfo,
		Message: "user logged in",
		Metadata: map[string]interface{}{
			"password": "secret",
			"safe":     "kept",
		},
	})
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}

	if event.ID == "" {
		t.Fatal("expected id to be assigned")
	}
	if event.Environment != "development" {
		t.Fatalf("expected default environment, got %q", event.Environment)
	}
	if event.Metadata["password"] != "[REDACTED]" {
		t.Fatalf("expected password to be redacted, got %v", event.Metadata["password"])
	}
	if event.Metadata["safe"] != "kept" {
		t.Fatalf("expected safe metadata to remain, got %v", event.Metadata["safe"])
	}
}

func TestSearchFiltersByLevelAndMessage(t *testing.T) {
	svc := NewInMemoryLogService(10)

	_, _ = svc.Ingest(context.Background(), model.IngestLogRequest{
		Service: "payments",
		Level:   model.LevelError,
		Message: "database timeout",
	})
	_, _ = svc.Ingest(context.Background(), model.IngestLogRequest{
		Service: "payments",
		Level:   model.LevelInfo,
		Message: "payment completed",
	})

	result, err := svc.Search(context.Background(), model.SearchLogsQuery{
		Level:  model.LevelError,
		Search: "database",
	})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if result.Total != 1 {
		t.Fatalf("expected one match, got %d", result.Total)
	}
	if result.Logs[0].Message != "database timeout" {
		t.Fatalf("unexpected result: %q", result.Logs[0].Message)
	}
}

func TestInMemoryStoreHonorsMaxItems(t *testing.T) {
	svc := NewInMemoryLogService(2)

	_, _ = svc.Ingest(context.Background(), model.IngestLogRequest{Service: "api", Level: model.LevelInfo, Message: "one"})
	_, _ = svc.Ingest(context.Background(), model.IngestLogRequest{Service: "api", Level: model.LevelInfo, Message: "two"})
	_, _ = svc.Ingest(context.Background(), model.IngestLogRequest{Service: "api", Level: model.LevelInfo, Message: "three"})

	result, err := svc.Search(context.Background(), model.SearchLogsQuery{})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if result.Total != 2 {
		t.Fatalf("expected two retained logs, got %d", result.Total)
	}
	if result.Logs[0].Message != "three" || result.Logs[1].Message != "two" {
		t.Fatalf("expected newest retained logs, got %#v", result.Logs)
	}
}
