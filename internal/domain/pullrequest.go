package domain

import (
	"strings"
	"time"
)

// PullRequest represents an Azure DevOps pull request.
type PullRequest struct {
	PullRequestID int        `json:"pullRequestId"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Status        string     `json:"status"` // active, abandoned, completed
	CreatedBy     Identity   `json:"createdBy"`
	CreationDate  time.Time  `json:"creationDate"`
	SourceRefName string     `json:"sourceRefName"`
	TargetRefName string     `json:"targetRefName"`
	MergeStatus   string     `json:"mergeStatus"`
	IsDraft       bool       `json:"isDraft"`
	URL           string     `json:"url"`
	Repository    RepoRef    `json:"repository"`
	Reviewers     []Reviewer `json:"reviewers"`
}

// SourceBranch returns the source branch name without refs/heads/ prefix.
func (pr *PullRequest) SourceBranch() string {
	return strings.TrimPrefix(pr.SourceRefName, "refs/heads/")
}

// TargetBranch returns the target branch name without refs/heads/ prefix.
func (pr *PullRequest) TargetBranch() string {
	return strings.TrimPrefix(pr.TargetRefName, "refs/heads/")
}

// IsActive returns true if the PR is still active.
func (pr *PullRequest) IsActive() bool {
	return pr.Status == "active"
}

// RepoRef is a reference to a repository in a PR.
type RepoRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Reviewer represents a PR reviewer.
type Reviewer struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	UniqueName  string `json:"uniqueName"`
	Vote        int    `json:"vote"` // 10=approved, 5=approved with suggestions, 0=no vote, -5=waiting, -10=rejected
}

// VoteStatus returns a human-readable vote status.
func (r *Reviewer) VoteStatus() string {
	switch r.Vote {
	case 10:
		return "Approved"
	case 5:
		return "Approved with suggestions"
	case 0:
		return "No vote"
	case -5:
		return "Waiting for author"
	case -10:
		return "Rejected"
	default:
		return "Unknown"
	}
}

// VoteIcon returns an emoji for the vote status.
func (r *Reviewer) VoteIcon() string {
	switch r.Vote {
	case 10:
		return "✅"
	case 5:
		return "👍"
	case 0:
		return "⏳"
	case -5:
		return "⏸️"
	case -10:
		return "❌"
	default:
		return "•"
	}
}

// PullRequestList is the response from listing pull requests.
type PullRequestList struct {
	Count int           `json:"count"`
	Value []PullRequest `json:"value"`
}
