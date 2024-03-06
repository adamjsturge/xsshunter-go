package main

import (
	"fmt"
	"runtime"
)

var (
	// Populated by the Go linker during build
	version   = "dev"
	gitCommit = "unknown"
	gitBranch = "unknown"
	buildDate = "unknown"
)

// PrintVersion prints the version information
func PrintVersion() {
	fmt.Printf("Version: %s\nGit Commit: %s\nGit Branch: %s\nGo Version: %s\nGo OS/Arch: %s/%s\nBuild Date: %s\n",
		version, gitCommit, gitBranch, runtime.Version(), runtime.GOOS, runtime.GOARCH, buildDate)
}
