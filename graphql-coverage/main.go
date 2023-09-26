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
	queriesMutations := getQueriesMutations(baseUrl)
	testsForQueriesMutations := getTestsForQueriesMutations(baseUrl)
	computeCoverage(queriesMutations, testsForQueriesMutations)
	//fmt.Println("queriesMutations: ", queriesMutations)
	//fmt.Println("testsForQueriesMutations: ", testsForQueriesMutations)
}

func computeCoverage(queryMutations []queryMutation, testsForQueryMutations []testsForQueryMutation) {
	fmt.Println("\nQueries and mutations / tests count")
	queryMutationsCount := 0
	for _, queryMutation := range queryMutations {
		queryMutationsCount = queryMutationsCount + len(queryMutation.queries) + len(queryMutation.mutations)
		for _, testQueryMutation := range testsForQueryMutations {
			if testQueryMutation.fileName == queryMutation.fileName {
				fmt.Println(queryMutation.fileName,
					" - ",
					len(queryMutation.queries)+len(queryMutation.mutations),
					"/",
					len(testQueryMutation.testsForQueries)+len(testQueryMutation.testsForMutation))
			}
		}
	}

	testQueryMutationsCount := 0
	for _, testQueryMutation := range testsForQueryMutations {
		testQueryMutationsCount = testQueryMutationsCount + len(testQueryMutation.testsForQueries) + len(testQueryMutation.testsForMutation)
	}

	fmt.Println("\nTotal of queries and mutations / Total tests count: ", queryMutationsCount, "/", testQueryMutationsCount, "\n")
}

func getTestsForQueriesMutations(baseUrl string) []testsForQueryMutation {
	resolversIntegrationTestsSource := baseUrl + "resolver"
	resp, err := http.Get(resolversIntegrationTestsSource)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("HTTP Status:", resp.Status)
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil
	}

	var contents []GitHubContent
	if err := json.Unmarshal(body, &contents); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil
	}

	var testsForQueryMutations []testsForQueryMutation

	for _, content := range contents {
		if isFile(content) && strings.HasSuffix(content.Name, ".resolvers_it_test.go") { //&& (content.Name == "attachment.resolvers_it_test.go" || content.Name == "contact.resolvers_it_test.go") {
			resolversIntegrationTestsFile := baseUrl + "resolver/" + content.Name
			client := &http.Client{}
			req, _ := http.NewRequest("GET", resolversIntegrationTestsFile, nil)
			req.Header.Set("Accept", "application/vnd.github.v3.raw")

			resp, _ := client.Do(req)

			if err != nil {
				fmt.Println("Error:", err)
				return nil
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				fmt.Println("HTTP Status:", resp.Status)
				return nil
			}

			fileContent, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading file content:", err)
				return nil
			}

			testMutationsPattern := `func\s+TestMutationResolver_[A-Za-z0-9_]+\s*\(`
			testMutations := getTestQueryMutations(testMutationsPattern, fileContent)
			testQueriesPattern := `func\s+TestQueryResolver_[A-Za-z0-9_]+\s*\(`
			testQueries := getTestQueryMutations(testQueriesPattern, fileContent)

			testsForQueryMutation := testsForQueryMutation{
				fileName:         strings.TrimSuffix(content.Name, ".resolvers_it_test.go"),
				testsForQueries:  testQueries,
				testsForMutation: testMutations,
			}
			testsForQueryMutations = append(testsForQueryMutations, testsForQueryMutation)
		}
	}

	return testsForQueryMutations
}

func getTestQueryMutations(mutationsPattern string, fileContent []byte) []string {
	re := regexp.MustCompile(mutationsPattern)

	matches := re.FindAllString(string(fileContent), -1)

	var testMutations []string
	for _, match := range matches {
		testName := match[len("func ") : len(match)-1]
		testMutations = append(testMutations, testName)
	}
	return testMutations
}

func getQueriesMutations(baseUrl string) []queryMutation {
	schemasSource := baseUrl + "schemas"

	resp, err := http.Get(schemasSource)
	if err != nil {
		fmt.Println("Error making the request:", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("GitHub API returned a non-OK status code: %d\n", resp.StatusCode)
		return nil
	}

	var contents []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		fmt.Println("Error decoding JSON response:", err)
		return nil
	}

	var queryMutations []queryMutation

	//fmt.Println("Files with .graphqls extension:")
	for _, file := range contents {
		if file.Type == "file" && strings.HasSuffix(file.Name, ".graphqls") { //&& (file.Name == "attachment.graphqls" || file.Name == "contact.graphqls") {
			//fmt.Println("\nFile:", file.Name)
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
				//fmt.Println("Query Names:", strings.Join(queries, ", "))
				mutableQueryMutation := &queryMutation
				mutableQueryMutation.updateQueries(queries)
			}

			mutationsSnippet := regexp.MustCompile(`extend type Mutation {([\s\S]*?)}`)
			mutations := getQueryMutation(mutationsSnippet, schemaContent)
			if len(mutations) > 0 {
				//fmt.Println("Mutation Names:", strings.Join(mutations, ", "))
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

type testsForQueryMutation struct {
	fileName         string
	testsForQueries  []string
	testsForMutation []string
}
