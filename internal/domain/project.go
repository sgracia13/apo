package domain

// Project represents an Azure DevOps project.
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	State       string `json:"state"`
	Visibility  string `json:"visibility"`
}

// ProjectList is the response from listing projects.
type ProjectList struct {
	Count int       `json:"count"`
	Value []Project `json:"value"`
}
