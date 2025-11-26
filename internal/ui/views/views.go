package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/user/apo/internal/agent"
	"github.com/user/apo/internal/domain"
	"github.com/user/apo/internal/ui/components"
	"github.com/user/apo/internal/ui/terminal"
)

// DashboardView provides an overview.
type DashboardView struct {
	BaseView
	workItems []domain.WorkItem
	builds    []domain.Build
	prs       []domain.PullRequest
}

// NewDashboardView creates a dashboard view.
func NewDashboardView(term *terminal.Terminal) *DashboardView {
	return &DashboardView{BaseView: NewBaseView(term, ViewDashboard, "Dashboard")}
}

// SetData sets dashboard data.
func (v *DashboardView) SetData(items []domain.WorkItem, builds []domain.Build, prs []domain.PullRequest) {
	v.workItems = items
	v.builds = builds
	v.prs = prs
}

// Render renders the dashboard.
func (v *DashboardView) Render(startRow, width, height int) {
	colWidth := (width - 4) / 2
	halfHeight := (height - 2) / 2

	// Work Items
	v.term.MoveTo(startRow, 2)
	fmt.Print(terminal.Style("📋 My Work Items", terminal.Bold, terminal.FgYellow))
	row := startRow + 1
	for i, item := range v.workItems {
		if i >= halfHeight-1 {
			break
		}
		v.term.MoveTo(row, 2)
		icon := agent.GetWorkItemIcon(item.Type())
		fmt.Printf("%s #%d %s", icon, item.ID, terminal.Truncate(item.Title(), colWidth-15))
		row++
	}
	if len(v.workItems) == 0 {
		v.term.MoveTo(row, 4)
		fmt.Print(terminal.Style("No work items", terminal.Dim))
	}

	// Builds
	v.term.MoveTo(startRow, colWidth+3)
	fmt.Print(terminal.Style("🔧 Recent Builds", terminal.Bold, terminal.FgYellow))
	row = startRow + 1
	for i, b := range v.builds {
		if i >= halfHeight-1 {
			break
		}
		v.term.MoveTo(row, colWidth+3)
		icon := agent.GetBuildIcon(b.Result)
		fmt.Printf("%s #%s %s", icon, b.BuildNumber, terminal.Truncate(b.Definition.Name, colWidth-15))
		row++
	}
	if len(v.builds) == 0 {
		v.term.MoveTo(row, colWidth+5)
		fmt.Print(terminal.Style("No builds", terminal.Dim))
	}

	// PRs
	v.term.MoveTo(startRow+halfHeight+1, 2)
	fmt.Print(terminal.Style("🔀 Active Pull Requests", terminal.Bold, terminal.FgYellow))
	row = startRow + halfHeight + 2
	for i, pr := range v.prs {
		if i >= halfHeight-2 {
			break
		}
		v.term.MoveTo(row, 2)
		icon := "🔀"
		if pr.IsDraft {
			icon = "📝"
		}
		fmt.Printf("%s #%d %s", icon, pr.PullRequestID, terminal.Truncate(pr.Title, width-20))
		row++
	}
	if len(v.prs) == 0 {
		v.term.MoveTo(row, 4)
		fmt.Print(terminal.Style("No active PRs", terminal.Dim))
	}
}

// HandleKey handles input.
func (v *DashboardView) HandleKey(key terminal.Key) bool { return false }

// BoardsView displays work items.
type BoardsView struct {
	BaseView
	list      *components.List
	workItems []domain.WorkItem
	onSelect  func(*domain.WorkItem)
}

// NewBoardsView creates a boards view.
func NewBoardsView(term *terminal.Terminal) *BoardsView {
	v := &BoardsView{BaseView: NewBaseView(term, ViewBoards, "Boards")}
	v.list = components.NewList(term, "📋 Work Items")
	return v
}

// SetWorkItems sets work items.
func (v *BoardsView) SetWorkItems(items []domain.WorkItem) {
	v.workItems = items
	listItems := make([]components.ListItem, len(items))
	for i, item := range items {
		listItems[i] = components.ListItem{
			ID:    fmt.Sprintf("%d", item.ID),
			Icon:  agent.GetWorkItemIcon(item.Type()),
			Label: fmt.Sprintf("#%d %s [%s]", item.ID, item.Title(), item.State()),
			Data:  item,
		}
	}
	v.list.SetItems(listItems)
}

// OnSelectItem sets the select callback.
func (v *BoardsView) OnSelectItem(fn func(*domain.WorkItem)) { v.onSelect = fn }

