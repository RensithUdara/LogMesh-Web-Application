package model

import "time"

type APIKey struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Prefix       string     `json:"prefix"`
	CreatedAt    time.Time  `json:"created_at"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
	PlaintextKey string     `json:"plaintext_key,omitempty"`
}

type CreateAPIKeyRequest struct {
	Name string `json:"name"`
}
