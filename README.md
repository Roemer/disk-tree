# Introduction
Disk Tree is a cross plattform utility to visually show the disk space used by folders and subfolders.

# Usage
Just open the executable, browse for a folder and hit start. The data is shown during processing so you can already have a peek at the data while it still runs.

# Development
This application uses the [fyne](https://fyne.io/) framework and Go. So make sure you have the required dependencies.
Alternatively, you can develop in a container and build the application there and download it to test or if you have an X Window System forwarded to your host (eg. when using WSL2), you can directly open the application from the container.

As build system, [gotaskr](https://github.com/Roemer/gotaskr) is used. Best experienced in VSCode with the [gotaskr extension](https://marketplace.visualstudio.com/items?itemName=Roemer.gotaskr-vscode).
See `build/build.go` for the tasks that are available and support you with building.
