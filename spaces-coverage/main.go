package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	repoToClone := "openline-customer-os"
	repoPath := cloneRepo(repoToClone)
	fmt.Println("Cloned repository:", repoPath)
	clonedRepo := "openline-customer-os"

	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	clonedRepoPath := filepath.Join(currentDir, clonedRepo)
	graphQlFiles, graphQlTestFiles := scanForFiles(clonedRepoPath, []string{".git", "/.git", "/.git/", ".gitignore", ".DS_Store", ".idea", "/.idea/", "/.idea"})

	fmt.Printf("Spaces test coverage: %.2f%%", (float64(len(graphQlTestFiles)*100))/float64(len(graphQlFiles)))
}

func scanForFiles(dirPath string, ignore []string) ([]string, []string) {

	var graphQlFiles []string
	var graphQlTestFiles []string

	// Scan
	err := filepath.Walk(dirPath, func(path string, f os.FileInfo, err error) error {
		_continue := false
		// Loop : Ignore Files & Folders
		for _, i := range ignore {
			// If ignored path
			if strings.Index(path, i) != -1 {
				// Continue
				_continue = true
			}
		}

		if _continue == false {
			f, err = os.Stat(path)
			// If no error
			if err != nil {
				log.Fatal(err)
			}
			// File & Folder Mode
			fMode := f.Mode()
			// Is folder
			if fMode.IsRegular() {
				// Append to Files Array
				filename := filepath.Base(path)
				filePath := filepath.FromSlash(path)
				if filepath.Ext(filePath) == ".graphql" {
					graphQlFiles = append(graphQlFiles, strings.TrimSuffix(filename, filepath.Ext(filename)))
				}
				if strings.HasSuffix(filename, ".test.ts") {
					testFileWithNoExtension := strings.TrimSuffix(filename, filepath.Ext(filename))
					testFileWithNoTest := strings.TrimSuffix(testFileWithNoExtension, ".test")
					graphQlTestFiles = append(graphQlTestFiles, testFileWithNoTest)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil
	}

	return graphQlFiles, graphQlTestFiles
}

func cloneRepo(repoToClone string) string {
	repo, err := git.PlainClone(repoToClone, false, &git.CloneOptions{
		URL:      "https://github.com/openline-ai/openline-customer-os.git",
		Progress: os.Stdout,
	})
	if err.Error() == "repository already exists" {
		return repoToClone
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return worktree.Filesystem.Root()
}