// SelectedWorkItem returns selected item.
func (v *BoardsView) SelectedWorkItem() *domain.WorkItem {
	idx := v.list.SelectedIndex()
	if idx >= 0 && idx < len(v.workItems) {
		return &v.workItems[idx]
	}
	return nil
}

// Render renders the view.
func (v *BoardsView) Render(startRow, width, height int) {
	v.list.Render(startRow, 2, width, height)
}

// HandleKey handles input.
func (v *BoardsView) HandleKey(key terminal.Key) bool {
	if v.list.IsFilterMode() {
		switch key.Type {
		case terminal.KeyEnter, terminal.KeyEscape:
			v.list.ToggleFilterMode()
			return true
		case terminal.KeyBackspace:
			q := v.list.FilterQuery()
			if len(q) > 0 {
				v.list.SetFilter(q[:len(q)-1])
			}
			return true
		case terminal.KeyRune:
			v.list.SetFilter(v.list.FilterQuery() + string(key.Rune))
			return true
		}
		return false
	}

	switch key.Type {
	case terminal.KeyUp:
		v.list.MoveUp()
		return true
	case terminal.KeyDown:
		v.list.MoveDown()
		return true
	case terminal.KeyEnter:
		if v.onSelect != nil {
			if wi := v.SelectedWorkItem(); wi != nil {
				v.onSelect(wi)
			}
		}
		return true
	case terminal.KeyRune:
		switch key.Rune {
		case 'j':
			v.list.MoveDown()
			return true
		case 'k':
			v.list.MoveUp()
			return true
		case 'g':
			v.list.MoveToTop()
			return true
		case 'G':
			v.list.MoveToBottom()
			return true
		case 'f', '/':
			v.list.ToggleFilterMode()
			return true
		}
	}
	return false
}

// IsFilterMode returns filter state.
func (v *BoardsView) IsFilterMode() bool { return v.list.IsFilterMode() }

// PipelinesView displays pipelines.
type PipelinesView struct {
	BaseView
	list      *components.List
	pipelines []domain.Pipeline
}

// NewPipelinesView creates a pipelines view.
func NewPipelinesView(term *terminal.Terminal) *PipelinesView {
	v := &PipelinesView{BaseView: NewBaseView(term, ViewPipelines, "Pipelines")}
	v.list = components.NewList(term, "🔧 Pipelines")
	return v
}

// SetPipelines sets pipelines.
func (v *PipelinesView) SetPipelines(items []domain.Pipeline) {
	v.pipelines = items
	listItems := make([]components.ListItem, len(items))
	for i, p := range items {
		listItems[i] = components.ListItem{
			ID:    fmt.Sprintf("%d", p.ID),
			Icon:  "🔧",
			Label: fmt.Sprintf("[%d] %s", p.ID, p.FullPath()),
		}
	}
	v.list.SetItems(listItems)
}

// Render renders the view.
func (v *PipelinesView) Render(startRow, width, height int) {
	v.list.Render(startRow, 2, width, height)
}

// HandleKey handles input.
func (v *PipelinesView) HandleKey(key terminal.Key) bool {
	if v.list.IsFilterMode() {
		switch key.Type {
		case terminal.KeyEnter, terminal.KeyEscape:
			v.list.ToggleFilterMode()
			return true
		case terminal.KeyBackspace:
			q := v.list.FilterQuery()
			if len(q) > 0 {
				v.list.SetFilter(q[:len(q)-1])
			}
			return true
		case terminal.KeyRune:
			v.list.SetFilter(v.list.FilterQuery() + string(key.Rune))
			return true
		}
		return false
	}

	switch key.Type {
	case terminal.KeyUp:
		v.list.MoveUp()
		return true
	case terminal.KeyDown:
		v.list.MoveDown()
		return true
	case terminal.KeyRune:
		switch key.Rune {
		case 'j':
			v.list.MoveDown()
			return true
		case 'k':
			v.list.MoveUp()
			return true
		case 'g':
			v.list.MoveToTop()
			return true
		case 'G':
			v.list.MoveToBottom()
			return true
		case 'f', '/':
			v.list.ToggleFilterMode()
			return true
		}
	}
	return false
}

// IsFilterMode returns filter state.
func (v *PipelinesView) IsFilterMode() bool { return v.list.IsFilterMode() }

// ReposView displays repositories.
type ReposView struct {
	BaseView
	list  *components.List
	repos []domain.Repository
}

