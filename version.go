package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

type Tag struct {
	Name   string `json:"name"`
	Commit struct {
		SHA string `json:"sha"`
		URL string `json:"url"`
	} `json:"commit"`
}

type Commit struct {
	SHA string `json:"sha"`
}

var (
	// Populated by the Go linker during build
	version   = "unknown"
	gitCommit = "unknown"
	gitBranch = "unknown"
	buildDate = "unknown"
)

func versionHandler(w http.ResponseWriter, r *http.Request) {
	set_secure_headers(w, r)
	set_no_cache(w)

	latestCommit := "unknown"

	latestCommitObj := getLatestCommitFromBranch(gitBranch)
	if latestCommitObj != nil {
		latestCommit = latestCommitObj.SHA
	}

	err := json.NewEncoder(w).Encode(map[string]string{
		"current_git_commit": gitCommit,
		"git_branch":         gitBranch,
		"latest_git_commit":  latestCommit,
	})
	if err != nil {
		fmt.Println("JSON encode error: ", err)
	}
}

func PrintVersion() {
	fmt.Printf("Version: %s\nGit Commit: %s\nGit Branch: %s\nGo Version: %s\nGo OS/Arch: %s/%s\nBuild Date: %s\nTime: %s\n",
		version, gitCommit, gitBranch, runtime.Version(), runtime.GOOS, runtime.GOARCH, buildDate, time.Now().Format("2006-01-02 15:04:05 MST-07:00"))
}

// func getLatestGit() *Tag {
// 	url := "https://api.github.com/repos/adamjsturge/xsshunter-go/tags"
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		return nil
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		fmt.Printf("error fetching tags: %s", resp.Status)
// 		return nil
// 	}

// 	var tags []Tag
// 	err = json.NewDecoder(resp.Body).Decode(&tags)
// 	if err != nil {
// 		fmt.Printf("error decoding response: %v", err)
// 		return nil
// 	}

// 	// sort.Slice(tags, func(i, j int) bool {
// 	// 	return tags[i].Name > tags[j].Name
// 	// })

// 	if len(tags) > 0 {
// 		return &tags[0]
// 	}

// 	return nil
// }

func getLatestCommitFromBranch(branch string) *Commit {
	if branch == "unknown" {
		return nil
	}
	url := "https://api.github.com/repos/adamjsturge/xsshunter-go/commits/" + branch
	resp, err := http.Get(url) // #nosec G107
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("error fetching commit: %s", resp.Status)
		return nil
	}

	var commit Commit
	err = json.NewDecoder(resp.Body).Decode(&commit)
	if err != nil {
		fmt.Printf("error decoding response: %v", err)
		return nil
	}

	return &commit
}
