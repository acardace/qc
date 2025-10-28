# Quarterly Connection (QC)

A CLI tool to generate quarterly performance reports for associates by aggregating data from Jira and GitHub.

## Features

- **Jira Integration**: Fetches completed issues with detailed metrics:
  - Story points
  - Priority
  - Issue type
  - Resolution date
  - Reporter and assignee information
- **GitHub Integration**: Retrieves comprehensive contribution data:
  - Pull requests (with commits, lines changed, files modified)
  - Issues (created and participated in)
  - Code reviews performed
- **Rich HTML Reports**: Beautiful, responsive reports with:
  - Summary statistics (story points, commits, code changes)
  - Detailed breakdowns of all work items
  - Visual badges for status, priority, and metrics
- **Hardcoded Quarters**: Q1-Q4 with automatic date range calculation
- **Multi-user Support**: Configurable via YAML for multiple associates

## Quarters

- **Q1**: January 1 - March 31
- **Q2**: April 1 - June 30
- **Q3**: July 1 - September 30
- **Q4**: October 1 - December 31

## Prerequisites

- Go 1.21 or higher
- Jira Personal Access Token (for Jira Data Center/Server)
- GitHub Personal Access Token

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd qc

# Build the binary
go build -o qc

# Optional: Install globally
go install
```

## Configuration

1. Copy the example configuration file:
```bash
cp config.example.yaml config.yaml
```

2. Edit `config.yaml` with your credentials:

```yaml
jira:
  url: "https://jira.your-company.com"
  token: "your-jira-personal-access-token"

github:
  token: "your-github-personal-access-token"

associates:
  john_doe:
    jira_username: "john.doe"
    github_username: "johndoe"
    full_name: "John Doe"
```

### Getting Tokens

**Jira Personal Access Token:**
- For Jira Data Center/Server: Navigate to your profile > Personal Access Tokens > Create token
- Ensure the token has read permissions for issues

**GitHub Personal Access Token:**
- Go to GitHub Settings > Developer settings > Personal access tokens > Tokens (classic)
- Create a token with the following scopes:
  - `repo` (for private repositories)
  - `read:org` (for organization data)
  - `read:user` (for user data)

## Usage

```bash
# Generate a report for a single associate
./qc --quarter Q1 --year 2024 --associate john_doe

# Generate reports for ALL associates in the config file
./qc --quarter Q2 --year 2024

# Generate reports for all associates (current year)
./qc --quarter Q3

# Specify custom config file
./qc --quarter Q3 --associate john_doe --config /path/to/config.yaml

# Specify custom output directory
./qc --quarter Q4 --output /path/to/reports
```

### Command-line Options

- `--quarter` (required): Quarter to generate report for (Q1, Q2, Q3, or Q4)
- `--associate` (optional): Associate name from config file. If not specified, generates reports for ALL associates
- `--year` (optional): Year for the quarter (default: current year)
- `--config` (optional): Path to config file (default: config.yaml)
- `--output` (optional): Output directory for reports (default: reports)

## Output

Reports are generated as HTML files in the output directory with the naming format:
```
<associate>_<quarter>_<year>.html
```

Example: `john_doe_Q1_2024.html`

The report includes:

- **Summary Statistics**:
  - Jira issues completed (with total story points)
  - Pull requests created and merged
  - Total commits across all PRs
  - Lines of code added/deleted
  - GitHub issues created or participated in
  - Code reviews performed
  - Unique repositories worked on

- **Detailed Breakdowns**:
  - **Jira Issues**: Key, summary, status, type, priority, story points, resolution date
  - **Pull Requests**: Number, title, repo, commits, additions/deletions, files changed, merge status
  - **GitHub Issues**: Number, title, repo, state, creation/closure dates
  - **Code Reviews**: PRs reviewed with repository information

## Examples

### Single Associate

```bash
$ ./qc --quarter Q1 --year 2024 --associate john_doe
[1/1] Generating report for john_doe (Q1 2024: 2024-01-01 to 2024-03-31)...
  Fetching Jira data...
  Fetching GitHub data...
    - Fetching pull requests...
    - Fetching issues...
    - Fetching code reviews...
  Generating HTML report...
  ✓ Report generated: reports/john_doe_Q1_2024.html

✓ All reports generated successfully in reports/
```

### All Associates

```bash
$ ./qc --quarter Q2 --year 2024
No associate specified - generating reports for all 3 associates

[1/3] Generating report for john_doe (Q2 2024: 2024-04-01 to 2024-06-30)...
  Fetching Jira data...
  Fetching GitHub data...
    - Fetching pull requests...
    - Fetching issues...
    - Fetching code reviews...
  Generating HTML report...
  ✓ Report generated: reports/john_doe_Q2_2024.html

[2/3] Generating report for jane_smith (Q2 2024: 2024-04-01 to 2024-06-30)...
  Fetching Jira data...
  Fetching GitHub data...
    - Fetching pull requests...
    - Fetching issues...
    - Fetching code reviews...
  Generating HTML report...
  ✓ Report generated: reports/jane_smith_Q2_2024.html

[3/3] Generating report for bob_jones (Q2 2024: 2024-04-01 to 2024-06-30)...
  Fetching Jira data...
  Fetching GitHub data...
    - Fetching pull requests...
    - Fetching issues...
    - Fetching code reviews...
  Generating HTML report...
  ✓ Report generated: reports/bob_jones_Q2_2024.html

✓ All reports generated successfully in reports/
```

## Project Structure

```
.
├── cmd/
│   └── qc/
│       └── main.go              # CLI entry point and main logic
├── internal/
│   ├── clients/
│   │   ├── jira.go              # Jira API client
│   │   └── github.go            # GitHub API client
│   ├── config/
│   │   └── config.go            # Configuration loading
│   └── report/
│       └── report.go            # HTML report generation
├── Makefile                     # Build automation
├── config.yaml                  # Your configuration (gitignored)
├── config.example.yaml          # Example configuration
├── go.mod                       # Go module file
├── go.sum                       # Go dependencies
└── README.md                    # This file
```

## Troubleshooting

**Jira authentication fails:**
- Verify your Jira URL is correct
- Ensure your PAT has the necessary permissions
- Check that you're using Jira Data Center/Server (not Cloud)

**GitHub rate limiting:**
- GitHub has rate limits (30 requests/minute for search API)
- The tool automatically handles rate limits and retries after waiting
- You'll see messages like "Rate limit detected. Waiting..." if this occurs
- For large teams or date ranges, the tool may take several minutes to complete
- Primary rate limits reset every hour
- Secondary (abuse detection) limits may require 1-minute waits between retries

**No data returned:**
- Verify the associate's username is correct in both systems
- Check that the date range contains actual activity
- Ensure the associate has the necessary permissions

## License

MIT
