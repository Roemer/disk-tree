package core

import (
	"fmt"
	"io/fs"
	"os"
	"path"
)

type Entry struct {
	Path       string
	Name       string
	IsFolder   bool
	Size       int64
	Processing bool
	Folders    []*Entry
	Files      []*Entry
	parent     *Entry
}

func (entry *Entry) addFile(path string, info fs.FileInfo) {
	childEntry := &Entry{
		Path:     path,
		Name:     info.Name(),
		IsFolder: false,
		Size:     info.Size(),
	}
	entry.Size += childEntry.Size
	entry.Files = append(entry.Files, childEntry)
}

func Prepare(currentPath string) *Entry {
	return &Entry{Path: currentPath, Name: "", IsFolder: true}
}

func BuildTree(rootEntry *Entry) error {
	dirStack := []*Entry{rootEntry}
	for len(dirStack) > 0 {
		// Get the next directory to process
		dirIndex := len(dirStack) - 1
		dir := dirStack[dirIndex]
		setProcessing(dir, true)
		// Pop it from the stack
		dirStack = dirStack[:dirIndex]
		// Read the directory
		childEntries, err := os.ReadDir(dir.Path)
		if err != nil {
			fmt.Println(err)
			setProcessing(dir, false)
			continue
			//return nil, err
		}
		// Process the child entries
		for _, entry := range childEntries {
			entryPath := path.Join(dir.Path, entry.Name())
			// Process the folder
			if entry.IsDir() {
				childDir := &Entry{Path: entryPath, Name: entry.Name(), IsFolder: true, parent: dir}
				dirStack = append(dirStack, childDir)
				dir.Folders = append(dir.Folders, childDir)
				continue
			}
			// Process the file
			info, err := entry.Info()
			if err != nil {
				fmt.Println(err)
				continue
			}
			dir.addFile(entryPath, info)
		}
		// Increase the space of the parent if needed
		parent := dir.parent
		for parent != nil {
			parent.Size += dir.Size
			parent = parent.parent
		}
		setProcessing(dir, false)
	}
	return nil
}

func setProcessing(entry *Entry, processing bool) {
	entry.Processing = processing
	parent := entry.parent
	for parent != nil {
		parent.Processing = processing
		parent = parent.parent
	}
}
