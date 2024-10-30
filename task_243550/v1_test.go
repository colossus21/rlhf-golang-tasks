//go:build v1
// +build v1

package main

import (
	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestFileExplorer1(t *testing.T) {
	// Test 1: Basic Directory Loading
	t.Run("Basic Directory Loading", func(t *testing.T) {
		app := test.NewApp()
		window := app.NewWindow("File Explorer")

		homeDir, _ := os.UserHomeDir()
		expectedFiles, _ := os.ReadDir(homeDir)

		list := window.Canvas().Objects()[0].(*widget.List)
		assert.Equal(t, len(expectedFiles), list.Length())
	})

	// Test 2: Navigation
	t.Run("Navigation", func(t *testing.T) {
		app := test.NewApp()
		window := app.NewWindow("File Explorer")

		list := window.Canvas().Objects()[0].(*widget.List)
		test.Tap(list)

		// Verify directory change
		homeDir, _ := os.UserHomeDir()
		newPath := filepath.Join(homeDir, "Documents")
		assert.Equal(t, newPath, currentDir)
	})

	// Test 3: Path Display
	t.Run("Path Display", func(t *testing.T) {
		app := test.NewApp()
		window := app.NewWindow("File Explorer")

		label := window.Canvas().Objects()[1].(*widget.Label)
		homeDir, _ := os.UserHomeDir()
		assert.Equal(t, homeDir, label.Text)
	})

	// Test 4: Refresh Functionality
	t.Run("Refresh", func(t *testing.T) {
		app := test.NewApp()
		window := app.NewWindow("File Explorer")

		list := window.Canvas().Objects()[0].(*widget.List)
		initialCount := list.Length()

		// Create temporary file
		homeDir, _ := os.UserHomeDir()
		tmpFile := filepath.Join(homeDir, "test.txt")
		os.Create(tmpFile)

		list.Refresh()
		assert.Equal(t, initialCount+1, list.Length())

		// Cleanup
		os.Remove(tmpFile)
	})
}
