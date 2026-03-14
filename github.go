package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// ── Workflow Runs ────────────────────────────────────────────────────────────

type Run struct {
	ID         int64     `json:"databaseId"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	Conclusion string    `json:"conclusion"`
	Branch     string    `json:"headBranch"`
	Event      string    `json:"event"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func FetchRuns(limit int) ([]Run, error) {
	out, err := exec.Command(
		"gh", "run", "list",
		"--limit", fmt.Sprintf("%d", limit),
		"--json", "databaseId,name,status,conclusion,headBranch,event,createdAt,updatedAt",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("gh run list: %w", err)
	}
	var runs []Run
	if err := json.Unmarshal(out, &runs); err != nil {
		return nil, fmt.Errorf("parse runs: %w", err)
	}
	return runs, nil
}

func RunStatusIcon(r Run) string {
	if r.Status == "in_progress" {
		return "⟳"
	}
	if r.Status == "queued" || r.Status == "waiting" {
		return "◷"
	}
	switch r.Conclusion {
	case "success":
		return "✓"
	case "failure":
		return "✗"
	case "cancelled":
		return "⊘"
	case "skipped":
		return "→"
	default:
		return "?"
	}
}

func RunStatusColor(r Run) string {
	if r.Status == "in_progress" {
		return string(colorYellow)
	}
	switch r.Conclusion {
	case "success":
		return string(colorGreen)
	case "failure":
		return string(colorRed)
	default:
		return string(colorDim)
	}
}

func RunElapsed(r Run) string {
	ref := r.CreatedAt
	if r.Status == "completed" {
		ref = r.UpdatedAt
	}
	return formatDuration(time.Since(ref))
}

// ── Pull Requests ───────────────────────────────────────────────────────────

type PR struct {
	Number         int       `json:"number"`
	Title          string    `json:"title"`
	Branch         string    `json:"headRefName"`
	State          string    `json:"state"`
	IsDraft        bool      `json:"isDraft"`
	ReviewDecision string    `json:"reviewDecision"`
	Author         PRAuthor  `json:"author"`
	Checks         []PRCheck `json:"statusCheckRollup"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type PRAuthor struct {
	Login string `json:"login"`
}

type PRCheck struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
}

func FetchPRs() ([]PR, error) {
	out, err := exec.Command(
		"gh", "pr", "list",
		"--state", "open",
		"--limit", "20",
		"--json", "number,title,headRefName,state,isDraft,reviewDecision,statusCheckRollup,author,createdAt,updatedAt",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("gh pr list: %w", err)
	}
	var prs []PR
	if err := json.Unmarshal(out, &prs); err != nil {
		return nil, fmt.Errorf("parse prs: %w", err)
	}
	return prs, nil
}

func ChecksSummary(pr PR) (passed, failed, pending int) {
	for _, c := range pr.Checks {
		switch {
		case c.Conclusion == "SUCCESS" || c.Conclusion == "SKIPPED" || c.Conclusion == "NEUTRAL":
			passed++
		case c.Conclusion == "FAILURE" || c.Conclusion == "CANCELLED" || c.Conclusion == "TIMED_OUT":
			failed++
		default:
			pending++
		}
	}
	return
}

func ReviewIcon(decision string) string {
	switch decision {
	case "APPROVED":
		return "✓ approved"
	case "CHANGES_REQUESTED":
		return "✗ changes"
	case "REVIEW_REQUIRED":
		return "◷ review"
	default:
		return "- no review"
	}
}

func ReviewColor(decision string) string {
	switch decision {
	case "APPROVED":
		return string(colorGreen)
	case "CHANGES_REQUESTED":
		return string(colorRed)
	default:
		return string(colorDim)
	}
}

// ── Rate Limits ─────────────────────────────────────────────────────────────

type GitHubRateLimit struct {
	Limit     int
	Used      int
	Remaining int
	Reset     time.Time
}

type GitHubLimits struct {
	Core    GitHubRateLimit
	GraphQL GitHubRateLimit
	Search  GitHubRateLimit
	Err     error
}

type ghRateLimitResponse struct {
	Resources struct {
		Core struct {
			Limit     int   `json:"limit"`
			Used      int   `json:"used"`
			Remaining int   `json:"remaining"`
			Reset     int64 `json:"reset"`
		} `json:"core"`
		GraphQL struct {
			Limit     int   `json:"limit"`
			Used      int   `json:"used"`
			Remaining int   `json:"remaining"`
			Reset     int64 `json:"reset"`
		} `json:"graphql"`
		Search struct {
			Limit     int   `json:"limit"`
			Used      int   `json:"used"`
			Remaining int   `json:"remaining"`
			Reset     int64 `json:"reset"`
		} `json:"search"`
	} `json:"resources"`
}

func FetchGitHubLimits() GitHubLimits {
	out, err := exec.Command("gh", "api", "rate_limit").Output()
	if err != nil {
		return GitHubLimits{Err: fmt.Errorf("gh cli: %w", err)}
	}

	var resp ghRateLimitResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return GitHubLimits{Err: fmt.Errorf("parse: %w", err)}
	}

	return GitHubLimits{
		Core: GitHubRateLimit{
			Limit:     resp.Resources.Core.Limit,
			Used:      resp.Resources.Core.Used,
			Remaining: resp.Resources.Core.Remaining,
			Reset:     time.Unix(resp.Resources.Core.Reset, 0),
		},
		GraphQL: GitHubRateLimit{
			Limit:     resp.Resources.GraphQL.Limit,
			Used:      resp.Resources.GraphQL.Used,
			Remaining: resp.Resources.GraphQL.Remaining,
			Reset:     time.Unix(resp.Resources.GraphQL.Reset, 0),
		},
		Search: GitHubRateLimit{
			Limit:     resp.Resources.Search.Limit,
			Used:      resp.Resources.Search.Used,
			Remaining: resp.Resources.Search.Remaining,
			Reset:     time.Unix(resp.Resources.Search.Reset, 0),
		},
	}
}

// ── Helpers ─────────────────────────────────────────────────────────────────

func formatDuration(d time.Duration) string {
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	default:
		return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
	}
}
