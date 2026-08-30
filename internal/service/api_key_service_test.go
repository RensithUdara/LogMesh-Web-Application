package service

import (
	"context"
	"testing"
)

func TestAPIKeyCreateListAuthenticateAndRevoke(t *testing.T) {
	svc := NewInMemoryAPIKeyService()

	created, err := svc.Create(context.Background(), "Production collector")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.PlaintextKey == "" {
		t.Fatal("expected plaintext key to be returned once")
	}

	if _, ok := svc.Authenticate(context.Background(), created.PlaintextKey); !ok {
		t.Fatal("expected key to authenticate")
	}

	keys, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	for _, key := range keys {
		if key.PlaintextKey != "" {
			t.Fatal("list must not return plaintext api keys")
		}
	}

	if err := svc.Revoke(context.Background(), created.ID); err != nil {
		t.Fatalf("revoke failed: %v", err)
	}
	if _, ok := svc.Authenticate(context.Background(), created.PlaintextKey); ok {
		t.Fatal("revoked key should not authenticate")
	}
}
