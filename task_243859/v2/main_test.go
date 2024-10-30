// main_test.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

type TestRunner struct {
	currentTest int
	passed      int
	failed      int
	errors      []string
}

func NewTestRunner() *TestRunner {
	return &TestRunner{
		currentTest: 0,
		passed:      0,
		failed:      0,
		errors:      []string{},
	}
}

func (tr *TestRunner) Run(name string, testFunc func() error) {
	tr.currentTest++

	// Defer panic recovery
	defer func() {
		if r := recover(); r != nil {
			tr.failed++
			tr.errors = append(tr.errors, fmt.Sprintf("Test %d# Failed (PANIC) - %s: %v", tr.currentTest, name, r))
			fmt.Printf("Test %d# Failed (PANIC) - %s: %v\n", tr.currentTest, name, r)
		}
	}()

	// Run the test
	err := testFunc()
	if err != nil {
		tr.failed++
		tr.errors = append(tr.errors, fmt.Sprintf("Test %d# Failed - %s: %s", tr.currentTest, name, err))
		fmt.Printf("Test %d# Failed - %s: %s\n", tr.currentTest, name, err)
	} else {
		tr.passed++
		fmt.Printf("Test %d# Passed - %s\n", tr.currentTest, name)
	}
}

func (tr *TestRunner) Summary() {
	fmt.Printf("\nTest Summary:\nTotal: %d\nPassed: %d\nFailed: %d\n",
		tr.currentTest, tr.passed, tr.failed)
}

func setupTestApp(t *testing.T) (*Application, *httptest.Server) {
	gin.SetMode(gin.TestMode)

	// Create mock database
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}

	// Setup router with session middleware
	router := gin.New()
	store := cookie.NewStore([]byte("test-secret-key"))
	router.Use(sessions.Sessions("test-session", store))

	app := &Application{
		DB:      db,
		Router:  router,
		UserSvc: NewMockUserService(),
	}

	app.setupRoutes()

	server := httptest.NewServer(app.Router)
	return app, server
}

