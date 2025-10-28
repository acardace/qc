package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/acardace/qc/internal/clients"
	"github.com/acardace/qc/internal/config"
	"github.com/acardace/qc/internal/report"
)

func main() {
	quarter := flag.String("quarter", "", "Quarter to generate report for (Q1, Q2, Q3, Q4)")
	year := flag.Int("year", time.Now().Year(), "Year for the quarter (default: current year)")
	associate := flag.String("associate", "", "Associate name from config file")
	configFile := flag.String("config", "config.yaml", "Path to config file")
	outputDir := flag.String("output", "reports", "Output directory for reports")

	flag.Parse()

	if *quarter == "" {
		fmt.Println("Usage: qc --quarter <Q1|Q2|Q3|Q4> [--associate <name>] [--year <year>] [--config <path>] [--output <dir>]")
		fmt.Println("\nIf --associate is not specified, reports will be generated for all associates in the config file.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Get quarter date range
	startDate, endDate, err := getQuarterDates(*quarter, *year)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Load config
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Determine which associates to process
	var associatesToProcess []string
	if *associate != "" {
		// Single associate specified
		if _, ok := cfg.Associates[*associate]; !ok {
			log.Fatalf("Associate '%s' not found in config file", *associate)
		}
		associatesToProcess = []string{*associate}
	} else {
		// No associate specified - process all
		for name := range cfg.Associates {
			associatesToProcess = append(associatesToProcess, name)
		}
		if len(associatesToProcess) == 0 {
			log.Fatalf("No associates found in config file")
		}
		fmt.Printf("No associate specified - generating reports for all %d associates\n", len(associatesToProcess))
	}

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// Process each associate
	for i, assocName := range associatesToProcess {
		associateInfo := cfg.Associates[assocName]

		fmt.Printf("\n[%d/%d] Generating report for %s (%s %d: %s to %s)...\n",
			i+1, len(associatesToProcess), assocName, *quarter, *year,
			startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

		// Fetch Jira data
		fmt.Println("  Fetching Jira data...")
		jiraClient := clients.NewJiraClient(cfg.Jira.URL, cfg.Jira.Token)
		jiraIssues, err := jiraClient.FetchCompletedIssues(associateInfo.JiraUsername, startDate, endDate)
		if err != nil {
			log.Printf("  Warning: Error fetching Jira data for %s: %v", assocName, err)
			continue
		}

		// Fetch GitHub data
		fmt.Println("  Fetching GitHub data...")
		githubClient := clients.NewGitHubClient(cfg.GitHub.Token)
		githubData, err := githubClient.FetchContributions(associateInfo.GitHubUsername, startDate, endDate)
		if err != nil {
			log.Printf("  Warning: Error fetching GitHub data for %s: %v", assocName, err)
			continue
		}

		// Generate report
		fmt.Println("  Generating HTML report...")
		reportHTML := report.Generate(assocName, *quarter, *year, startDate, endDate, cfg.Jira.URL, jiraIssues, githubData)

		// Save report
		outputFile := fmt.Sprintf("%s/%s_%s_%d.html", *outputDir, assocName, *quarter, *year)
		if err := os.WriteFile(outputFile, []byte(reportHTML), 0644); err != nil {
			log.Printf("  Warning: Error writing report for %s: %v", assocName, err)
			continue
		}

		fmt.Printf("  ✓ Report generated: %s\n", outputFile)
	}

	fmt.Printf("\n✓ All reports generated successfully in %s/\n", *outputDir)
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
