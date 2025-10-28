package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type JiraClient struct {
	baseURL string
	token   string
	client  *http.Client
}

type JiraIssue struct {
	Key         string `json:"key"`
	Summary     string
	Status      string
	Type        string
	Priority    string
	StoryPoints *float64
	Assignee    string
	Reporter    string
	Created     time.Time
	Updated     time.Time
	Resolved    time.Time
}

type jiraSearchResponse struct {
	Issues []struct {
		Key    string `json:"key"`
		Fields struct {
			Summary string `json:"summary"`
			Status  struct {
				Name string `json:"name"`
			} `json:"status"`
			IssueType struct {
				Name string `json:"name"`
			} `json:"issuetype"`
			Priority struct {
				Name string `json:"name"`
			} `json:"priority"`
			Assignee struct {
				DisplayName string `json:"displayName"`
			} `json:"assignee"`
			Reporter struct {
				DisplayName string `json:"displayName"`
			} `json:"reporter"`
			StoryPoints      interface{} `json:"customfield_12310243"` // Common story points field
			Created          string      `json:"created"`
			Updated          string      `json:"updated"`
			ResolutionDate   string      `json:"resolutiondate"`
		} `json:"fields"`
	} `json:"issues"`
	Total int `json:"total"`
}

func NewJiraClient(baseURL, token string) *JiraClient {
	return &JiraClient{
		baseURL: baseURL,
		token:   token,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (j *JiraClient) FetchCompletedIssues(username string, startDate, endDate time.Time) ([]JiraIssue, error) {
	// JQL query to find issues completed by the user in the date range
	jql := fmt.Sprintf(
		`assignee = "%s" AND status in (Done, Closed, Resolved) AND resolved >= "%s" AND resolved <= "%s" ORDER BY resolved DESC`,
		username,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	)

	params := url.Values{}
	params.Add("jql", jql)
	params.Add("maxResults", "1000")
	params.Add("fields", "summary,status,issuetype,priority,assignee,reporter,customfield_12310243,created,updated,resolutiondate")

	apiURL := fmt.Sprintf("%s/rest/api/2/search?%s", j.baseURL, params.Encode())

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+j.token)
	req.Header.Set("Accept", "application/json")

	resp, err := j.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Jira API error (status %d): %s", resp.StatusCode, string(body))
	}

	var searchResp jiraSearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		// Show the first 200 characters of the response to help debug
		preview := string(body)
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		return nil, fmt.Errorf("decoding JSON response: %w\nResponse preview: %s", err, preview)
	}

	issues := make([]JiraIssue, 0, len(searchResp.Issues))
	for _, issue := range searchResp.Issues {
		created, _ := time.Parse(time.RFC3339, issue.Fields.Created)
		updated, _ := time.Parse(time.RFC3339, issue.Fields.Updated)
		resolved, _ := time.Parse(time.RFC3339, issue.Fields.ResolutionDate)

		// Parse story points - could be float64 or nil
		var storyPoints *float64
		if issue.Fields.StoryPoints != nil {
			switch v := issue.Fields.StoryPoints.(type) {
			case float64:
				storyPoints = &v
			case int:
				fp := float64(v)
				storyPoints = &fp
			}
		}

		issues = append(issues, JiraIssue{
			Key:         issue.Key,
			Summary:     issue.Fields.Summary,
			Status:      issue.Fields.Status.Name,
			Type:        issue.Fields.IssueType.Name,
			Priority:    issue.Fields.Priority.Name,
			StoryPoints: storyPoints,
			Assignee:    issue.Fields.Assignee.DisplayName,
			Reporter:    issue.Fields.Reporter.DisplayName,
			Created:     created,
			Updated:     updated,
			Resolved:    resolved,
		})
	}

	return issues, nil
}
