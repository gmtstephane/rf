package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
)

// findGitFolders recursively builds a list of paths, stopping at Git folders.
func findGitFolders(root string) ([]repo, error) {
	var gitFolders []repo

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if it's a directory
		if info.IsDir() {
			if strings.HasSuffix(path, ".terraform") {
				return filepath.SkipDir
			}
			if strings.HasSuffix(path, "/.git") {
				// Stop at the Git folder
				repoPath := strings.TrimSuffix(path, "/.git")
				if hasGoMod(repoPath) {

					last := getLastTwoElements(repoPath)
					gitFolders = append(gitFolders, repo{ShortName: last, FullPath: repoPath})
				} else {
					gitFolders = append(gitFolders, repo{ShortName: getLastTwoElements(repoPath), FullPath: repoPath})
				}
				return filepath.SkipDir
			}
		}

		return nil
	})

	return gitFolders, err
}

func hasGoMod(path string) bool {
	goModPath := filepath.Join(path, "go.mod")
	_, err := os.Stat(goModPath)
	return err == nil
}

type repo struct {
	ShortName string `json:"short_name,omitempty"`
	FullPath  string `json:"full_path,omitempty"`
}

func main() {
	rootPath := "/Users/stephane.guillemot/git"
	cacheFile := "/tmp/.repos"

	repos := readCache(rootPath, cacheFile)
	idx, err := fuzzyfinder.Find(repos, func(i int) string {
		return repos[i].ShortName
	}, fuzzyfinder.WithPromptString("select a repository:  "))
	if err != nil {
		fmt.Println("Error selecting repository:", err)
		return
	}
	fmt.Println(repos[idx].FullPath)

}
func getLastTwoElements(path string) string {
	elements := strings.Split(path, string(os.PathSeparator))
	if len(elements) < 2 {
		return path
	}
	return filepath.Join(elements[len(elements)-2], elements[len(elements)-1])
}

func readRepos(rootPath, cache string) []repo {
	gitFolders, err := findGitFolders(rootPath)
	if err != nil {
		fmt.Println("Error:", err)
		return []repo{}
	}

	writeCache(cache, gitFolders)

	return gitFolders
}

func writeCache(outputFile string, repos []repo) {
	// Marshal the array of repo to JSON
	reposJSON, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling repositories to JSON:", err)
		return
	}

	// Write the JSON to the outputFile
	err = os.WriteFile(outputFile, reposJSON, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

func readCache(rootPath, cache string) []repo {
	// Check if the cache file exists
	_, err := os.Stat(cache)
	if err == nil {
		// Cache file exists, read repositories from cache
		repos, err := readReposFromCache(cache)
		if err != nil {
			fmt.Println("Error reading cache:", err)
			// If there's an error reading from cache, fallback to fetching repositories
			return readRepos(rootPath, cache)
		}
		return repos
	}

	// Cache file does not exist, fetch repositories and write to cache
	return readRepos(rootPath, cache)
}

func readReposFromCache(cacheFile string) ([]repo, error) {
	// Read JSON from the cache file
	file, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON to an array of repo
	var repos []repo
	err = json.Unmarshal(file, &repos)
	if err != nil {
		return nil, err
	}

	return repos, nil
}
