package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

type GitHubClient struct {
	client *github.Client
}

type GitHubData struct {
	PullRequests []PullRequest
	Issues       []Issue
	CodeReviews  []CodeReview
}

type PullRequest struct {
	Number       int
	Title        string
	URL          string
	State        string
	CreatedAt    time.Time
	MergedAt     *time.Time
	Repo         string
	Commits      int
	Additions    int
	Deletions    int
	ChangedFiles int
}

type Issue struct {
	Number    int
	Title     string
	URL       string
	State     string
	CreatedAt time.Time
	ClosedAt  *time.Time
	Repo      string
}

type CodeReview struct {
	PRNumber  int
	PRTitle   string
	URL       string
	State     string
	CreatedAt time.Time
	Repo      string
}

func NewGitHubClient(token string) *GitHubClient {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	return &GitHubClient{
		client: github.NewClient(tc),
	}
}

func (g *GitHubClient) FetchContributions(username string, startDate, endDate time.Time) (*GitHubData, error) {
	ctx := context.Background()
	data := &GitHubData{
		PullRequests: []PullRequest{},
		Issues:       []Issue{},
		CodeReviews:  []CodeReview{},
	}

	// Fetch Pull Requests
	fmt.Println("  - Fetching pull requests...")
	prs, err := g.fetchPullRequests(ctx, username, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("fetching pull requests: %w", err)
	}
	data.PullRequests = prs

	// Fetch Issues
	fmt.Println("  - Fetching issues...")
	issues, err := g.fetchIssues(ctx, username, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("fetching issues: %w", err)
	}
	data.Issues = issues

	// Fetch Code Reviews
	fmt.Println("  - Fetching code reviews...")
	reviews, err := g.fetchCodeReviews(ctx, username, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("fetching code reviews: %w", err)
	}
	data.CodeReviews = reviews

	return data, nil
}

func (g *GitHubClient) fetchPullRequests(ctx context.Context, username string, startDate, endDate time.Time) ([]PullRequest, error) {
	query := fmt.Sprintf("author:%s type:pr created:%s..%s",
		username,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	)

	opts := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allPRs []PullRequest
	for {
		result, resp, err := g.client.Search.Issues(ctx, query, opts)
		if err != nil {
			return nil, err
		}

		for _, issue := range result.Issues {
			repoName := extractRepo(*issue.HTMLURL)
			owner, repo := splitRepoName(repoName)

			pr := PullRequest{
				Number:    *issue.Number,
				Title:     *issue.Title,
				URL:       *issue.HTMLURL,
				State:     *issue.State,
				CreatedAt: issue.CreatedAt.Time,
				Repo:      repoName,
			}

			if issue.PullRequestLinks != nil && issue.ClosedAt != nil {
				mergedAt := issue.ClosedAt.Time
				pr.MergedAt = &mergedAt
			}

			// Fetch detailed PR info
			if owner != "" && repo != "" {
				prDetail, _, err := g.client.PullRequests.Get(ctx, owner, repo, *issue.Number)
				if err == nil && prDetail != nil {
					pr.Commits = safeInt(prDetail.Commits)
					pr.Additions = safeInt(prDetail.Additions)
					pr.Deletions = safeInt(prDetail.Deletions)
					pr.ChangedFiles = safeInt(prDetail.ChangedFiles)
				}
			}

			allPRs = append(allPRs, pr)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allPRs, nil
}

func (g *GitHubClient) fetchIssues(ctx context.Context, username string, startDate, endDate time.Time) ([]Issue, error) {
	// Fetch created issues
	query := fmt.Sprintf("author:%s type:issue created:%s..%s",
		username,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	)

	opts := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allIssues []Issue
	for {
		result, resp, err := g.client.Search.Issues(ctx, query, opts)
		if err != nil {
			return nil, err
		}

		for _, issue := range result.Issues {
			if issue.PullRequestLinks != nil {
				continue // Skip PRs
			}

			var closedAt *time.Time
			if issue.ClosedAt != nil {
				closed := issue.ClosedAt.Time
				closedAt = &closed
			}

			allIssues = append(allIssues, Issue{
				Number:    *issue.Number,
				Title:     *issue.Title,
				URL:       *issue.HTMLURL,
				State:     *issue.State,
				CreatedAt: issue.CreatedAt.Time,
				ClosedAt:  closedAt,
				Repo:      extractRepo(*issue.HTMLURL),
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	// Also fetch issues where the user participated (commented)
	participatedQuery := fmt.Sprintf("involves:%s type:issue updated:%s..%s -author:%s",
		username,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
		username,
	)

	opts = &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		result, resp, err := g.client.Search.Issues(ctx, participatedQuery, opts)
		if err != nil {
			return nil, err
		}

		for _, issue := range result.Issues {
			if issue.PullRequestLinks != nil {
				continue // Skip PRs
			}

			var closedAt *time.Time
			if issue.ClosedAt != nil {
				closed := issue.ClosedAt.Time
				closedAt = &closed
			}

			allIssues = append(allIssues, Issue{
				Number:    *issue.Number,
				Title:     *issue.Title,
				URL:       *issue.HTMLURL,
				State:     *issue.State,
				CreatedAt: issue.CreatedAt.Time,
				ClosedAt:  closedAt,
				Repo:      extractRepo(*issue.HTMLURL),
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allIssues, nil
}

func (g *GitHubClient) fetchCodeReviews(ctx context.Context, username string, startDate, endDate time.Time) ([]CodeReview, error) {
	query := fmt.Sprintf("reviewed-by:%s type:pr reviewed:%s..%s -author:%s",
		username,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
		username,
	)

	opts := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allReviews []CodeReview
	for {
		result, resp, err := g.client.Search.Issues(ctx, query, opts)
		if err != nil {
			return nil, err
		}

		for _, issue := range result.Issues {
			allReviews = append(allReviews, CodeReview{
				PRNumber:  *issue.Number,
				PRTitle:   *issue.Title,
				URL:       *issue.HTMLURL,
				State:     *issue.State,
				CreatedAt: issue.CreatedAt.Time,
				Repo:      extractRepo(*issue.HTMLURL),
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allReviews, nil
}

func extractRepo(url string) string {
	// Extract repo name from URL like https://github.com/owner/repo/pull/123
	// This is a simple extraction, you might want to make it more robust
	var repo string
	fmt.Sscanf(url, "https://github.com/%s", &repo)
	// Remove the /pull/123 or /issues/123 part
	for i := 0; i < len(repo); i++ {
		if repo[i] == '/' {
			count := 0
			for j := 0; j <= i; j++ {
				if repo[j] == '/' {
					count++
				}
			}
			if count == 2 {
				repo = repo[:i]
				break
			}
		}
	}
	return repo
}

func splitRepoName(repoName string) (owner, repo string) {
	// Split "owner/repo" into separate parts
	parts := []rune(repoName)
	slashIdx := -1
	for i, r := range parts {
		if r == '/' {
			slashIdx = i
			break
		}
	}
	if slashIdx > 0 && slashIdx < len(parts)-1 {
		owner = string(parts[:slashIdx])
		repo = string(parts[slashIdx+1:])
	}
	return
}

func safeInt(val *int) int {
	if val == nil {
		return 0
	}
	return *val
}
