package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
)

type Tag struct {
	Name string `json:"name"`
}

var (
	// Populated by the Go linker during build
	version   = "dev"
	gitCommit = "unknown"
	gitBranch = "unknown"
	buildDate = "unknown"
)

func versionHandler(w http.ResponseWriter, r *http.Request) {
	set_secure_headers(w, r)
	set_no_cache(w)

	json.NewEncoder(w).Encode(map[string]string{
		"current_version": version,
		"git_branch":      gitBranch,
		"latest_version":  getLatestTag(),
	})
}

func PrintVersion() {
	fmt.Printf("Version: %s\nGit Commit: %s\nGit Branch: %s\nGo Version: %s\nGo OS/Arch: %s/%s\nBuild Date: %s\n",
		version, gitCommit, gitBranch, runtime.Version(), runtime.GOOS, runtime.GOARCH, buildDate)
}

func getLatestTag() string {
	url := "https://api.github.com/repos/adamjsturge/xsshunter-go/tags"
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("error fetching tags: %s", resp.Status)
		return ""
	}

	var tags []Tag
	err = json.NewDecoder(resp.Body).Decode(&tags)
	if err != nil {
		fmt.Printf("error decoding response: %v", err)
		return ""
	}

	// sort.Slice(tags, func(i, j int) bool {
	// 	return tags[i].Name > tags[j].Name
	// })

	if len(tags) > 0 {
		return tags[0].Name
	}

	return ""
}
