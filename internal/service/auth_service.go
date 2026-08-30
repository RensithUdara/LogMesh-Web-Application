package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"logmesh/internal/model"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	mu        sync.RWMutex
	users     map[string]model.User
	jwtSecret []byte
}

func NewAuthService(jwtSecret string) *AuthService {
	return &AuthService{
		users:     make(map[string]model.User),
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *AuthService) Register(_ context.Context, name, email, password string) (model.AuthResponse, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	if name == "" || email == "" || len(password) < 8 {
		return model.AuthResponse{}, ErrInvalidCredentials
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.AuthResponse{}, err
	}

	user := model.User{
		ID:           uuid.NewString(),
		ProjectID:    uuid.NewString(),
		Name:         name,
		Email:        email,
		PasswordHash: string(passwordHash),
		CreatedAt:    time.Now().UTC(),
	}

	s.mu.Lock()
	s.users[email] = user
	s.mu.Unlock()

	token, err := s.tokenFor(user)
	if err != nil {
		return model.AuthResponse{}, err
	}
	return model.AuthResponse{Token: token, User: user}, nil
}

func (s *AuthService) Login(_ context.Context, email, password string) (model.AuthResponse, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	s.mu.RLock()
	user, ok := s.users[email]
	s.mu.RUnlock()
	if !ok {
		return model.AuthResponse{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return model.AuthResponse{}, ErrInvalidCredentials
	}

	token, err := s.tokenFor(user)
	if err != nil {
		return model.AuthResponse{}, err
	}
	return model.AuthResponse{Token: token, User: user}, nil
}

func (s *AuthService) ParseToken(tokenValue string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenValue, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		return nil, ErrInvalidCredentials
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidCredentials
	}
	return claims, nil
}

func (s *AuthService) tokenFor(user model.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":        user.ID,
		"project_id": user.ProjectID,
		"name":       user.Name,
		"email":      user.Email,
		"exp":        time.Now().Add(24 * time.Hour).Unix(),
		"iat":        time.Now().Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
}
