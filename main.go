package main

import (
	"disk-tree/core"
	"fmt"
	"sort"

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
	folderPathBinding := binding.BindPreferenceString("path", fyne.CurrentApp().Preferences())
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
				core.BuildTreeRecursive(rootEntry)
				fileTree.Refresh()
				progressIndicator.Stop()
				progressIndicator.Hide()
			}()
		}
	})

	// Sort Button
	sortButton := widget.NewButton("Sort", func() {
		dirStack := []*core.Entry{rootEntry}
		for len(dirStack) > 0 {
			// Get the next item and remove it from the stack
			dirIndex := len(dirStack) - 1
			dir := dirStack[dirIndex]
			dirStack = dirStack[:dirIndex]
			// Sort the folders from the dir
			sort.Slice(dir.Folders, func(i, j int) bool {
				return dir.Folders[i].Size > dir.Folders[j].Size
			})
			// Sort the files from the dir
			sort.Slice(dir.Files, func(i, j int) bool {
				return dir.Files[i].Size > dir.Files[j].Size
			})
			// Add all folders to the stack to process
			dirStack = append(dirStack, dir.Folders...)
		}

		fileTree.Refresh()
	})

	// The file tree
	fileTree = widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			if id == "" {
				return []widget.TreeNodeID{rootEntry.Path}
			}
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
			icon := widget.NewFileIcon(storage.NewFileURI("/"))
			progress := widget.NewProgressBar()
			progress.TextFormatter = func() string { return fmt.Sprintf("%06.2f%%", progress.Value) }
			progress.Min = 0
			progress.Max = 100
			return container.NewBorder(
				nil,
				nil,
				container.NewHBox(
					icon,
					widget.NewLabel("Left"),
				),
				container.NewHBox(
					widget.NewLabel("Right"), progress,
				),
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
			leftContainer := rootContainer.Objects[0].(*fyne.Container)
			rightContainer := rootContainer.Objects[1].(*fyne.Container)

			fileIcon := leftContainer.Objects[0].(*widget.FileIcon)
			nameLabel := leftContainer.Objects[1].(*widget.Label)
			sizeLabel := rightContainer.Objects[0].(*widget.Label)
			sizeBar := rightContainer.Objects[1].(*widget.ProgressBar)

			fileIcon.SetURI(storage.NewFileURI(currEntry.Path))
			switch currEntry.State {
			case core.UnprocessedState:
				nameLabel.Importance = widget.LowImportance
			case core.ProcessingState:
				nameLabel.Importance = widget.WarningImportance
			case core.ErrorState:
				nameLabel.Importance = widget.DangerImportance
			default:
				nameLabel.Importance = widget.MediumImportance
			}
			nameLabel.SetText(text)
			sizeLabel.SetText(core.BytesToIECString(currEntry.Size))
			sizeBar.SetValue(percentage)
		})

	// Window content
	w.SetContent(
		container.NewBorder(
			// Top
			container.NewVBox(
				container.NewBorder(
					nil, nil, widget.NewLabel("Path"), folderBrowseButton, folderEdit,
				),
				startButton,
				progressIndicator,
				sortButton,
			),
			nil, nil, nil,
			// Fill
			fileTree,
		),
	)

	// Center and start the application
	w.CenterOnScreen()
	w.ShowAndRun()
}

func getEntryFromTreeId(rootEntry *core.Entry, path string) *core.Entry {
	if rootEntry == nil {
		return &core.Entry{Path: "LOADING", Name: "LOADING"}
	}
	if path == rootEntry.Path {
		return rootEntry
	}
	return rootEntry.GetChildFromPath(path)
}
