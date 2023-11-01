package core

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
)

type EntryState int

const (
	UnprocessedState EntryState = iota
	ProcessingState
	ProcessedState
	ErrorState
)

type Entry struct {
	Path     string
	Name     string
	IsFolder bool
	Size     int64
	State    EntryState
	Error    error
	Folders  []*Entry
	Files    []*Entry
	parent   *Entry
}

func (e *Entry) GetChildFromPath(path string) *Entry {
	// Split the path into parts
	pathParts := strings.Split(path, string('/'))
	// Walk thru the parts until the item is found
	currentEntry := e
	for _, part := range pathParts {
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

func (entry *Entry) addFile(path string, info fs.FileInfo) {
	childEntry := &Entry{
		Path:     path,
		Name:     info.Name(),
		IsFolder: false,
		Size:     info.Size(),
		State:    ProcessedState,
	}
	entry.Size += childEntry.Size
	entry.Files = append(entry.Files, childEntry)
}

func Prepare(currentPath string) *Entry {
	return &Entry{Path: currentPath, Name: currentPath, IsFolder: true}
}

// This is a recursive approach to build the tree from top to down
func BuildTreeRecursive(rootEntry *Entry, ctx context.Context) {
	processEntry(rootEntry, ctx)
}

func processEntry(currentEntry *Entry, ctx context.Context) {
	// Handle cancellation
	select {
	case <-ctx.Done():
		return
	default:
	}

	// Set processing
	currentEntry.State = ProcessingState
	// Read the folder
	childEntries, err := os.ReadDir(currentEntry.Path)
	if err != nil {
		fmt.Println(err)
		currentEntry.Error = err
		currentEntry.State = ErrorState
		return
	}
	foldersToProcess := []*Entry{}
	// Process the content of the folder
	for _, entry := range childEntries {
		// Compute the path of the entry
		entryPath := path.Join(currentEntry.Path, entry.Name())
		// Process a folder
		if entry.IsDir() {
			childDir := &Entry{Path: entryPath, Name: entry.Name(), IsFolder: true, parent: currentEntry}
			currentEntry.Folders = append(currentEntry.Folders, childDir)
			// Store the folder to process later
			foldersToProcess = append(foldersToProcess, childDir)
			continue
		}
		// Process a file
		info, err := entry.Info()
		if err != nil {
			fmt.Println(err)
			continue
		}
		currentEntry.addFile(entryPath, info)
	}
	// Increase the space of the folder and all parents
	parent := currentEntry.parent
	for parent != nil {
		parent.Size += currentEntry.Size
		parent = parent.parent
	}
	// Now process the folders
	for _, folder := range foldersToProcess {
		processEntry(folder, ctx)
	}
	// Set processed
	currentEntry.State = ProcessedState
}

// This is an iterative approach to build the tree which processes layer by layer
func BuildTreeIterative(rootEntry *Entry) {
	dirStack := []*Entry{rootEntry}
	for len(dirStack) > 0 {
		// Get the next folder to process
		currentEntry := dirStack[0]
		// Pop it from the stack
		dirStack = dirStack[1:]
		// Mark it as processing
		setProcessing(currentEntry, true)
		// Read the folder
		childEntries, err := os.ReadDir(currentEntry.Path)
		if err != nil {
			fmt.Println(err)
			currentEntry.Error = err
			setProcessing(currentEntry, false)
			currentEntry.State = ErrorState
			continue
		}
		// Process the child entries
		for _, entry := range childEntries {
			entryPath := path.Join(currentEntry.Path, entry.Name())
			// Process a folder
			if entry.IsDir() {
				childDir := &Entry{Path: entryPath, Name: entry.Name(), IsFolder: true, parent: currentEntry}
				dirStack = append(dirStack, childDir)
				currentEntry.Folders = append(currentEntry.Folders, childDir)
				continue
			}
			// Process a file
			info, err := entry.Info()
			if err != nil {
				fmt.Println(err)
				continue
			}
			currentEntry.addFile(entryPath, info)
		}
		// Increase the space of all parents
		parent := currentEntry.parent
		for parent != nil {
			parent.Size += currentEntry.Size
			parent = parent.parent
		}
		setProcessing(currentEntry, false)
	}
}

func setProcessing(entry *Entry, processing bool) {
	var newState EntryState
	if processing {
		newState = ProcessingState
	} else {
		newState = ProcessedState
	}
	entry.State = newState
	parent := entry.parent
	for parent != nil {
		parent.State = newState
		parent = parent.parent
	}
}
