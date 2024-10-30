//go:build v2
// +build v2

package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type User struct {
	ID   int
	Name string
	Role string
}

type ContextKey string

const UserKey ContextKey = "user"

type AuthService interface {
	GetUserFromToken(token string) (*User, error)
}

type defaultAuthService struct{}

func (s *defaultAuthService) GetUserFromToken(token string) (*User, error) {
	return &User{ID: 1, Name: "Test", Role: "Admin"}, nil
}

func NewAuthService() AuthService {
	return &defaultAuthService{}
}

func AuthMiddleware(authService AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			user, err := authService.GetUserFromToken(token)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(UserKey).(*User)
	fmt.Println(user)
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
