// File: v1_test.go

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Config struct from config package
type Config struct {
	CertFile                string
	DatabasePath            string
	KeyFile                 string
	TagIdFieldName          string
	TrainingFieldName       string
	WildApricotAccountId    int
	ContactFilterQuery      string
	SSOClientID             string
	SSOClientSecret         string
	SSORedirectURI          string
	CookieStoreSecret       string
	WildApricotApiKey       string
	WildApricotWebhookToken string
	LogDir                  string
}

// CreateLogDir function from config package
func CreateLogDir(cfg *Config) error {
	if cfg.LogDir == "" {
		return nil // No custom log directory specified, use default
	}

	err := os.MkdirAll(cfg.LogDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}
	return nil
}

// Logger interface (simplified for testing)
type Logger interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}

// SimpleLogger implements Logger interface for testing
type SimpleLogger struct {
	file *os.File
}

func (l *SimpleLogger) log(level, message string) {
	timestamp := time.Now().Format(time.RFC3339)
	fmt.Fprintf(l.file, "%s [%s] %s\n", timestamp, level, message)
}

func (l *SimpleLogger) Info(args ...interface{}) {
	l.log("INFO", fmt.Sprint(args...))
}

func (l *SimpleLogger) Warn(args ...interface{}) {
	l.log("WARN", fmt.Sprint(args...))
}

func (l *SimpleLogger) Error(args ...interface{}) {
	l.log("ERROR", fmt.Sprint(args...))
}

func (l *SimpleLogger) Fatal(args ...interface{}) {
	l.log("FATAL", fmt.Sprint(args...))
	os.Exit(1)
}

// SetupLogger function from setup package
func SetupLogger(cfg *Config) (Logger, error) {
	if cfg.LogDir == "" {
		return nil, fmt.Errorf("log directory not specified")
	}

	logFilePath := filepath.Join(cfg.LogDir, "rfid-backend.log")
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	return &SimpleLogger{file: file}, nil
}

// Test cases
func TestCase01LogDirectoryCreation(t *testing.T) {
	t.Run("Test Case #01: Log Directory Creation", func(t *testing.T) {
		// Setup
		tempDir := filepath.Join(os.TempDir(), "rfid-backend-test-logs")
		cfg := &Config{
			LogDir: tempDir,
		}

		// Check: Implement a function to create this directory if it doesn't already exist.
		err := CreateLogDir(cfg)
		if err != nil {
			t.Errorf("CreateLogDir returned an error: %v", err)
		}
		if _, err := os.Stat(tempDir); os.IsNotExist(err) {
			t.Errorf("Log directory was not created")
		}

		// Cleanup
		os.RemoveAll(tempDir)
	})
}

func TestCase02LogFileCreationAndWriting(t *testing.T) {
	t.Run("Test Case #02: Log File Creation and Writing", func(t *testing.T) {
		// Setup
		tempDir := filepath.Join(os.TempDir(), "rfid-backend-test-logs")
		cfg := &Config{
			LogDir: tempDir,
		}
		err := CreateLogDir(cfg)
		if err != nil {
			t.Fatalf("CreateLogDir returned an error: %v", err)
		}

		// Check: Update the logging setup to write logs to a file in this directory
		logger, err := SetupLogger(cfg)
		if err != nil {
			t.Fatalf("SetupLogger returned an error: %v", err)
		}
		if logger == nil {
			t.Fatalf("SetupLogger returned nil logger")
		}

		logFile := filepath.Join(tempDir, "rfid-backend.log")
		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			t.Errorf("Log file was not created")
		}

		// Write a log entry
		logger.Info("Test log entry")

		// Read the log file
		content, err := os.ReadFile(logFile)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}
		if string(content) == "" {
			t.Errorf("Log file is empty")
		}

		// Cleanup
		os.RemoveAll(tempDir)
	})
}

func TestCase03LogTimestamps(t *testing.T) {
	t.Run("Test Case #03: Log Timestamps", func(t *testing.T) {
		// Setup
		tempDir := filepath.Join(os.TempDir(), "rfid-backend-test-logs")
		cfg := &Config{
			LogDir: tempDir,
		}
		err := CreateLogDir(cfg)
		if err != nil {
			t.Fatalf("CreateLogDir returned an error: %v", err)
		}

		logger, err := SetupLogger(cfg)
		if err != nil {
			t.Fatalf("SetupLogger returned an error: %v", err)
		}

		// Check: ensure that logs include timestamps for better traceability
		logger.Info("Test log entry for timestamp")

		logFile := filepath.Join(tempDir, "rfid-backend.log")
		content, err := os.ReadFile(logFile)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}

		if len(content) == 0 {
			t.Fatalf("Log file is empty")
		}

		// Check for timestamp in RFC3339 format
		timestamp := string(content[:20]) // First 20 characters should be the timestamp
		_, err = time.Parse(time.RFC3339, timestamp)
		if err != nil {
			t.Errorf("Log entry does not start with a valid RFC3339 timestamp: %v", err)
		}

		// Cleanup
		os.RemoveAll(tempDir)
	})
}

func TestCase04ErrorHandling(t *testing.T) {
	t.Run("Test Case #04: Error Handling", func(t *testing.T) {
		// Setup
		invalidDir := "/root/invalid-dir" // This should be an invalid directory for most systems
		cfg := &Config{
			LogDir: invalidDir,
		}

		// Check: implement error handling that logs any issues encountered during the creation of the log directory
		err := CreateLogDir(cfg)
		if err == nil {
			t.Errorf("CreateLogDir should return an error for invalid directory")
		}

		// Check: implement error handling when opening the log file
		logger, err := SetupLogger(cfg)
		if err == nil {
			t.Errorf("SetupLogger should return an error for invalid log directory")
		}
		if logger != nil {
			t.Errorf("SetupLogger should return nil logger for invalid log directory")
		}
	})
}
