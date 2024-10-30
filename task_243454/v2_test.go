//go:build v2
// +build v2

package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockAuthService struct {
	shouldError bool
}

func (m *mockAuthService) GetUserFromToken(token string) (*User, error) {
	if m.shouldError {
		return nil, errors.New("invalid token")
	}
	return &User{ID: 123, Name: "Test User", Role: "Admin"}, nil
}

func TestAuthMiddlewareWithUser(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		wantStatus int
		authSvc    AuthService
	}{
		{
			name:       "valid token returns user",
			token:      "valid-token",
			wantStatus: http.StatusOK,
			authSvc:    &mockAuthService{shouldError: false},
		},
		{
			name:       "missing token",
			token:      "",
			wantStatus: http.StatusUnauthorized,
			authSvc:    &mockAuthService{shouldError: false},
		},
		{
			name:       "invalid token",
			token:      "invalid-token",
			wantStatus: http.StatusUnauthorized,
			authSvc:    &mockAuthService{shouldError: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(UserHandler)
			middleware := AuthMiddleware(tt.authSvc)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest("GET", "/", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}
			rr := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.wantStatus)
			}
		})
	}
}

func TestContextPropagation(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(UserKey).(*User)
		if !ok {
			t.Error("user not found in context")
		}
		if user.ID != 123 {
			t.Errorf("got wrong user ID: got %v want %v", user.ID, 123)
		}
		w.WriteHeader(http.StatusOK)
	})

	authSvc := &mockAuthService{shouldError: false}
	middleware := AuthMiddleware(authSvc)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "valid-token")
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
