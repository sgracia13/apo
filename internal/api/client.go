// Package api provides the Azure DevOps REST API client.
package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/user/apo/internal/config"
	"github.com/user/apo/internal/domain"
)

// Client is the Azure DevOps API client.
type Client struct {
	baseURL    string
	org        string
	project    string
	pat        string
	apiVersion string
	http       *http.Client
}

// NewClient creates a new API client.
func NewClient(cfg *config.Config) *Client {
	return &Client{
		baseURL:    cfg.APIURL,
		org:        cfg.Organization,
		project:    cfg.Project,
		pat:        cfg.PAT,
		apiVersion: cfg.APIVersion,
		http:       &http.Client{Timeout: config.DefaultTimeout},
	}
}

func (c *Client) do(method, url string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	auth := base64.StdEncoding.EncodeToString([]byte(":" + c.pat))
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}
	}
	return nil
}

func (c *Client) url(path string, params ...string) string {
	u := fmt.Sprintf("%s/%s/%s/%s?api-version=%s", c.baseURL, c.org, c.project, path, c.apiVersion)
	for i := 0; i < len(params)-1; i += 2 {
		u += fmt.Sprintf("&%s=%s", params[i], url.QueryEscape(params[i+1]))
	}
	return u
}

func (c *Client) orgURL(path string, params ...string) string {
	u := fmt.Sprintf("%s/%s/%s?api-version=%s", c.baseURL, c.org, path, c.apiVersion)
	for i := 0; i < len(params)-1; i += 2 {
		u += fmt.Sprintf("&%s=%s", params[i], url.QueryEscape(params[i+1]))
	}
	return u
}

// GetMyWorkItems returns work items assigned to the current user.
func (c *Client) GetMyWorkItems() ([]domain.WorkItem, error) {
	wiql := `SELECT [System.Id] FROM WorkItems 
             WHERE [System.AssignedTo] = @Me 
             AND [System.State] <> 'Closed' 
             AND [System.State] <> 'Removed'
             ORDER BY [System.ChangedDate] DESC`

	var result domain.WorkItemList
	if err := c.do("POST", c.url("_apis/wit/wiql"), map[string]string{"query": wiql}, &result); err != nil {
		return nil, err
	}

	if len(result.WorkItems) == 0 {
		return []domain.WorkItem{}, nil
	}

	ids := make([]string, len(result.WorkItems))
	for i, ref := range result.WorkItems {
		ids[i] = fmt.Sprintf("%d", ref.ID)
	}

	var batch domain.WorkItemBatch
	batchURL := c.url("_apis/wit/workitems", "ids", strings.Join(ids, ","))
	if err := c.do("GET", batchURL, nil, &batch); err != nil {
		return nil, err
	}
	return batch.Value, nil
}

// ListBuilds returns recent builds.
func (c *Client) ListBuilds(status, result string, top int) ([]domain.Build, error) {
	params := []string{}
	if status != "" {
		params = append(params, "statusFilter", status)
	}
	if result != "" {
		params = append(params, "resultFilter", result)
	}
	if top > 0 {
		params = append(params, "$top", fmt.Sprintf("%d", top))
	}

	var resp domain.BuildList
	if err := c.do("GET", c.url("_apis/build/builds", params...), nil, &resp); err != nil {
		return nil, err
	}
	return resp.Value, nil
}

// GetFailedBuilds returns recently failed builds.
func (c *Client) GetFailedBuilds(top int) ([]domain.Build, error) {
	return c.ListBuilds("completed", "failed", top)
}

// GetRunningBuilds returns currently running builds.
func (c *Client) GetRunningBuilds() ([]domain.Build, error) {
	return c.ListBuilds("inProgress", "", 50)
}

// ListPipelines returns all pipelines.
func (c *Client) ListPipelines() ([]domain.Pipeline, error) {
	var resp domain.PipelineList
	if err := c.do("GET", c.url("_apis/pipelines"), nil, &resp); err != nil {
		return nil, err
	}
	return resp.Value, nil
}

// ListRepositories returns all Git repositories.
func (c *Client) ListRepositories() ([]domain.Repository, error) {
	var resp domain.RepositoryList
	if err := c.do("GET", c.url("_apis/git/repositories"), nil, &resp); err != nil {
		return nil, err
	}
	return resp.Value, nil
}

// GetActivePullRequests returns active pull requests.
func (c *Client) GetActivePullRequests(top int) ([]domain.PullRequest, error) {
	params := []string{"searchCriteria.status", "active"}
	if top > 0 {
		params = append(params, "$top", fmt.Sprintf("%d", top))
	}

	var resp domain.PullRequestList
	if err := c.do("GET", c.url("_apis/git/pullrequests", params...), nil, &resp); err != nil {
		return nil, err
	}
	return resp.Value, nil
}

// ListProjects returns all projects in the organization.
func (c *Client) ListProjects() ([]domain.Project, error) {
	var resp domain.ProjectList
	if err := c.do("GET", c.orgURL("_apis/projects"), nil, &resp); err != nil {
		return nil, err
	}
	return resp.Value, nil
}
