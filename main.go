package main

import (
	"context"
	"disk-tree/core"
	"fmt"
	"time"

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
	window := a.NewWindow("Disk Tree")
	window.Resize(fyne.NewSize(800, 600))

	// UI elements used in various places
	var fileTree *widget.Tree

	// Settings
	separateFoldersAndFilesSetting := true
	sortBySetting := core.SortByName

	// Main Menu
	menuFile := fyne.NewMenu("File")
	menuSettings := fyne.NewMenu("Settings")
	// Menu Sorting
	var sortByNameMenu, sortBySizeMenu *fyne.MenuItem
	// By Name
	sortByNameMenu = fyne.NewMenuItem("By Name", func() {
		if sortBySetting != core.SortByName {
			sortBySetting = core.SortByName
			sortByNameMenu.Checked = true
			sortBySizeMenu.Checked = false
			menuSettings.Refresh()
			fileTree.Refresh()
		}
	})
	sortByNameMenu.Checked = sortBySetting == 0
	// By Size
	sortBySizeMenu = fyne.NewMenuItem("By Size", func() {
		if sortBySetting != core.SortBySize {
			sortBySetting = core.SortBySize
			sortBySizeMenu.Checked = true
			sortByNameMenu.Checked = false
			menuSettings.Refresh()
			fileTree.Refresh()
		}
	})
	sortBySizeMenu.Checked = sortBySetting == 1
	// Sort Menu
	sortMenuItem := fyne.NewMenuItem("Sort Order", nil)
	sortMenuItem.ChildMenu = fyne.NewMenu("Sort Options", sortByNameMenu, sortBySizeMenu)
	// Separate Folders / Files
	var separateFoldersAndFilesMenu *fyne.MenuItem
	separateFoldersAndFilesMenu = fyne.NewMenuItem("Separate Folders/Files", func() {
		separateFoldersAndFilesSetting = !separateFoldersAndFilesSetting
		separateFoldersAndFilesMenu.Checked = separateFoldersAndFilesSetting
		menuSettings.Refresh()
		fileTree.Refresh()
	})
	separateFoldersAndFilesMenu.Checked = separateFoldersAndFilesSetting
	// Exclusions
	exclusionsMenu := fyne.NewMenuItem("Exclusions", func() {
		dialog.ShowCustomConfirm("Exclusion List", "Ok", "Cancel", container.NewStack(widget.NewMultiLineEntry()), nil, window)
	})

	// Build the menu
	menuSettings.Items = append(menuSettings.Items, sortMenuItem, separateFoldersAndFilesMenu, exclusionsMenu)
	mainMenu := fyne.NewMainMenu(menuFile, menuSettings)
	window.SetMainMenu(mainMenu)

	// Folder selection
	folderPathBinding := binding.BindPreferenceString("path", fyne.CurrentApp().Preferences())
	folderEdit := widget.NewEntryWithData(folderPathBinding)
	folderBrowseButton := widget.NewButton("Browse", func() {
		folderDialog := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
			if lu != nil {
				folderPathBinding.Set(lu.Path())
			}
		}, window)
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

	// Start/Stop buttons
	var rootEntry *core.Entry
	var ctx context.Context
	var cancel context.CancelFunc
	var startButton *widget.Button
	var stopButton *widget.Button

	startButton = widget.NewButton("Start", func() {
		startButton.Hide()
		stopButton.Show()
		progressIndicator.Show()
		progressIndicator.Start()
		currentPath, err := folderPathBinding.Get()
		if err == nil {
			// Prepare the context
			ctx, cancel = context.WithCancel(context.Background())

			// Prepare the root entry
			rootEntry = core.Prepare(currentPath)

			// Start a goroutine which builds the tree in another subroutine and regularly refreshes the tree
			go func() {
				go func() {
					core.BuildTreeRecursive(rootEntry, ctx)
					defer cancel()
				}()
				for {
					select {
					case <-ctx.Done():
						fileTree.Refresh()
						progressIndicator.Stop()
						progressIndicator.Hide()
						stopButton.Hide()
						startButton.Show()
						return
					case <-time.After(5 * time.Second):
						fileTree.Refresh()
					}
				}
			}()
		}
	})
	stopButton = widget.NewButton("Stop", func() {
		if cancel != nil {
			cancel()
		}
	})
	stopButton.Hide()

	// The file tree
	fileTree = widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			if id == "" {
				return []widget.TreeNodeID{rootEntry.Path}
			}
			ids := []widget.TreeNodeID{}
			currEntry := getEntryFromTreeId(rootEntry, id)

			// Sorting
			entries := []*core.Entry{}
			if separateFoldersAndFilesSetting {
				// Separately sort folders, then files, then merge them
				dirEntries := []*core.Entry{}
				dirEntries = append(dirEntries, currEntry.Folders...)
				core.SortEntries(sortBySetting, dirEntries)

				fileEntries := []*core.Entry{}
				fileEntries = append(fileEntries, currEntry.Files...)
				core.SortEntries(sortBySetting, fileEntries)

				entries = append(entries, dirEntries...)
				entries = append(entries, fileEntries...)
			} else {
				// Sort folders and files together
				entries = append(entries, currEntry.Folders...)
				entries = append(entries, currEntry.Files...)
				core.SortEntries(sortBySetting, entries)
			}
			// Add all the ids
			for _, entry := range entries {
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
	window.SetContent(
		container.NewBorder(
			// Top
			container.NewVBox(
				container.NewBorder(
					nil, nil, widget.NewLabel("Path"), folderBrowseButton, folderEdit,
				),
				startButton,
				stopButton,
				progressIndicator,
			),
			nil, nil, nil,
			// Fill
			fileTree,
		),
	)

	// Center and start the application
	window.CenterOnScreen()
	window.ShowAndRun()
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