// NewReposView creates a repos view.
func NewReposView(term *terminal.Terminal) *ReposView {
	v := &ReposView{BaseView: NewBaseView(term, ViewRepos, "Repos")}
	v.list = components.NewList(term, "📁 Repositories")
	return v
}

// SetRepositories sets repos.
func (v *ReposView) SetRepositories(items []domain.Repository) {
	v.repos = items
	listItems := make([]components.ListItem, len(items))
	for i, r := range items {
		listItems[i] = components.ListItem{
			ID:    r.ID,
			Icon:  "📁",
			Label: fmt.Sprintf("%s (%s)", r.Name, r.DefaultBranchName()),
		}
	}
	v.list.SetItems(listItems)
}

// Render renders the view.
func (v *ReposView) Render(startRow, width, height int) {
	v.list.Render(startRow, 2, width, height)
}

// HandleKey handles input.
func (v *ReposView) HandleKey(key terminal.Key) bool {
	if v.list.IsFilterMode() {
		switch key.Type {
		case terminal.KeyEnter, terminal.KeyEscape:
			v.list.ToggleFilterMode()
			return true
		case terminal.KeyBackspace:
			q := v.list.FilterQuery()
			if len(q) > 0 {
				v.list.SetFilter(q[:len(q)-1])
			}
			return true
		case terminal.KeyRune:
			v.list.SetFilter(v.list.FilterQuery() + string(key.Rune))
			return true
		}
		return false
	}

	switch key.Type {
	case terminal.KeyUp:
		v.list.MoveUp()
		return true
	case terminal.KeyDown:
		v.list.MoveDown()
		return true
	case terminal.KeyRune:
		switch key.Rune {
		case 'j':
			v.list.MoveDown()
			return true
		case 'k':
			v.list.MoveUp()
			return true
		case 'g':
			v.list.MoveToTop()
			return true
		case 'G':
			v.list.MoveToBottom()
			return true
		case 'f', '/':
			v.list.ToggleFilterMode()
			return true
		}
	}
	return false
}

// IsFilterMode returns filter state.
func (v *ReposView) IsFilterMode() bool { return v.list.IsFilterMode() }

// PullRequestsView displays PRs.
type PullRequestsView struct {
	BaseView
	list     *components.List
	prs      []domain.PullRequest
	onSelect func(*domain.PullRequest)
}

// NewPullRequestsView creates a PRs view.
func NewPullRequestsView(term *terminal.Terminal) *PullRequestsView {
	v := &PullRequestsView{BaseView: NewBaseView(term, ViewPullRequests, "PRs")}
	v.list = components.NewList(term, "🔀 Pull Requests")
	return v
}

// SetPullRequests sets PRs.
func (v *PullRequestsView) SetPullRequests(items []domain.PullRequest) {
	v.prs = items
	listItems := make([]components.ListItem, len(items))
	for i, pr := range items {
		icon := "🔀"
		if pr.IsDraft {
			icon = "📝"
		}
		listItems[i] = components.ListItem{
			ID:    fmt.Sprintf("%d", pr.PullRequestID),
			Icon:  icon,
			Label: fmt.Sprintf("#%d %s (%s→%s)", pr.PullRequestID, pr.Title, pr.SourceBranch(), pr.TargetBranch()),
		}
	}
	v.list.SetItems(listItems)
}

// OnSelectItem sets select callback.
func (v *PullRequestsView) OnSelectItem(fn func(*domain.PullRequest)) { v.onSelect = fn }

// SelectedPR returns selected PR.
func (v *PullRequestsView) SelectedPR() *domain.PullRequest {
	idx := v.list.SelectedIndex()
	if idx >= 0 && idx < len(v.prs) {
		return &v.prs[idx]
	}
	return nil
}

// Render renders the view.
func (v *PullRequestsView) Render(startRow, width, height int) {
	v.list.Render(startRow, 2, width, height)
}

