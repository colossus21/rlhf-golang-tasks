//go:build v1
// +build v1

package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockUserService struct {
	shouldError bool
}

func (m *mockUserService) ValidateToken(token string) (int, error) {
	if m.shouldError {
		return 0, errors.New("invalid token")
	}
	return 123, nil
}

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		wantStatus int
		userSvc    UserService
	}{
		{
			name:       "valid token",
			token:      "valid-token",
			wantStatus: http.StatusOK,
			userSvc:    &mockUserService{shouldError: false},
		},
		{
			name:       "missing token",
			token:      "",
			wantStatus: http.StatusUnauthorized,
			userSvc:    &mockUserService{shouldError: false},
		},
		{
			name:       "invalid token",
			token:      "invalid-token",
			wantStatus: http.StatusUnauthorized,
			userSvc:    &mockUserService{shouldError: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(UserHandler)
			middleware := AuthMiddleware(tt.userSvc)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest("GET", "/", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}
			rr := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.wantStatus)
			}
		})
	}
}

func TestTimeoutMiddleware(t *testing.T) {
	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(6 * time.Second):
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
			return
		}
	})

	fastHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name       string
		handler    http.Handler
		timeout    time.Duration
		wantStatus int
	}{
		{
			name:       "request times out",
			handler:    slowHandler,
			timeout:    1 * time.Second,
			wantStatus: http.StatusOK,
		},
		{
			name:       "request completes",
			handler:    fastHandler,
			timeout:    5 * time.Second,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := TimeoutMiddleware(tt.timeout)
			wrappedHandler := middleware(tt.handler)

			req := httptest.NewRequest("GET", "/", nil)
			rr := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.wantStatus)
			}
		})
	}
}
