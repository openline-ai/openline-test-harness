package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

func main() {
	repoURL := "https://api.github.com/repos/openline-ai/openline-customer-os/contents/packages/server/customer-os-api/graph/schemas"

	// Make a GET request to the GitHub API to get the repository contents
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
		if file.Type == "file" && strings.HasSuffix(file.Name, ".graphqls") && file.Name == "user.graphqls" {
			fmt.Println("\nFile:", file.Name)
			// Fetch the content of the GraphQL schema file
			schemaURL := "https://raw.githubusercontent.com/openline-ai/openline-customer-os/main/packages/server/customer-os-api/graph/schemas/" + file.Name
			schemaResp, err := http.Get(schemaURL)
			if err != nil {
				fmt.Println("Error fetching schema content:", err)
				continue
			}
			defer schemaResp.Body.Close()

			// Read and parse the content of the schema file
			schemaContent, err := ioutil.ReadAll(schemaResp.Body)
			if err != nil {
				fmt.Println("Error reading schema content:", err)
				continue
			}

			// Use regular expressions to find and display mutation names
			queries := regexp.MustCompile(`extend type Query {([\s\S]*?)}`)
			matchQueries := queries.FindStringSubmatch(string(schemaContent))
			if len(matchQueries) >= 2 {
				queryBlock := matchQueries[1]
				queries = regexp.MustCompile(`\b(\w+)\(`) // Updated regex pattern
				matches := queries.FindAllStringSubmatch(queryBlock, -1)
				if matches != nil {
					var queryNames []string
					for _, match := range matches {
						mutationName := match[1]
						queryNames = append(queryNames, mutationName)
					}
					fmt.Println("Query Names:", strings.Join(queryNames, ", "))
				}
			}

			mutations := regexp.MustCompile(`extend type Mutation {([\s\S]*?)}`)
			matchMutations := mutations.FindStringSubmatch(string(schemaContent))
			if len(matchMutations) >= 2 {
				mutationBlock := matchMutations[1]
				mutations = regexp.MustCompile(`\b(\w+)\(`) // Updated regex pattern
				matches := mutations.FindAllStringSubmatch(mutationBlock, -1)
				if matches != nil {
					var mutationNames []string
					for _, match := range matches {
						mutationName := match[1]
						mutationNames = append(mutationNames, mutationName)
					}
					fmt.Println("Mutation Names:", strings.Join(mutationNames, ", "))
				}
			}
		}
	}
}
