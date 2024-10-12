//go:build v2
// +build v2

package main

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"testing"
)

func TestCreateProduct(t *testing.T) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		product        InventoryItem
		expectedStatus int
	}{
		{
			name: "Valid Product",
			product: InventoryItem{
				ProductName: "Test Product",
				Units:       10,
				Price:       9.99,
			},
			expectedStatus: 201,
		},
		{
			name: "Empty Product Name",
			product: InventoryItem{
				ProductName: "",
				Units:       10,
				Price:       9.99,
			},
			expectedStatus: 201,
		},
		{
			name: "Negative Units",
			product: InventoryItem{
				ProductName: "Test Product",
				Units:       -1,
				Price:       9.99,
			},
			expectedStatus: 201,
		},
		{
			name: "Zero Price",
			product: InventoryItem{
				ProductName: "Test Product",
				Units:       10,
				Price:       0,
			},
			expectedStatus: 201,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new Gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set the user in the context
			c.Set("user", "test-user-id")

			// Create request body
			jsonProduct, _ := json.Marshal(tt.product)
			c.Request = httptest.NewRequest("POST", "/createProduct", bytes.NewBuffer(jsonProduct))
			c.Request.Header.Set("Content-Type", "application/json")

			// Call the function
			createProduct(c)

			// Check the status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check the response body
			var response map[string]string
			json.Unmarshal(w.Body.Bytes(), &response)
			if response["message"] != "Product created successfully!" {
				t.Errorf("Expected message 'Product created successfully!', got '%s'", response["message"])
			}
		})
	}
}
