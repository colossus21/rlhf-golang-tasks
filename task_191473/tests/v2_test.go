// File: v2_test.go

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Config struct from v2
type Config struct {
	LogDir string
	// Other fields omitted for brevity
}

// createLogDir function from v2
func createLogDir(dir string) error {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	return nil
}

// Simple logger interface for testing
type Logger interface {
	Info(args ...interface{})
}

// Simple logger implementation for testing
type testLogger struct {
	writer io.Writer
}

func (l *testLogger) Info(args ...interface{}) {
	fmt.Fprintf(l.writer, "%s INFO %s\n", time.Now().Format(time.RFC3339), fmt.Sprint(args...))
}

// SetupLogger function adapted for testing
func SetupLogger(cfg *Config) (Logger, error) {
	if cfg.LogDir == "" {
		return &testLogger{writer: os.Stdout}, nil
	}

	logFilePath := filepath.Join(cfg.LogDir, "app.log")
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("error opening log file: %s", err)
	}

	multiWriter := io.MultiWriter(file, os.Stdout)
	return &testLogger{writer: multiWriter}, nil
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
		err := createLogDir(cfg.LogDir)
		if err != nil {
			t.Errorf("createLogDir returned an error: %v", err)
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
		err := createLogDir(cfg.LogDir)
		if err != nil {
			t.Fatalf("createLogDir returned an error: %v", err)
		}

		// Check: Update the logging setup to write logs to a file in this directory
		logger, err := SetupLogger(cfg)
		if err != nil {
			t.Fatalf("SetupLogger returned an error: %v", err)
		}
		if logger == nil {
			t.Fatalf("SetupLogger returned nil logger")
		}

		logFile := filepath.Join(tempDir, "app.log")
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
		err := createLogDir(cfg.LogDir)
		if err != nil {
			t.Fatalf("createLogDir returned an error: %v", err)
		}

		logger, err := SetupLogger(cfg)
		if err != nil {
			t.Fatalf("SetupLogger returned an error: %v", err)
		}

		// Check: ensure that logs include timestamps for better traceability
		logger.Info("Test log entry for timestamp")

		logFile := filepath.Join(tempDir, "app.log")
		content, err := os.ReadFile(logFile)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}

		if len(content) == 0 {
			t.Fatalf("Log file is empty")
		}

		// Simple check for timestamp format (YYYY-MM-DD)
		if len(content) < 10 || string(content[:4]) != time.Now().Format("2006") {
			t.Errorf("Log entry does not start with a timestamp")
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
		err := createLogDir(cfg.LogDir)
		if err == nil {
			t.Errorf("createLogDir should return an error for invalid directory")
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
