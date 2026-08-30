package collector

import (
	"testing"

	"logmesh/internal/model"
)

func TestParseTextLineStructuredPrefix(t *testing.T) {
	req, err := ParseTextLine(model.ParseTextLogRequest{
		Service: "payments",
		Line:    "2026-08-30 10:21:22 ERROR Payment failed",
	})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if req.Level != model.LevelError {
		t.Fatalf("expected ERROR level, got %s", req.Level)
	}
	if req.Message != "Payment failed" {
		t.Fatalf("unexpected message: %q", req.Message)
	}
	if req.Timestamp == nil {
		t.Fatal("expected timestamp")
	}
}

func TestParseTextLinePlainMessage(t *testing.T) {
	req, err := ParseTextLine(model.ParseTextLogRequest{
		Service: "api",
		Line:    "plain log message",
	})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if req.Level != model.LevelInfo {
		t.Fatalf("expected INFO default, got %s", req.Level)
	}
	if req.Message != "plain log message" {
		t.Fatalf("unexpected message: %q", req.Message)
	}
}
