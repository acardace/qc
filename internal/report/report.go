package report

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/acardace/qc/internal/clients"
)

type ReportData struct {
	AssociateName string
	Quarter       string
	Year          int
	StartDate     string
	EndDate       string
	GeneratedAt   string
	JiraURL       string

	// Jira Stats
	JiraIssues      []clients.JiraIssue
	TotalJiraIssues int

	// GitHub Stats
	PullRequests       []clients.PullRequest
	Issues             []clients.Issue
	CodeReviews        []clients.CodeReview
	TotalPRs           int
	TotalIssues        int
	TotalCodeReviews   int
	MergedPRs          int
	ClosedIssues       int
	UniqueReposWorked  int
	TotalCommits       int
	TotalLinesAdded    int
	TotalLinesDeleted  int
	TotalStoryPoints   float64
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Quarterly Connection Report - {{.AssociateName}} - {{.Quarter}} {{.Year}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            line-height: 1.6;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            border-radius: 8px;
            margin-bottom: 30px;
        }
        .header h1 {
            margin: 0 0 10px 0;
        }
        .header p {
            margin: 5px 0;
            opacity: 0.9;
        }
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .stat-card {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            text-align: center;
        }
        .stat-number {
            font-size: 36px;
            font-weight: bold;
            color: #667eea;
            margin: 10px 0;
        }
        .stat-label {
            color: #666;
            font-size: 14px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        .section {
            background: white;
            padding: 25px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            margin-bottom: 20px;
        }
        .section h2 {
            color: #333;
            border-bottom: 2px solid #667eea;
            padding-bottom: 10px;
            margin-top: 0;
        }
        .item-list {
            list-style: none;
            padding: 0;
        }
        .item {
            padding: 15px;
            border-left: 3px solid #667eea;
            margin-bottom: 10px;
            background-color: #f9f9f9;
            border-radius: 4px;
        }
        .item-title {
            font-weight: bold;
            color: #333;
            margin-bottom: 5px;
        }
        .item-meta {
            font-size: 13px;
            color: #666;
        }
        .badge {
            display: inline-block;
            padding: 3px 8px;
            border-radius: 12px;
            font-size: 12px;
            margin-right: 5px;
        }
        .badge-success {
            background-color: #d4edda;
            color: #155724;
        }
        .badge-info {
            background-color: #d1ecf1;
            color: #0c5460;
        }
        .badge-warning {
            background-color: #fff3cd;
            color: #856404;
        }
        .footer {
            text-align: center;
            color: #666;
            margin-top: 30px;
            padding: 20px;
            font-size: 14px;
        }
        a {
            color: #667eea;
            text-decoration: none;
        }
        a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Quarterly Connection Report</h1>
        <p><strong>Associate:</strong> {{.AssociateName}}</p>
        <p><strong>Period:</strong> {{.Quarter}} {{.Year}} ({{.StartDate}} to {{.EndDate}})</p>
        <p><strong>Generated:</strong> {{.GeneratedAt}}</p>
    </div>

    <div class="stats-grid">
        <div class="stat-card">
            <div class="stat-label">Jira Issues Completed</div>
            <div class="stat-number">{{.TotalJiraIssues}}</div>
            {{if gt .TotalStoryPoints 0.0}}
            <div class="stat-label">{{printf "%.1f" .TotalStoryPoints}} story points</div>
            {{end}}
        </div>
        <div class="stat-card">
            <div class="stat-label">Pull Requests</div>
            <div class="stat-number">{{.TotalPRs}}</div>
            <div class="stat-label">{{.MergedPRs}} merged</div>
        </div>
        <div class="stat-card">
            <div class="stat-label">Commits</div>
            <div class="stat-number">{{.TotalCommits}}</div>
            <div class="stat-label">Across all PRs</div>
        </div>
        <div class="stat-card">
            <div class="stat-label">Code Changes</div>
            <div class="stat-number">{{.TotalLinesAdded}}</div>
            <div class="stat-label" style="color: #22863a;">+{{.TotalLinesAdded}} / <span style="color: #cb2431;">-{{.TotalLinesDeleted}}</span></div>
        </div>
        <div class="stat-card">
            <div class="stat-label">Issues</div>
            <div class="stat-number">{{.TotalIssues}}</div>
            <div class="stat-label">Created or participated</div>
        </div>
        <div class="stat-card">
            <div class="stat-label">Code Reviews</div>
            <div class="stat-number">{{.TotalCodeReviews}}</div>
        </div>
        <div class="stat-card">
            <div class="stat-label">Repositories</div>
            <div class="stat-number">{{.UniqueReposWorked}}</div>
            <div class="stat-label">Worked on</div>
        </div>
    </div>

    {{if .JiraIssues}}
    <div class="section">
        <h2>Jira Accomplishments ({{.TotalJiraIssues}})</h2>
        <ul class="item-list">
        {{range .JiraIssues}}
            <li class="item">
                <div class="item-title">
                    <a href="{{$.JiraURL}}/browse/{{.Key}}" target="_blank">{{.Key}}</a> - {{.Summary}}
                </div>
                <div class="item-meta">
                    <span class="badge badge-success">{{.Status}}</span>
                    <span class="badge badge-info">{{.Type}}</span>
                    {{if .Priority}}<span class="badge badge-warning">{{.Priority}}</span>{{end}}
                    {{if .HasStoryPoints}}<span class="badge badge-info">{{printf "%.1f SP" .StoryPoints}}</span>{{end}}
                    {{if not .Resolved.IsZero}}Resolved: {{.Resolved.Format "2006-01-02"}}{{end}}
                </div>
            </li>
        {{end}}
        </ul>
    </div>
    {{end}}

    {{if .PullRequests}}
    <div class="section">
        <h2>Pull Requests ({{.TotalPRs}})</h2>
        <ul class="item-list">
        {{range .PullRequests}}
            <li class="item">
                <div class="item-title">
                    <a href="{{.URL}}" target="_blank">#{{.Number}}</a> - {{.Title}}
                </div>
                <div class="item-meta">
                    {{if .MergedAt}}
                    <span class="badge badge-success">Merged</span>
                    {{else}}
                    <span class="badge badge-warning">{{.State}}</span>
                    {{end}}
                    <span class="badge badge-info">{{.Repo}}</span>
                    {{if gt .Commits 0}}<span class="badge badge-info">{{.Commits}} commits</span>{{end}}
                    {{if gt .Additions 0}}<span style="color: #22863a;">+{{.Additions}}</span>{{end}}
                    {{if gt .Deletions 0}}<span style="color: #cb2431;">-{{.Deletions}}</span>{{end}}
                    {{if gt .ChangedFiles 0}}<span class="badge badge-info">{{.ChangedFiles}} files</span>{{end}}
                    Created: {{.CreatedAt.Format "2006-01-02"}}
                </div>
            </li>
        {{end}}
        </ul>
    </div>
    {{end}}

    {{if .Issues}}
    <div class="section">
        <h2>GitHub Issues ({{.TotalIssues}})</h2>
        <ul class="item-list">
        {{range .Issues}}
            <li class="item">
                <div class="item-title">
                    <a href="{{.URL}}" target="_blank">#{{.Number}}</a> - {{.Title}}
                </div>
                <div class="item-meta">
                    {{if .ClosedAt}}
                    <span class="badge badge-success">Closed</span>
                    {{else}}
                    <span class="badge badge-warning">{{.State}}</span>
                    {{end}}
                    <span class="badge badge-info">{{.Repo}}</span>
                    Created: {{.CreatedAt.Format "2006-01-02"}}
                </div>
            </li>
        {{end}}
        </ul>
    </div>
    {{end}}

    {{if .CodeReviews}}
    <div class="section">
        <h2>Code Reviews ({{.TotalCodeReviews}})</h2>
        <ul class="item-list">
        {{range .CodeReviews}}
            <li class="item">
                <div class="item-title">
                    <a href="{{.URL}}" target="_blank">#{{.PRNumber}}</a> - {{.PRTitle}}
                </div>
                <div class="item-meta">
                    <span class="badge badge-info">{{.Repo}}</span>
                    Reviewed: {{.CreatedAt.Format "2006-01-02"}}
                </div>
            </li>
        {{end}}
        </ul>
    </div>
    {{end}}

    <div class="footer">
        <p>This report was automatically generated by the Quarterly Connection tool.</p>
    </div>
</body>
</html>`

func Generate(associateName, quarter string, year int, startDate, endDate time.Time, jiraURL string, jiraIssues []clients.JiraIssue, githubData *clients.GitHubData) string {
	// Count merged PRs
	mergedPRs := 0
	totalCommits := 0
	totalLinesAdded := 0
	totalLinesDeleted := 0
	for _, pr := range githubData.PullRequests {
		if pr.MergedAt != nil {
			mergedPRs++
		}
		totalCommits += pr.Commits
		totalLinesAdded += pr.Additions
		totalLinesDeleted += pr.Deletions
	}

	// Count closed issues
	closedIssues := 0
	for _, issue := range githubData.Issues {
		if issue.ClosedAt != nil {
			closedIssues++
		}
	}

	// Count story points
	totalStoryPoints := 0.0
	for _, issue := range jiraIssues {
		if issue.HasStoryPoints {
			totalStoryPoints += issue.StoryPoints
		}
	}

	// Count unique repositories
	repoMap := make(map[string]bool)
	for _, pr := range githubData.PullRequests {
		repoMap[pr.Repo] = true
	}
	for _, issue := range githubData.Issues {
		repoMap[issue.Repo] = true
	}
	for _, review := range githubData.CodeReviews {
		repoMap[review.Repo] = true
	}

	data := ReportData{
		AssociateName:      associateName,
		Quarter:            quarter,
		Year:               year,
		StartDate:          startDate.Format("2006-01-02"),
		EndDate:            endDate.Format("2006-01-02"),
		GeneratedAt:        time.Now().Format("2006-01-02 15:04:05"),
		JiraURL:            jiraURL,
		JiraIssues:         jiraIssues,
		TotalJiraIssues:    len(jiraIssues),
		PullRequests:       githubData.PullRequests,
		Issues:             githubData.Issues,
		CodeReviews:        githubData.CodeReviews,
		TotalPRs:           len(githubData.PullRequests),
		TotalIssues:        len(githubData.Issues),
		TotalCodeReviews:   len(githubData.CodeReviews),
		MergedPRs:          mergedPRs,
		ClosedIssues:       closedIssues,
		UniqueReposWorked:  len(repoMap),
		TotalCommits:       totalCommits,
		TotalLinesAdded:    totalLinesAdded,
		TotalLinesDeleted:  totalLinesDeleted,
		TotalStoryPoints:   totalStoryPoints,
	}

	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return fmt.Sprintf("Error parsing template: %v", err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return fmt.Sprintf("Error executing template: %v", err)
	}

	return result.String()
}
