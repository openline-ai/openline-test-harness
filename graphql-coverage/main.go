package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func main() {
	repoURL := "https://api.github.com/repos/openline-ai/openline-customer-os/contents/packages/server/customer-os-api/graph/schemas"

	// Make a GET request to the GitHub API to get the repository contents.
	resp, err := http.Get(repoURL)
	if err != nil {
		fmt.Println("Error making the request:", err)
		return
	}
	defer resp.Body.Close()

	// Check the response status code.
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("GitHub API returned a non-OK status code: %d\n", resp.StatusCode)
		return
	}

	// Parse the JSON response.
	var contents []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		fmt.Println("Error decoding JSON response:", err)
		return
	}

	// Filter and display files with the .graphqls extension
	fmt.Println("Files with .graphqls extension:")
	for _, file := range contents {
		if file.Type == "file" && strings.HasSuffix(file.Name, ".graphqls") {
			fmt.Println(file.Name)
		}
	}
}
