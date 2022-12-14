package main

import (
	"context"
	"os"
	"testing"
)

func mockFindPulls() [][]string {
	return [][]string{
		{"testorg", "testrepo", "1234", "Test PR", "01/01/2020", "ABC-123;DEF-456"},
	}
}

func TestAuthenticate(t *testing.T) {
	ctx := context.Background()
	_, err := authenticate(ctx)
	if err != nil {
		t.Errorf("Authentication failed: %v", err)
	}
}

func TestFindPulls(t *testing.T) {
	os.Setenv("ORG_NAME", "testorg")
	os.Setenv("REPO_NAME", "testrepo")

	pullsList := mockFindPulls()
	if len(pullsList) == 0 {
		t.Errorf("No pulls found")
	}
}

func TestParseBugTrackingTicket(t *testing.T) {
	input := []string{"test ticket ABC-123, ABC-123 test DEF-456"}
	uniqueIDs := ParseBugTrackingTicket(input)
	if uniqueIDs != "ABC-123;DEF-456" {
		t.Errorf("Test failed, expected ABC-123;DEF-456, got %s", uniqueIDs)
	}
}

func TestParseBugTrackingTickeWrongFormat(t *testing.T) {
	input := []string{"test ticket A-123 DeF-456, DEF-789"}
	uniqueIDs := ParseBugTrackingTicket(input)
	if uniqueIDs != "DEF-789" {
		t.Errorf("Test failed, expected DEF-789, got %s", uniqueIDs)
	}
}

func TestWriteToCSV(t *testing.T) {
	pullsList := [][]string{
		{"testorg", "testrepo", "1234", "Test PR", "01/01/2020", "ABC-123;DEF-456"},
	}
	err := writeToCSV(pullsList)
	if err != nil {
		t.Errorf("Error writing to CSV: %v", err)
	}
	if _, err := os.Stat("bugs_in_prs.csv"); os.IsNotExist(err) {
		t.Errorf("Test failed, CSV file not found")
	}
}
