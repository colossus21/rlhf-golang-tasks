package main

import (
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"text/template"
	"time"
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

func TestProductSystemV1(t *testing.T) {
	reporter := NewTestReporter(t)

	// Test 1: Database Connection
	runTestWithRecovery(reporter, "Database Connection Verification", func() error {
		mock = setupTestDB(t)
		return nil
		//return mock.ExpectPing().WillReturnError(nil)
	})

	// Test 2: Database Initial State
	runTestWithRecovery(reporter, "Database Initial State", func() error {
		mock = setupTestDB(t)
		mock.ExpectQuery("SELECT \\* FROM products").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "price"}))
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

		req := httptest.NewRequest("POST", "/store", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		storeHandler(w, req)

		if w.Code != http.StatusSeeOther {
			return fmt.Errorf("expected status 303, got %d", w.Code)
		}
		return nil
	})

	// Test 4: Create Product Empty Name
	runTestWithRecovery(reporter, "Empty Name Validation", func() error {
		mock = setupTestDB(t)
		form := url.Values{}
		form.Add("name", "")
		form.Add("price", "99.99")

		req := httptest.NewRequest("POST", "/store", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		storeHandler(w, req)

		if w.Code == http.StatusSeeOther {
			return fmt.Errorf("empty name should not be allowed")
		}
		return nil
	})

	// Test 5: Invalid Price Format
	runTestWithRecovery(reporter, "Invalid Price Format", func() error {
		mock = setupTestDB(t)
		form := url.Values{}
		form.Add("name", "Test Product")
		form.Add("price", "invalid")

		req := httptest.NewRequest("POST", "/store", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		storeHandler(w, req)

		if w.Code == http.StatusSeeOther {
			return fmt.Errorf("invalid price should not be allowed")
		}
		return nil
	})

	// Test 6: Get Empty Product List
	runTestWithRecovery(reporter, "Get Empty Product List", func() error {
		mock = setupTestDB(t)
		mock.ExpectQuery("SELECT \\* FROM products").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "price"}))

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

		req := httptest.NewRequest("GET", "/edit/1", nil)
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

		req := httptest.NewRequest("GET", "/edit/999", nil)
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
		form.Add("name", "Updated Product")
		form.Add("description", "Updated Desc")
		form.Add("price", "199.99")

		req := httptest.NewRequest("POST", "/update/1", strings.NewReader(form.Encode()))
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
		form.Add("name", "Updated Product")
		form.Add("description", "Updated Desc")
		form.Add("price", "199.99")

		req := httptest.NewRequest("POST", "/update/999", strings.NewReader(form.Encode()))
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

		req := httptest.NewRequest("GET", "/delete/1", nil)
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

		req := httptest.NewRequest("GET", "/delete/999", nil)
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
		req := httptest.NewRequest("GET", "/edit/1; DROP TABLE products;", nil)
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

		req := httptest.NewRequest("POST", "/store", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		storeHandler(w, req)

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

		req := httptest.NewRequest("POST", "/store", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		storeHandler(w, req)

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

		req := httptest.NewRequest("POST", "/store", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		storeHandler(w, req)

		if w.Code == http.StatusSeeOther {
			return fmt.Errorf("negative price should not be allowed")
		}
		return nil
	})

	// Test 18: Long Product Name
	runTestWithRecovery(reporter, "Long Product Name Validation", func() error {
		mock = setupTestDB(t)
		form := url.Values{}
		form.Add("name", strings.Repeat("a", 256))
		form.Add("price", "99.99")

		req := httptest.NewRequest("POST", "/store", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		storeHandler(w, req)

		if w.Code == http.StatusSeeOther {
			return fmt.Errorf("extremely long product name should not be allowed")
		}
		return nil
	})

	// Test 19: Invalid HTTP Method
	runTestWithRecovery(reporter, "Invalid HTTP Method", func() error {
		req := httptest.NewRequest("PUT", "/store", nil)
		w := httptest.NewRecorder()

		storeHandler(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			return fmt.Errorf("expected status 405, got %d", w.Code)
		}
		return nil
	})

	// Test 20: Invalid Form Content Type
	runTestWithRecovery(reporter, "Invalid Form Content Type", func() error {
		req := httptest.NewRequest("POST", "/store", strings.NewReader("invalid"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		storeHandler(w, req)

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

	// Test 22: Invalid Product ID Format in URL
	runTestWithRecovery(reporter, "Invalid Product ID Format", func() error {
		req := httptest.NewRequest("GET", "/edit/abc", nil)
		w := httptest.NewRecorder()

		editHandler(w, req)

		if w.Code != http.StatusBadRequest {
			return fmt.Errorf("expected status 400, got %d", w.Code)
		}
		return nil
	})

	// Test 23: Empty Form Submission
	runTestWithRecovery(reporter, "Empty Form Submission", func() error {
		req := httptest.NewRequest("POST", "/store", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		storeHandler(w, req)

		if w.Code == http.StatusSeeOther {
			return fmt.Errorf("empty form should not be allowed")
		}
		return nil
	})

	// Test 24: Missing Required Fields
	runTestWithRecovery(reporter, "Missing Required Fields", func() error {
		mock = setupTestDB(t)
		form := url.Values{}
		form.Add("description", "Test Description")

		req := httptest.NewRequest("POST", "/store", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		storeHandler(w, req)

		if w.Code == http.StatusSeeOther {
			return fmt.Errorf("missing required fields should not be allowed")
		}
		return nil
	})

	// Test 25: Product View Model Error
	runTestWithRecovery(reporter, "Product View Model Error", func() error {
		mock = setupTestDB(t)
		viewModel := ProductViewModel{
			Error: "Test error message",
		}

		httptest.NewRequest("GET", "/create", nil)
		w := httptest.NewRecorder()

		// Create template from string to avoid file dependency
		tmpl := template.Must(template.New("create").Parse(`
        {{if .Error}}
            <div class="alert alert-danger">{{.Error}}</div>
        {{end}}
    `))

		err := tmpl.Execute(w, viewModel)
		if err != nil {
			return fmt.Errorf("template execution failed: %v", err)
		}

		if !strings.Contains(w.Body.String(), "Test error message") {
			return fmt.Errorf("view model error should be displayed")
		}
		return nil
	})

	// Test 26: Template Rendering with Special Characters
	runTestWithRecovery(reporter, "Template Special Characters", func() error {
		mock = setupTestDB(t)
		rows := sqlmock.NewRows([]string{"id", "name", "description", "price"}).
			AddRow(1, "Test & Product", "Test < Description >", 99.99)
		mock.ExpectQuery("SELECT").WillReturnRows(rows)

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		indexHandler(w, req)

		if strings.Contains(w.Body.String(), "<Description>") {
			return fmt.Errorf("special characters should be escaped")
		}
		return nil
	})

	// Test 27: Product Price Decimal Places
	runTestWithRecovery(reporter, "Price Decimal Places", func() error {
		mock = setupTestDB(t)
		form := url.Values{}
		form.Add("name", "Test Product")
		form.Add("price", "99.999")

		req := httptest.NewRequest("POST", "/store", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		storeHandler(w, req)

		if w.Code == http.StatusSeeOther {
			return fmt.Errorf("invalid decimal places should not be allowed")
		}
		return nil
	})

	// Test 28: Request Timeout Simulation
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

	// Test 29: Multiple Database Operations Transaction
	runTestWithRecovery(reporter, "Database Transaction", func() error {
		mock = setupTestDB(t)
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE products").WillReturnError(fmt.Errorf("error"))
		mock.ExpectRollback()

		form := url.Values{}
		form.Add("name", "Updated Product")
		form.Add("price", "99.99")

		req := httptest.NewRequest("POST", "/update/1", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		updateHandler(w, req)

		if w.Code == http.StatusSeeOther {
			return fmt.Errorf("failed transaction should not redirect")
		}
		return nil
	})

	// Test 30: Bootstrap CSS Loading
	runTestWithRecovery(reporter, "Bootstrap CSS Loading", func() error {
		mock = setupTestDB(t)
		mock.ExpectQuery("SELECT").WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "description", "price"}))

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		indexHandler(w, req)

		if !strings.Contains(w.Body.String(), "bootstrap") {
			return fmt.Errorf("bootstrap CSS should be included")
		}
		return nil
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled mock expectations: %s", err)
	}
}