func TestGinCRUD(t *testing.T) {
	runner := NewTestRunner()
	app, server := setupTestApp(t)
	defer server.Close()

	// Test 1: Create new user with valid data
	runner.Run("Create User with Valid Data", func() error {
		payload := map[string]string{
			"username": "testuser",
			"password": "password123",
			"email":    "test@example.com",
		}
		jsonData, _ := json.Marshal(payload)
		w := performRequest(app.Router, "POST", "/register", bytes.NewBuffer(jsonData))

		if w.Code != http.StatusCreated {
			return fmt.Errorf("expected status 201, got %d", w.Code)
		}
		return nil
	})

	// Test 2: Create user with duplicate username
	runner.Run("Create User with Duplicate Username", func() error {
		payload := map[string]string{
			"username": "testuser",
			"password": "password123",
			"email":    "another@example.com",
		}
		jsonData, _ := json.Marshal(payload)
		w := performRequest(app.Router, "POST", "/register", bytes.NewBuffer(jsonData))

		if w.Code != http.StatusConflict {
			return fmt.Errorf("expected status 409, got %d", w.Code)
		}
		return nil
	})

	// Test 3: Create user with missing required fields
	runner.Run("Create User with Missing Fields", func() error {
		payload := map[string]string{
			"username": "testuser",
			// missing password and email
		}
		jsonData, _ := json.Marshal(payload)
		w := performRequest(app.Router, "POST", "/register", bytes.NewBuffer(jsonData))

		if w.Code != http.StatusBadRequest {
			return fmt.Errorf("expected status 400, got %d", w.Code)
		}
		return nil
	})

	// Test 4: Login with valid credentials
	runner.Run("Login with Valid Credentials", func() error {
		payload := map[string]string{
			"username": "testuser",
			"password": "password123",
		}
		jsonData, _ := json.Marshal(payload)
		w := performRequest(app.Router, "POST", "/login", bytes.NewBuffer(jsonData))

		if w.Code != http.StatusOK {
			return fmt.Errorf("expected status 200, got %d", w.Code)
		}
		return nil
	})

	// Test 5: Login with invalid password
	runner.Run("Login with Invalid Password", func() error {
		payload := map[string]string{
			"username": "testuser",
			"password": "wrongpassword",
		}
		jsonData, _ := json.Marshal(payload)
		w := performRequest(app.Router, "POST", "/login", bytes.NewBuffer(jsonData))

		if w.Code != http.StatusUnauthorized {
			return fmt.Errorf("expected status 401, got %d", w.Code)
		}
		return nil
	})

	// Test 6: Get user details with valid ID
	runner.Run("Get User Details with Valid ID", func() error {
		w := performRequestWithAuth(app.Router, "GET", "/users/1", nil)

		if w.Code != http.StatusOK {
			return fmt.Errorf("expected status 200, got %d", w.Code)
		}
		return nil
	})

	// Test 7: Get user details with non-existent ID
	runner.Run("Get User Details with Non-existent ID", func() error {
		w := performRequestWithAuth(app.Router, "GET", "/users/999", nil)

		if w.Code != http.StatusNotFound {
			return fmt.Errorf("expected status 404, got %d", w.Code)
		}
		return nil
	})

	// Test 8: Update user with valid data
	runner.Run("Update User with Valid Data", func() error {
		payload := map[string]string{
			"username": "updateduser",
			"email":    "updated@example.com",
		}
		jsonData, _ := json.Marshal(payload)
		w := performRequestWithAuth(app.Router, "PUT", "/users/1", bytes.NewBuffer(jsonData))

		if w.Code != http.StatusOK {
			return fmt.Errorf("expected status 200, got %d", w.Code)
		}
		return nil
	})

	// Test 9: Update user without authentication
	runner.Run("Update User without Authentication", func() error {
		payload := map[string]string{
			"username": "updateduser",
			"email":    "updated@example.com",
		}
		jsonData, _ := json.Marshal(payload)
		w := performRequest(app.Router, "PUT", "/users/1", bytes.NewBuffer(jsonData))

		if w.Code != http.StatusUnauthorized {
			return fmt.Errorf("expected status 401, got %d", w.Code)
		}
		return nil
	})

	// Test 10: Delete user with valid ID
	runner.Run("Delete User with Valid ID", func() error {
		w := performRequestWithAuth(app.Router, "DELETE", "/users/1", nil)

		if w.Code != http.StatusOK {
			return fmt.Errorf("expected status 200, got %d", w.Code)
		}
		return nil
	})

	// Test 11: Get users list with authentication
	runner.Run("Get Users List with Authentication", func() error {
		w := performRequestWithAuth(app.Router, "GET", "/users", nil)

		if w.Code != http.StatusOK {
			return fmt.Errorf("expected status 200, got %d", w.Code)
		}
		return nil
	})

	// Test 12: Create user with invalid email format
	runner.Run("Create User with Invalid Email Format", func() error {
		payload := map[string]string{
			"username": "newuser",
			"password": "password123",
			"email":    "invalid-email",
		}
		jsonData, _ := json.Marshal(payload)
		w := performRequest(app.Router, "POST", "/register", bytes.NewBuffer(jsonData))

		if w.Code != http.StatusBadRequest {
			return fmt.Errorf("expected status 400, got %d", w.Code)
		}
		return nil
	})

	runner.Summary()
}

// Helper function to perform request without authentication
func performRequest(r http.Handler, method, path string, body *bytes.Buffer) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, body)
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func createTestContextWithSession() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Initialize session store
	store := cookie.NewStore([]byte("test-secret-key"))
	sessions.Sessions("test-session", store)(c)

	return c, w
}

// Helper function to perform request with authentication
func performRequestWithAuth(r *gin.Engine, method, path string, body *bytes.Buffer) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, body)
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	// Create test context with session
	c, w := createTestContextWithSession()

	// Set authentication in session
	session := sessions.Default(c)
	session.Set("authenticated", true)
	session.Set("user_id", 1)
	session.Save()

	// Copy request details to context
	c.Request = req

	r.ServeHTTP(w, req)
	return w
}
