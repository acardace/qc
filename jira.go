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
	Key     string `json:"key"`
	Summary string
	Status  string
	Type    string
	Created time.Time
	Updated time.Time
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
			Created string `json:"created"`
			Updated string `json:"updated"`
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
	params.Add("fields", "summary,status,issuetype,created,updated")

	apiURL := fmt.Sprintf("%s/rest/api/3/search?%s", j.baseURL, params.Encode())

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

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Jira API error (status %d): %s", resp.StatusCode, string(body))
	}

	var searchResp jiraSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	issues := make([]JiraIssue, 0, len(searchResp.Issues))
	for _, issue := range searchResp.Issues {
		created, _ := time.Parse(time.RFC3339, issue.Fields.Created)
		updated, _ := time.Parse(time.RFC3339, issue.Fields.Updated)

		issues = append(issues, JiraIssue{
			Key:     issue.Key,
			Summary: issue.Fields.Summary,
			Status:  issue.Fields.Status.Name,
			Type:    issue.Fields.IssueType.Name,
			Created: created,
			Updated: updated,
		})
	}

	return issues, nil
}
