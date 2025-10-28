package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	quarter := flag.String("quarter", "", "Quarter to generate report for (Q1, Q2, Q3, Q4)")
	year := flag.Int("year", time.Now().Year(), "Year for the quarter (default: current year)")
	associate := flag.String("associate", "", "Associate name from config file")
	configFile := flag.String("config", "config.yaml", "Path to config file")
	outputDir := flag.String("output", "reports", "Output directory for reports")

	flag.Parse()

	if *quarter == "" || *associate == "" {
		fmt.Println("Usage: qc --quarter <Q1|Q2|Q3|Q4> --associate <name> [--year <year>] [--config <path>] [--output <dir>]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Get quarter date range
	startDate, endDate, err := getQuarterDates(*quarter, *year)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Load config
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Get associate info from config
	associateInfo, ok := config.Associates[*associate]
	if !ok {
		log.Fatalf("Associate '%s' not found in config file", *associate)
	}

	fmt.Printf("Generating report for %s (%s %d: %s to %s)...\n",
		*associate, *quarter, *year, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// Fetch Jira data
	fmt.Println("Fetching Jira data...")
	jiraClient := NewJiraClient(config.Jira.URL, config.Jira.Token)
	jiraIssues, err := jiraClient.FetchCompletedIssues(associateInfo.JiraUsername, startDate, endDate)
	if err != nil {
		log.Fatalf("Error fetching Jira data: %v", err)
	}

	// Fetch GitHub data
	fmt.Println("Fetching GitHub data...")
	githubClient := NewGitHubClient(config.GitHub.Token)
	githubData, err := githubClient.FetchContributions(associateInfo.GitHubUsername, startDate, endDate)
	if err != nil {
		log.Fatalf("Error fetching GitHub data: %v", err)
	}

	// Generate report
	fmt.Println("Generating HTML report...")
	report := GenerateReport(*associate, *quarter, *year, startDate, endDate, config.Jira.URL, jiraIssues, githubData)

	// Save report
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	outputFile := fmt.Sprintf("%s/%s_%s_%d.html", *outputDir, *associate, *quarter, *year)
	if err := os.WriteFile(outputFile, []byte(report), 0644); err != nil {
		log.Fatalf("Error writing report: %v", err)
	}

	fmt.Printf("Report generated successfully: %s\n", outputFile)
}

func getQuarterDates(quarter string, year int) (time.Time, time.Time, error) {
	var startMonth, endMonth time.Month
	var endDay int

	switch quarter {
	case "Q1":
		startMonth = time.January
		endMonth = time.March
		endDay = 31
	case "Q2":
		startMonth = time.April
		endMonth = time.June
		endDay = 30
	case "Q3":
		startMonth = time.July
		endMonth = time.September
		endDay = 30
	case "Q4":
		startMonth = time.October
		endMonth = time.December
		endDay = 31
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("invalid quarter: %s (must be Q1, Q2, Q3, or Q4)", quarter)
	}

	startDate := time.Date(year, startMonth, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, endMonth, endDay, 23, 59, 59, 0, time.UTC)

	return startDate, endDate, nil
}
