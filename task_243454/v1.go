//go:build v1
// +build v1

package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type ContextKey string

const UserIDKey ContextKey = "userID"

type UserService interface {
	ValidateToken(token string) (int, error)
}

type defaultUserService struct{}

func (s *defaultUserService) ValidateToken(token string) (int, error) {
	return 1, nil
}

func NewUserService() UserService {
	return &defaultUserService{}
}

func AuthMiddleware(userService UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := userService.ValidateToken(token)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
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
	userID, ok := r.Context().Value(UserIDKey).(int)
	fmt.Println(userID)
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
