package main

import (
	"os"
	"path"

	"github.com/roemer/gotaskr"
	"github.com/roemer/gotaskr/execr"
)

func main() {
	os.Exit(gotaskr.Execute())
}

func init() {
	gotaskr.Task("Setup:Fyne-Cmd", func() error {
		// fyne.io/fyne/v2/cmd/fyne@latest
		return execr.Run(true, "go", "install", "fyne.io/fyne/v2/cmd/fyne@develop")
	})

	gotaskr.Task("Compile:Windows", func() error {
		os.Setenv("CGO_ENABLED", "1")
		os.Setenv("CC", "x86_64-w64-mingw32-gcc")

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		fynePath := path.Join(homeDir, "go/bin/fyne")

		return execr.Run(true, fynePath, "package", "-os", "windows")
	}).DependsOn("Setup:Fyne-Cmd")

	gotaskr.Task("Compile:Linux", func() error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		fynePath := path.Join(homeDir, "go/bin/fyne")

		return execr.Run(true, fynePath, "package", "-os", "linux")
	}).DependsOn("Setup:Fyne-Cmd")
}
