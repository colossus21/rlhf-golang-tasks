//go:build v2
// +build v2

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("File Explorer")

	// Root directory to start browsing
	root, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	// Current directory being displayed
	currentDir := root

	// Display files in the list
	fileList := widget.NewList(
		func() int {
			files, _ := os.ReadDir(currentDir)
			return len(files)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			files, _ := os.ReadDir(currentDir)
			icon := widget.NewIcon(nil) // Placeholder for file/folder icon
			o.(*widget.Label).SetText(files[i].Name())
			o.(*widget.Label).Prepend(icon)
		})

	// Handle file/folder clicks
	fileList.OnSelected = func(id widget.ListItemID) {
		files, _ := os.ReadDir(currentDir)
		selected := files[id]
		newPath := filepath.Join(currentDir, selected.Name())
		if selected.IsDir() {
			currentDir = newPath
			fileList.Refresh()
		} else {
			// Handle file click (e.g., open in default app)
			fmt.Println("Open file:", newPath)
		}
	}

	// Navigate to parent directory
	backButton := widget.NewButton("Back", func() {
		newPath := filepath.Dir(currentDir)
		if newPath != currentDir {
			currentDir = newPath
			fileList.Refresh()
		}
	})

	// Display current directory path
	pathLabel := widget.NewLabel(currentDir)

	// Layout the widgets
	myWindow.SetContent(
		container.NewBorder(
			container.NewVBox(backButton, pathLabel),
			nil,
			nil,
			nil,
			fileList,
		))

	myWindow.Resize(fyne.NewSize(600, 400))
	myWindow.ShowAndRun()
}
