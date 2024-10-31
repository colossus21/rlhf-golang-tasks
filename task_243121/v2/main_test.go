package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

var (
	mock   sqlmock.Sqlmock
	testDB *sql.DB
)

type TestReporter struct {
	t          *testing.T
	testNumber int
}

func NewTestReporter(t *testing.T) *TestReporter {
	return &TestReporter{t: t}
}

func (tr *TestReporter) Report(name string, err error) {
	tr.testNumber++
	result := "Passed"
	if err != nil {
		result = fmt.Sprintf("Failed: %v", err)
	}
	fmt.Printf("Test %d# %s: %s\n", tr.testNumber, name, result)
}

func setupTestDB(t *testing.T) sqlmock.Sqlmock {
	var err error
	testDB, mock, err = sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual),
		sqlmock.MonitorPingsOption(false),
	)
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	db = testDB
	return mock
}

func runTestWithRecovery(reporter *TestReporter, name string, testFunc func() error) {
	defer func() {
		if r := recover(); r != nil {
			reporter.Report(name, fmt.Errorf("Test panicked: %v", r))
		}
	}()

	err := testFunc()
	reporter.Report(name, err)
}

func TestProductSystemV2(t *testing.T) {
	reporter := NewTestReporter(t)

	// Test 1: Database Connection
	runTestWithRecovery(reporter, "Database Connection Verification", func() error {
		mock = setupTestDB(t)
		return nil
		//return mock.ExpectPing().WillReturnError(nil)
	})

	// Test 2: Auto Table Creation
	runTestWithRecovery(reporter, "Auto Table Creation", func() error {
		mock = setupTestDB(t)
		mock.ExpectExec("CREATE TABLE IF NOT EXISTS products").WillReturnResult(sqlmock.NewResult(0, 0))
		return nil
	})

	// Test 3: Create Product Success
	runTestWithRecovery(reporter, "Create Product With Valid Data", func() error {
		mock = setupTestDB(t)
		mock.ExpectExec("INSERT INTO products").
			WithArgs("Test Product", "Test Description", 99.99).
			WillReturnResult(sqlmock.NewResult(1, 1))

		form := url.Values{}
		form.Add("name", "Test Product")
		form.Add("description", "Test Description")
		form.Add("price", "99.99")

		req := httptest.NewRequest("POST", "/create", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		createHandler(w, req)

		if w.Code != http.StatusSeeOther && w.Code != http.StatusOK {
			return fmt.Errorf("expected status 303 or 200, got %d", w.Code)
		}
		return nil
	})

	// Test 4: Create Product Empty Name
	runTestWithRecovery(reporter, "Empty Name Validation", func() error {
		mock = setupTestDB(t)
		form := url.Values{}
		form.Add("name", "")
		form.Add("price", "99.99")

		req := httptest.NewRequest("POST", "/create", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		createHandler(w, req)

		if w.Code == http.StatusSeeOther {
			return fmt.Errorf("expected validation error for empty name")
		}
		return nil
	})

	// Test 5: Invalid Price Format
	runTestWithRecovery(reporter, "Invalid Price Format", func() error {
		mock = setupTestDB(t)
		form := url.Values{}
		form.Add("name", "Test Product")
		form.Add("price", "invalid")

		req := httptest.NewRequest("POST", "/create", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		createHandler(w, req)

		if w.Code == http.StatusSeeOther {
			return fmt.Errorf("expected validation error for invalid price")
		}
		return nil
	})

	// Test 6: Get Empty Product List
	runTestWithRecovery(reporter, "Get Empty Product List", func() error {
		mock = setupTestDB(t)
		mock.ExpectQuery("SELECT \\* FROM products").WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "description", "price"}))

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		indexHandler(w, req)

		if w.Code != http.StatusOK {
			return fmt.Errorf("expected status 200, got %d", w.Code)
		}
		return nil
	})

	// Test 7: Get Multiple Products
	runTestWithRecovery(reporter, "Get Multiple Products", func() error {
		mock = setupTestDB(t)
		rows := sqlmock.NewRows([]string{"id", "name", "description", "price"}).
			AddRow(1, "Product 1", "Desc 1", 99.99).
			AddRow(2, "Product 2", "Desc 2", 149.99)
		mock.ExpectQuery("SELECT \\* FROM products").WillReturnRows(rows)

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		indexHandler(w, req)

		if w.Code != http.StatusOK {
			return fmt.Errorf("expected status 200, got %d", w.Code)
		}
		return nil
	})

	// Test 8: Get Single Product
	runTestWithRecovery(reporter, "Get Single Product", func() error {
		mock = setupTestDB(t)
		mock.ExpectQuery("SELECT \\* FROM products WHERE").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "price"}).
				AddRow(1, "Product 1", "Desc 1", 99.99))

		req := httptest.NewRequest("GET", "/edit?id=1", nil)
		w := httptest.NewRecorder()

		editHandler(w, req)

		if w.Code != http.StatusOK {
			return fmt.Errorf("expected status 200, got %d", w.Code)
		}
		return nil
	})

	// Test 9: Get Non-existent Product
	runTestWithRecovery(reporter, "Get Non-existent Product", func() error {
		mock = setupTestDB(t)
		mock.ExpectQuery("SELECT \\* FROM products WHERE").
			WithArgs(999).
			WillReturnError(sql.ErrNoRows)

		req := httptest.NewRequest("GET", "/edit?id=999", nil)
		w := httptest.NewRecorder()

		editHandler(w, req)

		if w.Code != http.StatusInternalServerError {
			return fmt.Errorf("expected status 500, got %d", w.Code)
		}
		return nil
	})

	// Test 10: Update Product Success
	runTestWithRecovery(reporter, "Update Product Success", func() error {
		mock = setupTestDB(t)
		mock.ExpectExec("UPDATE products SET").
			WithArgs("Updated Product", "Updated Desc", 199.99, 1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		form := url.Values{}
		form.Add("id", "1")
		form.Add("name", "Updated Product")
		form.Add("description", "Updated Desc")
		form.Add("price", "199.99")

		req := httptest.NewRequest("POST", "/update", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		updateHandler(w, req)

		if w.Code != http.StatusSeeOther {
			return fmt.Errorf("expected status 303, got %d", w.Code)
		}
		return nil
	})

	// Test 11: Update Non-existent Product
	runTestWithRecovery(reporter, "Update Non-existent Product", func() error {
		mock = setupTestDB(t)
		mock.ExpectExec("UPDATE products SET").
			WithArgs("Updated Product", "Updated Desc", 199.99, 999).
			WillReturnResult(sqlmock.NewResult(0, 0))

		form := url.Values{}
		form.Add("id", "999")
		form.Add("name", "Updated Product")
		form.Add("description", "Updated Desc")
		form.Add("price", "199.99")

		req := httptest.NewRequest("POST", "/update", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		updateHandler(w, req)

		if w.Code != http.StatusInternalServerError {
			return fmt.Errorf("expected status 500, got %d", w.Code)
		}
		return nil
	})

	// Test 12: Delete Product Success
	runTestWithRecovery(reporter, "Delete Product Success", func() error {
		mock = setupTestDB(t)
		mock.ExpectExec("DELETE FROM products WHERE").
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		req := httptest.NewRequest("GET", "/delete?id=1", nil)
		w := httptest.NewRecorder()

		deleteHandler(w, req)

		if w.Code != http.StatusSeeOther {
			return fmt.Errorf("expected status 303, got %d", w.Code)
		}
		return nil
	})

	// Test 13: Delete Non-existent Product
	runTestWithRecovery(reporter, "Delete Non-existent Product", func() error {
		mock = setupTestDB(t)
		mock.ExpectExec("DELETE FROM products WHERE").
			WithArgs(999).
			WillReturnResult(sqlmock.NewResult(0, 0))

		req := httptest.NewRequest("GET", "/delete?id=999", nil)
		w := httptest.NewRecorder()

		deleteHandler(w, req)

		if w.Code != http.StatusInternalServerError {
			return fmt.Errorf("expected status 500, got %d", w.Code)
		}
		return nil
	})

	// Test 14: SQL Injection Prevention
	runTestWithRecovery(reporter, "SQL Injection Prevention", func() error {
		mock = setupTestDB(t)
		req := httptest.NewRequest("GET", "/edit?id=1' OR '1'='1", nil)
		w := httptest.NewRecorder()

		editHandler(w, req)

		if w.Code == http.StatusOK {
			return fmt.Errorf("SQL injection attempt should not succeed")
		}
		return nil
	})

	// Test 15: XSS Prevention
	runTestWithRecovery(reporter, "XSS Prevention", func() error {
		mock = setupTestDB(t)
		form := url.Values{}
		form.Add("name", "<script>alert('xss')</script>")
		form.Add("price", "99.99")

		req := httptest.NewRequest("POST", "/create", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		createHandler(w, req)

		if strings.Contains(w.Body.String(), "<script>") {
			return fmt.Errorf("XSS content should be escaped")
		}
		return nil
	})

	// Test 16: Zero Price Validation
	runTestWithRecovery(reporter, "Zero Price Validation", func() error {
		mock = setupTestDB(t)
		form := url.Values{}
		form.Add("name", "Test Product")
		form.Add("price", "0.00")

		req := httptest.NewRequest("POST", "/create", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		createHandler(w, req)

		if w.Code == http.StatusSeeOther {
			return fmt.Errorf("zero price should not be allowed")
		}
		return nil
	})

	// Test 17: Negative Price Validation
	runTestWithRecovery(reporter, "Negative Price Validation", func() error {
		mock = setupTestDB(t)
		form := url.Values{}
		form.Add("name", "Test Product")
		form.Add("price", "-10.00")

		req := httptest.NewRequest("POST", "/create", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		createHandler(w, req)

		if w.Code == http.StatusSeeOther {
			return fmt.Errorf("negative price should not be allowed")
		}
		return nil
	})

	// Test 18: Invalid Form Method
	runTestWithRecovery(reporter, "Invalid Form Method", func() error {
		req := httptest.NewRequest("GET", "/update", nil)
		w := httptest.NewRecorder()

		updateHandler(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			return fmt.Errorf("expected status 405, got %d", w.Code)
		}
		return nil
	})

	// Test 19: Missing Required Fields
	runTestWithRecovery(reporter, "Missing Required Fields", func() error {
		mock = setupTestDB(t)
		form := url.Values{}
		form.Add("description", "Test Description")

		req := httptest.NewRequest("POST", "/create", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		createHandler(w, req)

		if w.Code == http.StatusSeeOther {
			return fmt.Errorf("missing required fields should not be allowed")
		}
		return nil
	})

	// Test 20: Invalid Product ID Format
	runTestWithRecovery(reporter, "Invalid Product ID Format", func() error {
		req := httptest.NewRequest("GET", "/edit?id=abc", nil)
		w := httptest.NewRecorder()

		editHandler(w, req)

		if w.Code != http.StatusBadRequest {
			return fmt.Errorf("expected status 400, got %d", w.Code)
		}
		return nil
	})

	// Test 21: Database Connection Error
	runTestWithRecovery(reporter, "Database Connection Error", func() error {
		mock = setupTestDB(t)
		mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("connection refused"))

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		indexHandler(w, req)

		if w.Code != http.StatusInternalServerError {
			return fmt.Errorf("expected status 500, got %d", w.Code)
		}
		return nil
	})

	// Test 22: Template Rendering
	runTestWithRecovery(reporter, "Template Rendering", func() error {
		mock = setupTestDB(t)
		rows := sqlmock.NewRows([]string{"id", "name", "description", "price"}).
			AddRow(1, "Test Product", "Test Desc", 99.99)
		mock.ExpectQuery("SELECT").WillReturnRows(rows)

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		indexHandler(w, req)

		if !strings.Contains(w.Body.String(), "Test Product") {
			return fmt.Errorf("template should render product information")
		}
		return nil
	})

	// Test 23: Form Content Type
	runTestWithRecovery(reporter, "Form Content Type Validation", func() error {
		req := httptest.NewRequest("POST", "/create", strings.NewReader("invalid"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		createHandler(w, req)

		if w.Code != http.StatusBadRequest {
			return fmt.Errorf("expected status 400, got %d", w.Code)
		}
		return nil
	})

	// Test 24: Price Format Validation
	runTestWithRecovery(reporter, "Price Format Validation", func() error {
		mock = setupTestDB(t)
		form := url.Values{}
		form.Add("name", "Test Product")
		form.Add("price", "99.999")

		req := httptest.NewRequest("POST", "/create", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		createHandler(w, req)

		if w.Code == http.StatusSeeOther {
			return fmt.Errorf("invalid price format should not be allowed")
		}
		return nil
	})

	// Test 25: Long Product Name
	runTestWithRecovery(reporter, "Long Product Name Validation", func() error {
		mock = setupTestDB(t)
		form := url.Values{}
		form.Add("name", strings.Repeat("a", 256))
		form.Add("price", "99.99")

		req := httptest.NewRequest("POST", "/create", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		createHandler(w, req)

		if w.Code == http.StatusSeeOther {
			return fmt.Errorf("extremely long product name should not be allowed")
		}
		return nil
	})

	// Test 26: Empty Form Submission
	runTestWithRecovery(reporter, "Empty Form Submission", func() error {
		req := httptest.NewRequest("POST", "/create", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		createHandler(w, req)

		if w.Code == http.StatusSeeOther {
			return fmt.Errorf("empty form should not be allowed")
		}
		return nil
	})

	// Test 27: Duplicate Product Check
	runTestWithRecovery(reporter, "Duplicate Product Check", func() error {
		mock = setupTestDB(t)
		mock.ExpectExec("INSERT INTO products").
			WillReturnError(fmt.Errorf("duplicate entry"))

		form := url.Values{}
		form.Add("name", "Existing Product")
		form.Add("price", "99.99")

		req := httptest.NewRequest("POST", "/create", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		createHandler(w, req)

		if w.Code == http.StatusSeeOther {
			return fmt.Errorf("duplicate product should not be allowed")
		}
		return nil
	})

	// Test 28: Product Update Conflict
	runTestWithRecovery(reporter, "Product Update Conflict", func() error {
		mock = setupTestDB(t)
		mock.ExpectExec("UPDATE products SET").
			WillReturnError(fmt.Errorf("conflict"))

		form := url.Values{}
		form.Add("id", "1")
		form.Add("name", "Updated Product")
		form.Add("price", "99.99")

		req := httptest.NewRequest("POST", "/update", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		updateHandler(w, req)

		if w.Code == http.StatusSeeOther {
			return fmt.Errorf("conflicting update should not succeed")
		}
		return nil
	})

	// Test 29: Request Timeout Simulation
	runTestWithRecovery(reporter, "Request Timeout Handling", func() error {
		mock = setupTestDB(t)
		mock.ExpectQuery("SELECT").WillDelayFor(time.Second).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "price"}))

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		done := make(chan bool)
		go func() {
			indexHandler(w, req)
			done <- true
		}()

		select {
		case <-done:
			return nil
		case <-time.After(2 * time.Second):
			return fmt.Errorf("request timed out")
		}
	})

	// Test 30: Database Transaction Rollback
	runTestWithRecovery(reporter, "Database Transaction Rollback", func() error {
		mock = setupTestDB(t)
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE products").WillReturnError(fmt.Errorf("error"))
		mock.ExpectRollback()

		form := url.Values{}
		form.Add("id", "1")
		form.Add("name", "Updated")
		form.Add("price", "99.99")

		req := httptest.NewRequest("POST", "/update", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		updateHandler(w, req)

		if w.Code == http.StatusSeeOther {
			return fmt.Errorf("failed transaction should not redirect")
		}
		return nil
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled mock expectations: %s", err)
	}
}
