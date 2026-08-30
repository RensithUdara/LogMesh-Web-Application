package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"logmesh/internal/model"
)

var ErrAPIKeyNotFound = errors.New("api key not found")

type APIKeyService interface {
	Create(ctx context.Context, name string) (model.APIKey, error)
	List(ctx context.Context) ([]model.APIKey, error)
	Revoke(ctx context.Context, id string) error
	Authenticate(ctx context.Context, plaintext string) (model.APIKey, bool)
}

type storedAPIKey struct {
	keyHash string
	record  model.APIKey
}

type InMemoryAPIKeyService struct {
	mu   sync.RWMutex
	keys []storedAPIKey
}

func NewInMemoryAPIKeyService() *InMemoryAPIKeyService {
	svc := &InMemoryAPIKeyService{keys: make([]storedAPIKey, 0)}
	_, _ = svc.Create(context.Background(), "Development collector")
	return svc
}

func (s *InMemoryAPIKeyService) Create(_ context.Context, name string) (model.APIKey, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Untitled key"
	}

	plaintext, err := generateAPIKey()
	if err != nil {
		return model.APIKey{}, err
	}

	now := time.Now().UTC()
	record := model.APIKey{
		ID:           uuid.NewString(),
		Name:         name,
		Prefix:       keyPrefix(plaintext),
		CreatedAt:    now,
		PlaintextKey: plaintext,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.keys = append(s.keys, storedAPIKey{
		keyHash: hashAPIKey(plaintext),
		record:  record,
	})

	return record, nil
}

func (s *InMemoryAPIKeyService) List(_ context.Context) ([]model.APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]model.APIKey, 0, len(s.keys))
	for _, item := range s.keys {
		record := item.record
		record.PlaintextKey = ""
		keys = append(keys, record)
	}

	slices.SortFunc(keys, func(a, b model.APIKey) int {
		return b.CreatedAt.Compare(a.CreatedAt)
	})

	return keys, nil
}

func (s *InMemoryAPIKeyService) Revoke(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	for index := range s.keys {
		if s.keys[index].record.ID == id {
			s.keys[index].record.RevokedAt = &now
			return nil
		}
	}

	return ErrAPIKeyNotFound
}

func (s *InMemoryAPIKeyService) Authenticate(_ context.Context, plaintext string) (model.APIKey, bool) {
	plaintext = strings.TrimSpace(plaintext)
	if plaintext == "" {
		return model.APIKey{}, false
	}

	hash := hashAPIKey(plaintext)
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	for index := range s.keys {
		item := &s.keys[index]
		if item.keyHash == hash && item.record.RevokedAt == nil {
			item.record.LastUsedAt = &now
			record := item.record
			record.PlaintextKey = ""
			return record, true
		}
	}

	return model.APIKey{}, false
}

func generateAPIKey() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "lm_live_" + base64.RawURLEncoding.EncodeToString(bytes), nil
}

func hashAPIKey(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func keyPrefix(value string) string {
	if len(value) <= 14 {
		return value
	}
	return value[:14]
}
