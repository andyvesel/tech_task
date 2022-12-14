package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/google/go-github/v48/github"
	"github.com/joho/godotenv"
)

// authenticate authenticates with GitHub using a GitHub API token provided in a .env file
func authenticate(ctx context.Context) (*github.Client, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	token := os.Getenv("GITHUB_API_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_API_TOKEN environment variable not set")
	}

	client := &http.Client{
		Transport: &github.BasicAuthTransport{
			Username: token,
			Password: "x-oauth-basic",
		},
	}

	githubClient := github.NewClient(client)

	_, _, err = githubClient.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("error verifying token: %w", err)
	}

	fmt.Println("User authenticated")
	return githubClient, nil
}

// FindPulls returns a list of all PRs in a given repo within a START_DATE and END_DATE range
// if no END_DATE is provided, it will default to the current date
func FindPulls() [][]string {
	ctx := context.Background()
	pullsList := [][]string{}
	client, err := authenticate(ctx)
	if err != nil {
		fmt.Printf("error authenticating with GitHub: %v ", err)
		os.Exit(1)
	}
	client.Users.Get(ctx, "")

	startDate, _ := time.Parse("01/02/2006", os.Getenv("START_DATE"))
	endDate, _ := time.Parse("01/02/2006", os.Getenv("END_DATE"))
	if os.Getenv("END_DATE") == "" {
		endDate = time.Now()
	}

	opt := &github.PullRequestListOptions{
		State: "all",
	}

	prs, _, err := client.PullRequests.List(ctx, os.Getenv("ORG_NAME"), os.Getenv("REPO_NAME"), opt)
	if err != nil {
		fmt.Printf("error getting list of PRs: %v\n", err)
		os.Exit(1)
	}

	for _, pr := range prs {
		if pr.CreatedAt.After(startDate) && pr.CreatedAt.Before(endDate) {
			pullsList = append(pullsList,
				[]string{
					os.Getenv("ORG_NAME"),
					os.Getenv("REPO_NAME"),
					fmt.Sprintf("%d", *pr.Number),
					pr.GetTitle(),
					pr.CreatedAt.Format("01/02/2006"),
					pr.GetBody()},
			)
		}
	}
	for _, prBody := range pullsList {
		uniqueIDs := ParseBugTrackingTicket(prBody)
		prBody[5] = uniqueIDs
	}
	writeToCSV(pullsList)
	return pullsList
}

// ParseBugTrackingTicket returns a list of unique bug tracking ticket IDs from a
// PR body separated with semicolons
func ParseBugTrackingTicket(input []string) string {
	var bugIDs []string
	var uniqueIDs string
	for _, el := range input {
		match := regexp.MustCompile("[A-Z]{2,5}-\\d+").FindAllString(el, -1)
		bugIDs = append(bugIDs, match...)
	}
	uniqueMap := make(map[string]bool)

	for _, match := range bugIDs {
		if _, ok := uniqueMap[match]; !ok {
			uniqueMap[match] = true
			uniqueIDs = uniqueIDs + match + ";"
		}
	}
	fmt.Printf("Unique IDs: %s parsed\n", uniqueIDs)
	return uniqueIDs[:len(uniqueIDs)-1]
}

// writeToCSV writes a list of PRs to a CSV file
// it creates a default file called bugs_in_prs.csv if no OUTPUT_FILE is specified
func writeToCSV(pullsList [][]string) error {
	var file *os.File
	filePath := os.Getenv("OUTPUT_FILE")
	if filePath == "" {
		file, _ = os.Create("bugs_in_prs.csv")
	} else {
		file, _ = os.Create(filePath)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.WriteAll(pullsList); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created CSV file: %s", file.Name())
	return nil
}

func main() {
	FindPulls()
}
