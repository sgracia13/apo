package domain

import (
	"fmt"
	"strings"
)

// Repository represents an Azure DevOps Git repository.
type Repository struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	URL           string  `json:"url"`
	RemoteURL     string  `json:"remoteUrl"`
	SSHURL        string  `json:"sshUrl"`
	WebURL        string  `json:"webUrl"`
	DefaultBranch string  `json:"defaultBranch"`
	Size          int64   `json:"size"`
	Project       Project `json:"project"`
}

// DefaultBranchName returns the branch name without refs/heads/ prefix.
func (r *Repository) DefaultBranchName() string {
	return strings.TrimPrefix(r.DefaultBranch, "refs/heads/")
}

// SizeFormatted returns the repository size in human-readable format.
func (r *Repository) SizeFormatted() string {
	if r.Size < 1024 {
		return fmt.Sprintf("%d KB", r.Size)
	}
	sizeMB := float64(r.Size) / 1024
	if sizeMB < 1024 {
		return fmt.Sprintf("%.1f MB", sizeMB)
	}
	sizeGB := sizeMB / 1024
	return fmt.Sprintf("%.2f GB", sizeGB)
}

// RepositoryList is the response from listing repositories.
type RepositoryList struct {
	Count int          `json:"count"`
	Value []Repository `json:"value"`
}