// HandleKey handles input.
func (v *PullRequestsView) HandleKey(key terminal.Key) bool {
	if v.list.IsFilterMode() {
		switch key.Type {
		case terminal.KeyEnter, terminal.KeyEscape:
			v.list.ToggleFilterMode()
			return true
		case terminal.KeyBackspace:
			q := v.list.FilterQuery()
			if len(q) > 0 {
				v.list.SetFilter(q[:len(q)-1])
			}
			return true
		case terminal.KeyRune:
			v.list.SetFilter(v.list.FilterQuery() + string(key.Rune))
			return true
		}
		return false
	}

	switch key.Type {
	case terminal.KeyUp:
		v.list.MoveUp()
		return true
	case terminal.KeyDown:
		v.list.MoveDown()
		return true
	case terminal.KeyEnter:
		if v.onSelect != nil {
			if pr := v.SelectedPR(); pr != nil {
				v.onSelect(pr)
			}
		}
		return true
	case terminal.KeyRune:
		switch key.Rune {
		case 'j':
			v.list.MoveDown()
			return true
		case 'k':
			v.list.MoveUp()
			return true
		case 'g':
			v.list.MoveToTop()
			return true
		case 'G':
			v.list.MoveToBottom()
			return true
		case 'f', '/':
			v.list.ToggleFilterMode()
			return true
		}
	}
	return false
}

// IsFilterMode returns filter state.
func (v *PullRequestsView) IsFilterMode() bool { return v.list.IsFilterMode() }

// CopilotMessage represents a chat message.
type CopilotMessage struct {
	IsUser  bool
	Content string
	Time    time.Time
}

// CopilotView provides NL interaction.
type CopilotView struct {
	BaseView
	input   *components.Input
	agent   *agent.Agent
	history []CopilotMessage
}

// NewCopilotView creates a copilot view.
func NewCopilotView(term *terminal.Terminal, ag *agent.Agent) *CopilotView {
	v := &CopilotView{
		BaseView: NewBaseView(term, ViewCopilot, "Copilot"),
		agent:    ag,
	}
	v.input = components.NewInput(term, "apo> ")
	return v
}

// OnEnter activates input.
func (v *CopilotView) OnEnter() {
	v.input.Activate()
	v.term.ShowCursor()
}

// OnExit deactivates input.
func (v *CopilotView) OnExit() {
	v.input.Deactivate()
	v.term.HideCursor()
}

// Render renders the view.
func (v *CopilotView) Render(startRow, width, height int) {
	v.term.MoveTo(startRow, 2)
	fmt.Print(terminal.Style("🤖 Copilot - Ask me about Azure DevOps", terminal.Bold, terminal.FgCyan))

	v.term.MoveTo(startRow+1, 2)
	fmt.Print(terminal.Style(strings.Repeat("─", width-4), terminal.Dim))

	if len(v.history) == 0 {
		v.term.MoveTo(startRow+3, 2)
		fmt.Print(terminal.Style("Try: \"What work items are assigned to me?\"", terminal.Dim))
	} else {
		maxLines := height - 6
		start := 0
		if len(v.history) > maxLines {
			start = len(v.history) - maxLines
		}
		row := startRow + 2
		for i := start; i < len(v.history); i++ {
			msg := v.history[i]
			v.term.MoveTo(row, 2)
			if msg.IsUser {
				fmt.Print(terminal.Style("> "+terminal.Truncate(msg.Content, width-6), terminal.FgCyan, terminal.Bold))
			} else if strings.HasPrefix(msg.Content, "💡") {
				fmt.Print(terminal.Style("  "+terminal.Truncate(msg.Content, width-6), terminal.Dim))
			} else {
				fmt.Print(terminal.Style(terminal.Truncate(msg.Content, width-4), terminal.FgGreen))
			}
			row++
		}
	}

	v.input.Render(startRow+height-2, 2, width-4)
}

// HandleKey handles input.
func (v *CopilotView) HandleKey(key terminal.Key) bool {
	switch key.Type {
	case terminal.KeyEnter:
		v.submit()
		return true
	case terminal.KeyBackspace:
		v.input.Backspace()
		return true
	case terminal.KeyRune:
		v.input.InsertChar(key.Rune)
		return true
	}
	return false
}

func (v *CopilotView) submit() {
	query := strings.TrimSpace(v.input.Value())
	if query == "" {
		return
	}

	v.history = append(v.history, CopilotMessage{IsUser: true, Content: query, Time: time.Now()})
	v.input.Clear()

	result := v.agent.Ask(query)
	v.history = append(v.history, CopilotMessage{Content: result.Message, Time: time.Now()})

	for _, s := range result.Suggestions {
		v.history = append(v.history, CopilotMessage{Content: "💡 " + s, Time: time.Now()})
	}

	if result.Data != nil {
		formatted := agent.FormatResult(result.Data)
		for _, line := range strings.Split(formatted, "\n") {
			if line != "" {
				v.history = append(v.history, CopilotMessage{Content: line, Time: time.Now()})
			}
		}
	}
}
