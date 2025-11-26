package domain

import "time"

// Build represents an Azure DevOps build.
type Build struct {
	ID           int             `json:"id"`
	BuildNumber  string          `json:"buildNumber"`
	Status       string          `json:"status"` // notStarted, inProgress, completed
	Result       string          `json:"result"` // succeeded, failed, canceled
	QueueTime    time.Time       `json:"queueTime"`
	StartTime    time.Time       `json:"startTime"`
	FinishTime   time.Time       `json:"finishTime"`
	Definition   BuildDefinition `json:"definition"`
	RequestedBy  Identity        `json:"requestedBy"`
	SourceBranch string          `json:"sourceBranch"`
	URL          string          `json:"url"`
}

// BuildDefinition contains build definition info.
type BuildDefinition struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
	URL  string `json:"url"`
}

// BuildList is the response from listing builds.
type BuildList struct {
	Count int     `json:"count"`
	Value []Build `json:"value"`
}
