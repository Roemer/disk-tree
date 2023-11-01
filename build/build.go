package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/google/go-github/v54/github"
	"github.com/roemer/gotaskr"
	"github.com/roemer/gotaskr/execr"
	"github.com/roemer/gotaskr/log"
)

// Build variables
var version = "0.2.0"

// Internal variables
var outputDirectory = ".build-output"
var windowsBuildOutput = "disk-tree.exe"
var linuxBuildOutput = "disk-tree.tar.xz"

func main() {
	os.Exit(gotaskr.Execute())
}

func init() {
	gotaskr.Task("Setup:Fyne-Cmd", func() error {
		version := "latest" // latest or develop
		return execr.Run(true, "go", "install", fmt.Sprintf("fyne.io/fyne/v2/cmd/fyne@%s", version))
	})

	gotaskr.Task("Compile:Windows", func() error {
		os.Setenv("CGO_ENABLED", "1")
		os.Setenv("CC", "x86_64-w64-mingw32-gcc")

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		fynePath := path.Join(homeDir, "go/bin/fyne")
		if err := execr.Run(true, fynePath, "package", "-os", "windows", "--appVersion", version); err != nil {
			return nil
		}
		os.Mkdir(outputDirectory, os.ModePerm)
		return os.Rename(windowsBuildOutput, path.Join(outputDirectory, windowsBuildOutput))
	}).DependsOn("Setup:Fyne-Cmd")

	gotaskr.Task("Compile:Linux", func() error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		fynePath := path.Join(homeDir, "go/bin/fyne")
		if err := execr.Run(true, fynePath, "package", "-os", "linux", "--appVersion", version); err != nil {
			return nil
		}
		os.Mkdir(outputDirectory, os.ModePerm)
		return os.Rename(linuxBuildOutput, path.Join(outputDirectory, linuxBuildOutput))
	}).DependsOn("Setup:Fyne-Cmd")

	gotaskr.Task("Release:Create", func() error {
		log.Informationf("Creating new release for version %s", version)
		gitHubRepoParts := strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")
		gitHubOwner := gitHubRepoParts[0]
		gitHubRepo := gitHubRepoParts[1]
		gitHubToken := os.Getenv("GITHUB_TOKEN")

		// Create the client
		ctx := context.Background()
		client := github.NewTokenClient(ctx, gitHubToken)

		// Create the new release
		newRelease := &github.RepositoryRelease{
			Name:    github.String(version),
			Draft:   github.Bool(true),
			TagName: github.String(version),
		}
		release, _, err := client.Repositories.CreateRelease(ctx, gitHubOwner, gitHubRepo, newRelease)
		if err != nil {
			return err
		}
		log.Informationf("Created release: %s", *release.URL)

		// Upload the Windows artifact
		log.Information("Uploading Windows artifact")
		fileWin, err := os.Open(path.Join(outputDirectory, windowsBuildOutput))
		if err != nil {
			return err
		}
		defer fileWin.Close()
		_, _, err = client.Repositories.UploadReleaseAsset(ctx, gitHubOwner, gitHubRepo, *release.ID, &github.UploadOptions{
			Name: getFileNameForVersion(windowsBuildOutput, version),
		}, fileWin)
		if err != nil {
			return err
		}
		// Upload the Linux artifact
		log.Information("Uploading Linux artifact")
		fileLinux, err := os.Open(path.Join(outputDirectory, linuxBuildOutput))
		if err != nil {
			return err
		}
		defer fileLinux.Close()
		_, _, err = client.Repositories.UploadReleaseAsset(ctx, gitHubOwner, gitHubRepo, *release.ID, &github.UploadOptions{
			Name: getFileNameForVersion(linuxBuildOutput, version),
		}, fileLinux)
		if err != nil {
			return err
		}

		return nil
	})
}

////////////////////////////////////////////////////////////
// Internal Functions
////////////////////////////////////////////////////////////

func getFileNameForVersion(fileName string, version string) string {
	parts := strings.SplitN(fileName, ".", 2)
	return fmt.Sprintf("%s-%s.%s", parts[0], version, parts[1])
}
