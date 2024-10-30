//go:build v1
// +build v1

package main

import (
	"io/fs"
	"path/filepath"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("File Explorer")

	// Start in the user's home directory
	homeDir, _ := os.UserHomeDir()
	currentDir := homeDir

	// Display files
	fileList := widget.NewList(
		func() int {
			files, _ := fs.ReadDir(os.DirFS(currentDir), ".")
			return len(files)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			files, _ := fs.ReadDir(os.DirFS(currentDir), ".")
			o.(*widget.Label).SetText(files[i].Name())
		})

	fileList.OnSelected = func(id widget.ListItemID) {
		files, _ := fs.ReadDir(os.DirFS(currentDir), ".")
		selectedFile := files[id]
		if selectedFile.IsDir() {
			currentDir = filepath.Join(currentDir, selectedFile.Name())
			fileList.Refresh()
		}
	}

	fileList.OnUnselected = func(id widget.ListItemID) {
		// Optional: Handle unselecting items
	}

	// Navigation
	backButton := widget.NewButton("Back", func() {
		currentDir = filepath.Dir(currentDir)
		fileList.Refresh()
	})

	// Displaying current directory
	currentDirLabel := widget.NewLabel(currentDir)

	// Layout
	topBar := container.NewHBox(backButton, currentDirLabel)
	mainContent := container.NewBorder(topBar, nil, nil, nil, fileList)
	myWindow.SetContent(mainContent)

	myWindow.Resize(fyne.NewSize(600, 400))
	myWindow.ShowAndRun()
}
