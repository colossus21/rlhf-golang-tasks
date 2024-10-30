// main_test.go
package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// MockUserService for testing
type MockUserService struct {
	users  map[int]*User
	nextID int
}

func NewMockUserService() *MockUserService {
	return &MockUserService{
		users:  make(map[int]*User),
		nextID: 1,
	}
}

// Implement UserService interface for MockUserService
func (m *MockUserService) Create(user *User) error {
	// Check for duplicate username
	for _, existingUser := range m.users {
		if existingUser.Username == user.Username {
			return ErrDuplicateUsername
		}
	}

	user.ID = m.nextID
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	m.nextID++
	return nil
}

func (m *MockUserService) GetByID(id int) (*User, error) {
	if user, exists := m.users[id]; exists {
		return user, nil
	}
	return nil, ErrUserNotFound
}

func (m *MockUserService) GetByUsername(username string) (*User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

func (m *MockUserService) List() ([]User, error) {
	users := make([]User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, *user)
	}
	return users, nil
}

func (m *MockUserService) Delete(id int) error {
	if _, exists := m.users[id]; !exists {
		return ErrUserNotFound
	}
	delete(m.users, id)
	return nil
}

func (m *MockUserService) Authenticate(username, password string) (*User, error) {
	for _, user := range m.users {
		if user.Username == username && user.Password == password {
			return user, nil
		}
	}
	return nil, ErrInvalidCredentials
}

// Custom errors for testing
var (
	ErrDuplicateUsername  = errors.New("username already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// setupTestDB creates a new mock database
func setupTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}

	// Setup common expectations
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS users").WillReturnResult(sqlmock.NewResult(0, 0))

	return db, mock
}

// setupTestApp creates a new application instance for testing
func setupTestApp(t *testing.T) (*Application, sqlmock.Sqlmock) {
	// Create mock database
	db, mock := setupTestDB(t)

	// Create test store with a fixed key for testing
	store := sessions.NewCookieStore([]byte("test-secret-key"))

	// Create application instance
	app := &Application{
		DB:      db,
		Router:  mux.NewRouter(),
		Store:   store,
		UserSvc: NewMockUserService(),
	}

	// Setup routes
	app.routes()

	return app, mock
}

// setupTestUser creates a test user and returns it
func setupTestUser(t *testing.T, app *Application) *User {
	user := &User{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
	}

	mockSvc := app.UserSvc.(*MockUserService)
	err := mockSvc.Create(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

// setupAuthenticatedRequest creates a new authenticated request for testing
func setupAuthenticatedRequest(t *testing.T, app *Application, method, path string, body []byte) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, bytes.NewBuffer(body))
	res := httptest.NewRecorder()

	// Create and save authenticated session
	session, err := app.Store.New(req, "session-name")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	session.Values["authenticated"] = true
	session.Values["user_id"] = 1 // Test user ID
	err = session.Save(req, res)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	return req, res
}

// clearTestData cleans up test data
func clearTestData(t *testing.T, app *Application) {
	mockSvc := app.UserSvc.(*MockUserService)
	mockSvc.users = make(map[int]*User)
	mockSvc.nextID = 1
}

// Helper function to compare users
func compareUsers(t *testing.T, expected, actual *User) {
	t.Helper()
	if expected.ID != actual.ID {
		t.Errorf("Expected user ID %d, got %d", expected.ID, actual.ID)
	}
	if expected.Username != actual.Username {
		t.Errorf("Expected username %s, got %s", expected.Username, actual.Username)
	}
	if expected.Email != actual.Email {
		t.Errorf("Expected email %s, got %s", expected.Email, actual.Email)
	}
}

// Helper function to check response status
func checkResponseStatus(t *testing.T, expected, actual int) {
	t.Helper()
	if expected != actual {
		t.Errorf("Expected status %d, got %d", expected, actual)
	}
}

// Helper function to check JSON response
func checkJSONResponse(t *testing.T, res *httptest.ResponseRecorder, target interface{}) {
	t.Helper()
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
}

// TestRunner handles test execution and formatting
type TestRunner struct {
	currentTest int
	passed      int
	failed      int
}

func NewTestRunner() *TestRunner {
	return &TestRunner{
		currentTest: 0,
		passed:      0,
		failed:      0,
	}
}

func (tr *TestRunner) Run(name string, testFunc func() error) {
	tr.currentTest++
	err := testFunc()
	if err != nil {
		fmt.Printf("Test %d# Failed - %s: %s\n", tr.currentTest, name, err)
		tr.failed++
	} else {
		fmt.Printf("Test %d# Passed - %s\n", tr.currentTest, name)
		tr.passed++
	}
}

func (tr *TestRunner) Summary() {
	fmt.Printf("\nTest Summary:\nTotal: %d\nPassed: %d\nFailed: %d\n",
		tr.currentTest, tr.passed, tr.failed)
}

