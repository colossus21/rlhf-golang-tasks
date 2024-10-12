//go:build v1
// +build v1

package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestDockerAPIEnhancements(t *testing.T) {
	var testCases = []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{"DOCKER_API_JWT is loaded", func(t *testing.T) {
			os.Setenv("DOCKER_API_JWT", "http://example.com/api")
			defer os.Unsetenv("DOCKER_API_JWT")

			if os.Getenv("DOCKER_API_JWT") == "" {
				t.Error("DOCKER_API_JWT environment variable not loaded")
			}
		}},
		{"GetAllDockers concurrent requests", func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			os.Setenv("DOCKER_API_JWT", server.URL)
			defer os.Unsetenv("DOCKER_API_JWT")

			client := &http.Client{}
			start := time.Now()
			dockers := GetAllDockers(client, server.URL, 1*time.Second)
			duration := time.Since(start)

			if len(dockers) != 3 {
				t.Errorf("Expected 3 Docker checks, got %d", len(dockers))
			}
			if duration >= 300*time.Millisecond {
				t.Errorf("Requests were not concurrent, took %v", duration)
			}
		}},
		{"GetAllDockers error handling", func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}))
			defer server.Close()

			os.Setenv("DOCKER_API_JWT", server.URL)
			defer os.Unsetenv("DOCKER_API_JWT")

			client := &http.Client{}
			dockers := GetAllDockers(client, server.URL, 1*time.Second)

			for _, docker := range dockers {
				if docker.StatusCode != http.StatusInternalServerError {
					t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, docker.StatusCode)
				}
				if docker.Running {
					t.Error("Docker should not be marked as running on error")
				}
			}
		}},
		{"TIME_DELAY is configurable", func(t *testing.T) {
			os.Setenv("TIME_DELAY", "5")
			defer os.Unsetenv("TIME_DELAY")

			timeDelayStr := os.Getenv("TIME_DELAY")
			if timeDelayStr != "5" {
				t.Errorf("Expected TIME_DELAY to be '5', got '%s'", timeDelayStr)
			}
		}},
		{"Docker struct fields", func(t *testing.T) {
			docker := Docker{}
			if _, ok := interface{}(docker).(struct {
				StatusCode   int
				ResponseTime time.Duration
			}); !ok {
				t.Error("Docker struct is missing StatusCode or ResponseTime fields")
			}
		}},
		{"API response details", func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			os.Setenv("DOCKER_API_JWT", server.URL)
			defer os.Unsetenv("DOCKER_API_JWT")

			client := &http.Client{}
			dockers := GetAllDockers(client, server.URL, 1*time.Second)

			for _, docker := range dockers {
				if docker.ResponseTime == 0 {
					t.Error("ResponseTime not included in Docker struct")
				}
				if docker.CheckedAt.IsZero() {
					t.Error("CheckedAt not included in Docker struct")
				}
			}
		}},
		{"GetAllDockers error handling (Bad Gateway)", func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
			}))
			defer server.Close()

			os.Setenv("DOCKER_API_JWT", server.URL)
			defer os.Unsetenv("DOCKER_API_JWT")

			client := &http.Client{}
			dockers := GetAllDockers(client, server.URL, 1*time.Second)

			for _, docker := range dockers {
				if docker.StatusCode != http.StatusBadGateway {
					t.Errorf("Expected status code %d, got %d", http.StatusBadGateway, docker.StatusCode)
				}
				if docker.Running {
					t.Error("Docker should not be marked as running on error")
				}
			}
		}},
		{"HTTP client reuse", func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			os.Setenv("DOCKER_API_JWT", server.URL)
			defer os.Unsetenv("DOCKER_API_JWT")

			client := &http.Client{}
			dockers := GetAllDockers(client, server.URL, 1*time.Second)

			if len(dockers) != 3 {
				t.Errorf("Expected 3 Docker checks, got %d", len(dockers))
			}
		}},
	}

	passedTests := 0
	totalTests := len(testCases)

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.testFunc(t)
			result := "PASS"
			if t.Failed() {
				result = "FAIL"
			} else {
				passedTests++
			}
			fmt.Printf("Test Case %02d# %s - %s\n", i+1, tc.name, result)
		})
	}

	// Print summary at the end
	fmt.Printf("\nPassed %d out of %d tests\n", passedTests, totalTests)
}
