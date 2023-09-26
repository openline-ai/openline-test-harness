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
	baseUrl := "https://api.github.com/repos/openline-ai/openline-customer-os/contents/packages/server/customer-os-api/graph/"
	//queriesMutations := getQueriesMutations(baseUrl)
	getTestsForQueriesMutations(baseUrl)
	//computeCoverage()
	//fmt.Println(queriesMutations)
}

func getTestsForQueriesMutations(baseUrl string) {
	resolversIntegrationTestsSource := baseUrl + "resolver"

	resp, err := http.Get(resolversIntegrationTestsSource)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("HTTP Status:", resp.Status)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	var contents []GitHubContent
	if err := json.Unmarshal(body, &contents); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	var fileNames []string

	for _, content := range contents {
		if isFile(content) && strings.HasSuffix(content.Name, ".resolvers_it_test.go") {
			fileNames = append(fileNames, content.Name)
		}
	}

	fmt.Println("File Names:")
	for _, fileName := range fileNames {
		fmt.Println(fileName)
	}
}

func getQueriesMutations(baseUrl string) []queryMutation {
	schemasSource := baseUrl + "schemas"

	resp, err := http.Get(schemasSource)
	if err != nil {
		fmt.Println("Error making the request:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("GitHub API returned a non-OK status code: %d\n", resp.StatusCode)
	}

	var contents []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		fmt.Println("Error decoding JSON response:", err)
	}

	var queryMutations []queryMutation

	fmt.Println("Files with .graphqls extension:")
	for _, file := range contents {
		if file.Type == "file" && strings.HasSuffix(file.Name, ".graphqls") { // && file.Name == "meeting.graphqls" {
			fmt.Println("\nFile:", file.Name)
			queryMutation := queryMutation{
				fileName: strings.TrimSuffix(file.Name, ".graphqls"),
			}
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
			queries := getQueryMutation(queriesSnippet, schemaContent)
			if len(queries) > 0 {
				fmt.Println("Query Names:", strings.Join(queries, ", "))
				mutableQueryMutation := &queryMutation
				mutableQueryMutation.updateQueries(queries)
			}

			mutationsSnippet := regexp.MustCompile(`extend type Mutation {([\s\S]*?)}`)
			mutations := getQueryMutation(mutationsSnippet, schemaContent)
			if len(mutations) > 0 {
				fmt.Println("Mutation Names:", strings.Join(mutations, ", "))
				mutableQueryMutation := &queryMutation
				mutableQueryMutation.updateMutations(mutations)
			}
			queryMutations = append(queryMutations, queryMutation)
		}
	}
	return queryMutations
}

func getQueryMutation(queriesMutations *regexp.Regexp, schemaContent []byte) []string {
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

func isFile(content GitHubContent) bool {
	return content.Type == "file" || content.Type == false
}

type GitHubContent struct {
	Name string      `json:"name"`
	Type interface{} `json:"type"`
}

type queryMutation struct {
	fileName  string
	queries   []string
	mutations []string
}

func (m *queryMutation) updateQueries(queries []string) {
	m.queries = queries
}

func (m *queryMutation) updateMutations(mutations []string) {
	m.mutations = mutations
}
