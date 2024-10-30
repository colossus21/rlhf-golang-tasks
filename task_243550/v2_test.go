//go:build v2
// +build v2

package main

import (
	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestFileExplorer2(t *testing.T) {
	// Test 1: Basic Directory Loading
	t.Run("Basic Directory Loading", func(t *testing.T) {
		app := test.NewApp()
		window := app.NewWindow("File Explorer")

		homeDir, _ := os.UserHomeDir()
		expectedFiles, _ := os.ReadDir(homeDir)

		fileList := window.Canvas().Objects()[0].(*widget.List)
		assert.Equal(t, len(expectedFiles), fileList.Length())
	})

	// Test 2: Navigation with Icons
	t.Run("Navigation", func(t *testing.T) {
		app := test.NewApp()
		window := app.NewWindow("File Explorer")

		list := window.Canvas().Objects()[0].(*widget.List)
		item := list.CreateItem()
		assert.NotNil(t, item.(*widget.Label).Leading) // Verify icon exists

		test.Tap(list)
		homeDir, _ := os.UserHomeDir()
		newPath := filepath.Join(homeDir, "Documents")
		assert.Equal(t, newPath, currentDir)
	})

	// Test 3: Path Display and Layout
	t.Run("Path Display", func(t *testing.T) {
		app := test.NewApp()
		window := app.NewWindow("File Explorer")

		vBox := window.Canvas().Objects()[1].(*container.VBox)
		pathLabel := vBox.Objects[1].(*widget.Label)
		homeDir, _ := os.UserHomeDir()
		assert.Equal(t, homeDir, pathLabel.Text)
	})

	// Test 4: Refresh and Selection
	t.Run("Refresh", func(t *testing.T) {
		app := test.NewApp()
		window := app.NewWindow("File Explorer")

		list := window.Canvas().Objects()[0].(*widget.List)
		homeDir, _ := os.UserHomeDir()

		// Test file selection
		test.Tap(list)
		files, _ := os.ReadDir(homeDir)
		assert.NotNil(t, list.Selected())

		// Test refresh after navigation
		backBtn := window.Canvas().Objects()[1].(*container.VBox).Objects[0].(*widget.Button)
		test.Tap(backBtn)
		assert.Nil(t, list.Selected()) // Selection should clear
	})
}