func TestCRUDOperations(t *testing.T) {
	runner := NewTestRunner()
	app, _ := setupTestApp(t)

	// Test 1: User Registration
	runner.Run("User Registration with Valid Data", func() error {
		payload := map[string]string{
			"username": "testuser",
			"password": "password123",
			"email":    "test@example.com",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		app.Router.ServeHTTP(res, req)

		if res.Code != http.StatusCreated {
			return fmt.Errorf("expected status 201, got %d", res.Code)
		}
		return nil
	})

	// Test 2: User Registration with Invalid Data
	runner.Run("User Registration with Invalid Data", func() error {
		payload := map[string]string{
			"username": "testuser",
			// Missing required fields
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		app.Router.ServeHTTP(res, req)

		if res.Code != http.StatusBadRequest {
			return fmt.Errorf("expected status 400, got %d", res.Code)
		}
		return nil
	})

	// Test 3: User Login with Valid Credentials
	runner.Run("User Login with Valid Credentials", func() error {
		payload := map[string]string{
			"username": "testuser",
			"password": "password123",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		app.Router.ServeHTTP(res, req)

		if res.Code != http.StatusOK {
			return fmt.Errorf("expected status 200, got %d", res.Code)
		}
		return nil
	})

	// Test 4: User Login with Invalid Credentials
	runner.Run("User Login with Invalid Credentials", func() error {
		payload := map[string]string{
			"username": "testuser",
			"password": "wrongpassword",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		app.Router.ServeHTTP(res, req)

		if res.Code != http.StatusUnauthorized {
			return fmt.Errorf("expected status 401, got %d", res.Code)
		}
		return nil
	})

	// Test 5: List Users without Authentication
	runner.Run("List Users without Authentication", func() error {
		req := httptest.NewRequest("GET", "/users", nil)
		res := httptest.NewRecorder()

		app.Router.ServeHTTP(res, req)

		if res.Code != http.StatusUnauthorized {
			return fmt.Errorf("expected status 401, got %d", res.Code)
		}
		return nil
	})

	// Test 6: List Users with Authentication
	runner.Run("List Users with Authentication", func() error {
		req := httptest.NewRequest("GET", "/users", nil)
		session, _ := app.Store.New(req, "session-name")
		session.Values["authenticated"] = true
		res := httptest.NewRecorder()
		session.Save(req, res)

		app.Router.ServeHTTP(res, req)

		if res.Code != http.StatusOK {
			return fmt.Errorf("expected status 200, got %d", res.Code)
		}
		return nil
	})

	// Test 7: Delete User without Authentication
	runner.Run("Delete User without Authentication", func() error {
		req := httptest.NewRequest("DELETE", "/users/1", nil)
		res := httptest.NewRecorder()

		app.Router.ServeHTTP(res, req)

		if res.Code != http.StatusUnauthorized {
			return fmt.Errorf("expected status 401, got %d", res.Code)
		}
		return nil
	})

	// Test 8: Delete User with Authentication
	runner.Run("Delete User with Authentication", func() error {
		req := httptest.NewRequest("DELETE", "/users/1", nil)
		session, _ := app.Store.New(req, "session-name")
		session.Values["authenticated"] = true
		res := httptest.NewRecorder()
		session.Save(req, res)

		app.Router.ServeHTTP(res, req)

		if res.Code != http.StatusOK {
			return fmt.Errorf("expected status 200, got %d", res.Code)
		}
		return nil
	})

	// Test 9: Delete Non-existent User
	runner.Run("Delete Non-existent User", func() error {
		req := httptest.NewRequest("DELETE", "/users/999", nil)
		session, _ := app.Store.New(req, "session-name")
		session.Values["authenticated"] = true
		res := httptest.NewRecorder()
		session.Save(req, res)

		app.Router.ServeHTTP(res, req)

		if res.Code != http.StatusNotFound {
			return fmt.Errorf("expected status 404, got %d", res.Code)
		}
		return nil
	})

	// Test 10: Create Duplicate User
	runner.Run("Create Duplicate User", func() error {
		payload := map[string]string{
			"username": "testuser", // Same username as Test 1
			"password": "password123",
			"email":    "test2@example.com",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		app.Router.ServeHTTP(res, req)

		if res.Code != http.StatusBadRequest {
			return fmt.Errorf("expected status 400, got %d", res.Code)
		}
		return nil
	})

	// Test 11: Invalid JSON in Request
	runner.Run("Invalid JSON in Request", func() error {
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		app.Router.ServeHTTP(res, req)

		if res.Code != http.StatusBadRequest {
			return fmt.Errorf("expected status 400, got %d", res.Code)
		}
		return nil
	})

	// Test 12: SQL Injection Attempt
	runner.Run("SQL Injection Attempt", func() error {
		payload := map[string]string{
			"username": "admin'; DROP TABLE users; --",
			"password": "password123",
			"email":    "hack@example.com",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		app.Router.ServeHTTP(res, req)

		if res.Code != http.StatusBadRequest {
			return fmt.Errorf("expected status 400, got %d", res.Code)
		}
		return nil
	})

	// Print test summary
	runner.Summary()
}
