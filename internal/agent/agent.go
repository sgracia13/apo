// Package agent provides natural language query interpretation.
package agent

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/user/apo/internal/api"
	"github.com/user/apo/internal/domain"
)

// Intent represents the type of user query.
type Intent int

const (
	IntentUnknown Intent = iota
	IntentHelp
	IntentMyWorkItems
	IntentFailedBuilds
	IntentRunningBuilds
	IntentRecentBuilds
	IntentListPipelines
	IntentListRepos
	IntentActivePRs
	IntentListProjects
)

// Result represents the result of a query execution.
type Result struct {
	Success     bool
	Message     string
	Data        interface{}
	Suggestions []string
}

type pattern struct {
	intent  Intent
	regexps []*regexp.Regexp
}

// Agent interprets natural language queries.
type Agent struct {
	client   *api.Client
	patterns []pattern
}

// New creates a new Agent.
func New(client *api.Client) *Agent {
	a := &Agent{client: client}
	a.initPatterns()
	return a
}

func (a *Agent) initPatterns() {
	a.patterns = []pattern{
		{IntentMyWorkItems, compile(
			`(?i)(my|assigned to me).*work\s*items?`,
			`(?i)what('s| is| are).*assigned to me`,
			`(?i)my (tasks?|bugs?|stories?)`,
		)},
		{IntentFailedBuilds, compile(
			`(?i)failed builds?`,
			`(?i)what('s| is| are) (failing|broken)`,
			`(?i)build failures?`,
		)},
		{IntentRunningBuilds, compile(
			`(?i)running builds?`,
			`(?i)what('s| is) (running|building)`,
			`(?i)active builds?`,
		)},
		{IntentRecentBuilds, compile(
			`(?i)recent builds?`,
			`(?i)build (history|status)`,
			`(?i)show.*builds?`,
		)},
		{IntentListPipelines, compile(
			`(?i)(list|show|get).*pipelines?`,
			`(?i)what pipelines?`,
		)},
		{IntentListRepos, compile(
			`(?i)(list|show|get).*repo`,
			`(?i)what repo`,
		)},
		{IntentActivePRs, compile(
			`(?i)(active|open) (pull requests?|prs?)`,
			`(?i)(list|show|get).*(pull requests?|prs?)`,
			`(?i)pending (reviews?|prs?)`,
		)},
		{IntentListProjects, compile(
			`(?i)(list|show|get).*projects?`,
		)},
		{IntentHelp, compile(
			`(?i)^help$`,
			`(?i)what can you do`,
		)},
	}
}

func compile(exprs ...string) []*regexp.Regexp {
	result := make([]*regexp.Regexp, len(exprs))
	for i, expr := range exprs {
		result[i] = regexp.MustCompile(expr)
	}
	return result
}

// Ask interprets a query and returns a result.
func (a *Agent) Ask(query string) *Result {
	query = strings.TrimSpace(query)
	if query == "" {
		return &Result{Success: true, Message: "Please ask me something!", Suggestions: []string{"Try: 'help'"}}
	}

	intent := a.matchIntent(query)
	return a.execute(intent)
}

func (a *Agent) matchIntent(query string) Intent {
	for _, p := range a.patterns {
		for _, re := range p.regexps {
			if re.MatchString(query) {
				return p.intent
			}
		}
	}
	return IntentUnknown
}

