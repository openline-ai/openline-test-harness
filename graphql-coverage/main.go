package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

func main() {
	repoURL := "https://api.github.com/repos/openline-ai/openline-customer-os/contents/packages/server/customer-os-api/graph/schemas"

	resp, err := http.Get(repoURL)
	if err != nil {
		fmt.Println("Error making the request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("GitHub API returned a non-OK status code: %d\n", resp.StatusCode)
		return
	}

	var contents []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		fmt.Println("Error decoding JSON response:", err)
		return
	}

	fmt.Println("Files with .graphqls extension:")
	for _, file := range contents {
		if file.Type == "file" && strings.HasSuffix(file.Name, ".graphqls") {
			fmt.Println("\nFile:", file.Name)
			schemaURL := "https://raw.githubusercontent.com/openline-ai/openline-customer-os/main/packages/server/customer-os-api/graph/schemas/" + file.Name
			schemaResp, err := http.Get(schemaURL)
			if err != nil {
				fmt.Println("Error fetching schema content:", err)
				continue
			}
			defer schemaResp.Body.Close()

			schemaContent, err := io.ReadAll(schemaResp.Body)
			if err != nil {
				fmt.Println("Error reading schema content:", err)
				continue
			}

			queriesSnippet := regexp.MustCompile(`extend type Query {([\s\S]*?)}`)
			queries := funcName(queriesSnippet, schemaContent)
			if len(queries) > 0 {
				fmt.Println("Query Names:", strings.Join(queries, ", "))
			}

			mutationsSnippet := regexp.MustCompile(`extend type Mutation {([\s\S]*?)}`)
			mutations := funcName(mutationsSnippet, schemaContent)
			if len(mutations) > 0 {
				fmt.Println("Mutation Names:", strings.Join(mutations, ", "))
			}
		}
	}
}

func funcName(queriesMutations *regexp.Regexp, schemaContent []byte) []string {
	queryMutationPattern := `\b(\w+)\(`
	annotationPattern := `@[^@\s]*`
	re := regexp.MustCompile(annotationPattern)
	matchQueries := queriesMutations.FindStringSubmatch(string(schemaContent))
	var queriesMutationsNames []string
	if len(matchQueries) >= 2 {
		queryBlock := matchQueries[1]
		sanitizedQueryBlock := re.ReplaceAllString(queryBlock, "")

		queriesMutations = regexp.MustCompile(queryMutationPattern)
		matches := queriesMutations.FindAllStringSubmatch(sanitizedQueryBlock, -1)
		if matches != nil {
			for _, match := range matches {
				queriesMutationName := match[1]
				queriesMutationsNames = append(queriesMutationsNames, queriesMutationName)
			}

		}
	}
	return queriesMutationsNames
}
