package main

import (
	"disk-tree/core"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

func main() {
	// Create the app and set the theme
	a := app.NewWithID("roemer.go-disk-tree")

	// Create the window
	w := a.NewWindow("Disk Tree")
	w.Resize(fyne.NewSize(800, 600))

	// Menu
	menuFile := fyne.NewMenu("File")
	menuSettings := fyne.NewMenu("Settings")

	var item *fyne.MenuItem
	item = fyne.NewMenuItem("My Feature", func() {
		item.Checked = !item.Checked
		menuSettings.Refresh()
	})

	menuSettings.Items = append(menuSettings.Items, item)
	mainMenu := fyne.NewMainMenu(menuFile, menuSettings)
	w.SetMainMenu(mainMenu)

	// Folder Selection
	folderPathBinding := binding.NewString()
	folderEdit := widget.NewEntryWithData(folderPathBinding)
	folderBrowseButton := widget.NewButton("Browse", func() {
		folderDialog := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
			if lu != nil {
				folderPathBinding.Set(lu.Path())
			}
		}, w)
		currentPath, err := folderPathBinding.Get()
		if err != nil {
			uri := storage.NewFileURI(currentPath)
			l, err := storage.ListerForURI(uri)
			if err != nil {
				folderDialog.SetLocation(l)
			}
		}
		folderDialog.Show()
	})

	// Running indicator
	progressIndicator := core.NewProgressBarInfiniteSmall()
	progressIndicator.Stop()
	progressIndicator.Hide()

	// Start button
	var fileTree *widget.Tree
	var rootEntry *core.Entry
	startButton := widget.NewButton("Start", func() {
		progressIndicator.Show()
		progressIndicator.Start()
		currentPath, err := folderPathBinding.Get()
		if err == nil {
			rootEntry = core.Prepare(currentPath)
			go func() {
				err := core.BuildTree(rootEntry)
				if err != nil {
					fmt.Println(err)
				}
				fileTree.Refresh()
				progressIndicator.Stop()
				progressIndicator.Hide()
			}()
		}
	})

	// The file tree
	fileTree = widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			ids := []widget.TreeNodeID{}
			currEntry := getEntryFromTreeId(rootEntry, id)
			for _, entry := range currEntry.Folders {
				ids = append(ids, entry.Path)
			}
			for _, entry := range currEntry.Files {
				ids = append(ids, entry.Path)
			}
			return ids
		},
		func(id widget.TreeNodeID) bool {
			currEntry := getEntryFromTreeId(rootEntry, id)
			return currEntry.IsFolder
		},
		func(branch bool) fyne.CanvasObject {
			progress := widget.NewProgressBar()
			progress.TextFormatter = func() string { return fmt.Sprintf("%.2f%%", progress.Value) }
			progress.Min = 0
			progress.Max = 100
			return container.NewBorder(
				nil, nil, widget.NewLabel("Left"), container.NewHBox(widget.NewLabel("Right"), progress),
			)
		},
		func(id widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
			currEntry := getEntryFromTreeId(rootEntry, id)
			text := currEntry.Name
			if branch {
				text += fmt.Sprintf(" : %d / %d", len(currEntry.Folders), len(currEntry.Files))
			}
			percentage := float64(currEntry.Size) / float64(rootEntry.Size) * 100
			rootContainer := o.(*fyne.Container)
			rightContainer := rootContainer.Objects[1].(*fyne.Container)
			rootContainer.Objects[0].(*widget.Label).TextStyle.Bold = currEntry.Processing
			rootContainer.Objects[0].(*widget.Label).SetText(text)
			rightContainer.Objects[0].(*widget.Label).SetText(ByteCountIEC(currEntry.Size))
			rightContainer.Objects[1].(*widget.ProgressBar).SetValue(percentage)
		})

	// Window content
	w.SetContent(
		container.NewBorder(
			container.NewVBox(
				container.NewBorder(
					nil, nil, widget.NewLabel("Path"), folderBrowseButton, folderEdit,
				),
				startButton,
				progressIndicator,
			), nil, nil, nil, fileTree,
		),
	)

	// Center and start the application
	w.CenterOnScreen()
	w.ShowAndRun()
}

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func getEntryFromTreeId(rootEntry *core.Entry, path string) *core.Entry {
	if rootEntry == nil {
		return &core.Entry{Path: "STILL LOADING", Name: "STILL LOADING"}
	}
	if path == "" {
		return rootEntry
	}
	parts := strings.Split(path, string('/'))
	currentEntry := rootEntry
	for _, part := range parts {
		for _, file := range currentEntry.Files {
			if file.Name == part {
				return file
			}
		}
		for _, dir := range currentEntry.Folders {
			if dir.Name == part {
				currentEntry = dir
				break
			}
		}
	}
	return currentEntry
}