func (a *Agent) execute(intent Intent) *Result {
	switch intent {
	case IntentHelp:
		return &Result{
			Success: true,
			Message: "I can help you with Azure DevOps! Try asking:",
			Suggestions: []string{
				"What work items are assigned to me?",
				"Show me failed builds",
				"List all pipelines",
				"What PRs are open?",
			},
		}
	case IntentMyWorkItems:
		items, err := a.client.GetMyWorkItems()
		if err != nil {
			return &Result{Success: false, Message: fmt.Sprintf("Error: %v", err)}
		}
		if len(items) == 0 {
			return &Result{Success: true, Message: "No work items assigned to you.", Data: items}
		}
		return &Result{Success: true, Message: fmt.Sprintf("Found %d work item(s):", len(items)), Data: items}
	case IntentFailedBuilds:
		builds, err := a.client.GetFailedBuilds(15)
		if err != nil {
			return &Result{Success: false, Message: fmt.Sprintf("Error: %v", err)}
		}
		if len(builds) == 0 {
			return &Result{Success: true, Message: "No failed builds! 🎉", Data: builds}
		}
		return &Result{Success: true, Message: fmt.Sprintf("Found %d failed build(s):", len(builds)), Data: builds}
	case IntentRunningBuilds:
		builds, err := a.client.GetRunningBuilds()
		if err != nil {
			return &Result{Success: false, Message: fmt.Sprintf("Error: %v", err)}
		}
		if len(builds) == 0 {
			return &Result{Success: true, Message: "No builds currently running.", Data: builds}
		}
		return &Result{Success: true, Message: fmt.Sprintf("Found %d running build(s):", len(builds)), Data: builds}
	case IntentRecentBuilds:
		builds, err := a.client.ListBuilds("", "", 15)
		if err != nil {
			return &Result{Success: false, Message: fmt.Sprintf("Error: %v", err)}
		}
		return &Result{Success: true, Message: fmt.Sprintf("Found %d recent build(s):", len(builds)), Data: builds}
	case IntentListPipelines:
		pipelines, err := a.client.ListPipelines()
		if err != nil {
			return &Result{Success: false, Message: fmt.Sprintf("Error: %v", err)}
		}
		return &Result{Success: true, Message: fmt.Sprintf("Found %d pipeline(s):", len(pipelines)), Data: pipelines}
	case IntentListRepos:
		repos, err := a.client.ListRepositories()
		if err != nil {
			return &Result{Success: false, Message: fmt.Sprintf("Error: %v", err)}
		}
		return &Result{Success: true, Message: fmt.Sprintf("Found %d repository(ies):", len(repos)), Data: repos}
	case IntentActivePRs:
		prs, err := a.client.GetActivePullRequests(20)
		if err != nil {
			return &Result{Success: false, Message: fmt.Sprintf("Error: %v", err)}
		}
		if len(prs) == 0 {
			return &Result{Success: true, Message: "No active pull requests.", Data: prs}
		}
		return &Result{Success: true, Message: fmt.Sprintf("Found %d active PR(s):", len(prs)), Data: prs}
	case IntentListProjects:
		projects, err := a.client.ListProjects()
		if err != nil {
			return &Result{Success: false, Message: fmt.Sprintf("Error: %v", err)}
		}
		return &Result{Success: true, Message: fmt.Sprintf("Found %d project(s):", len(projects)), Data: projects}
	default:
		return &Result{
			Success: true,
			Message: "I'm not sure what you're asking.",
			Suggestions: []string{"Try: 'help' to see what I can do"},
		}
	}
}

// FormatResult formats result data for display.
func FormatResult(data interface{}) string {
	var sb strings.Builder
	switch d := data.(type) {
	case []domain.WorkItem:
		for _, item := range d {
			icon := GetWorkItemIcon(item.Type())
			sb.WriteString(fmt.Sprintf("  %s #%d %s [%s]\n", icon, item.ID, truncate(item.Title(), 50), item.State()))
		}
	case []domain.Build:
		for _, b := range d {
			icon := GetBuildIcon(b.Result)
			sb.WriteString(fmt.Sprintf("  %s #%s %s\n", icon, b.BuildNumber, b.Definition.Name))
		}
	case []domain.Pipeline:
		for _, p := range d {
			sb.WriteString(fmt.Sprintf("  🔧 [%d] %s\n", p.ID, p.FullPath()))
		}
	case []domain.Repository:
		for _, r := range d {
			sb.WriteString(fmt.Sprintf("  📁 %s (%s)\n", r.Name, r.DefaultBranchName()))
		}
	case []domain.PullRequest:
		for _, pr := range d {
			icon := "🔀"
			if pr.IsDraft {
				icon = "📝"
			}
			sb.WriteString(fmt.Sprintf("  %s #%d %s\n", icon, pr.PullRequestID, truncate(pr.Title, 50)))
		}
	case []domain.Project:
		for _, p := range d {
			sb.WriteString(fmt.Sprintf("  📦 %s\n", p.Name))
		}
	}
	return sb.String()
}

// GetWorkItemIcon returns an emoji for a work item type.
func GetWorkItemIcon(itemType string) string {
	switch strings.ToLower(itemType) {
	case "bug":
		return "🐛"
	case "user story", "story":
		return "📖"
	case "task":
		return "✅"
	case "epic":
		return "🏔️"
	case "feature":
		return "⭐"
	default:
		return "📋"
	}
}

// GetBuildIcon returns an emoji for a build result.
func GetBuildIcon(result string) string {
	switch strings.ToLower(result) {
	case "succeeded":
		return "✅"
	case "failed":
		return "❌"
	case "canceled":
		return "⏹️"
	default:
		return "🔄"
	}
}

// GetStateIcon returns an emoji for a work item state.
func GetStateIcon(state string) string {
	switch strings.ToLower(state) {
	case "new":
		return "🆕"
	case "active", "in progress":
		return "🔄"
	case "resolved":
		return "✔️"
	case "closed", "done":
		return "✅"
	default:
		return "•"
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
