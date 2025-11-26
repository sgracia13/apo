// Package ui provides the terminal user interface.
package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/user/apo/internal/agent"
	"github.com/user/apo/internal/api"
	"github.com/user/apo/internal/config"
	"github.com/user/apo/internal/domain"
	"github.com/user/apo/internal/ui/components"
	"github.com/user/apo/internal/ui/terminal"
	"github.com/user/apo/internal/ui/views"
	"github.com/user/apo/internal/ui/views/details"
)

// App is the main TUI application.
type App struct {
	term   *terminal.Terminal
	client *api.Client
	agent  *agent.Agent
	config *config.Config

	tabBar    *components.TabBar
	statusBar *components.StatusBar

	dashboard      *views.DashboardView
	boards         *views.BoardsView
	pipelines      *views.PipelinesView
	repos          *views.ReposView
	prs            *views.PullRequestsView
	copilot        *views.CopilotView
	workItemDetail *details.WorkItemDetailView
	prDetail       *details.PRDetailView

	running      bool
	currentView  views.ViewID
	previousView views.ViewID

	mu          sync.RWMutex
	workItems   []domain.WorkItem
	builds      []domain.Build
	pipelineList []domain.Pipeline
	repoList    []domain.Repository
	prList      []domain.PullRequest
	lastRefresh time.Time
	loading     bool
}

// NewApp creates a new TUI application.
func NewApp(cfg *config.Config) (*App, error) {
	if err := cfg.ValidateWithProject(); err != nil {
		return nil, err
	}

	term := terminal.New()
	client := api.NewClient(cfg)
	ag := agent.New(client)

	tabs := []components.Tab{
		{ID: "dashboard", Name: "Dashboard", Key: "1", Icon: "🏠"},
		{ID: "boards", Name: "Boards", Key: "2", Icon: "📋"},
		{ID: "pipelines", Name: "Pipelines", Key: "3", Icon: "🔧"},
		{ID: "repos", Name: "Repos", Key: "4", Icon: "📁"},
		{ID: "prs", Name: "PRs", Key: "5", Icon: "🔀"},
		{ID: "copilot", Name: "Copilot", Key: "/", Icon: "🤖"},
	}

	detailCfg := details.DetailConfig{
		Organization: cfg.Organization,
		Project:      cfg.Project,
	}

	app := &App{
		term:           term,
		client:         client,
		agent:          ag,
		config:         cfg,
		tabBar:         components.NewTabBar(term, tabs),
		statusBar:      components.NewStatusBar(term),
		dashboard:      views.NewDashboardView(term),
		boards:         views.NewBoardsView(term),
		pipelines:      views.NewPipelinesView(term),
		repos:          views.NewReposView(term),
		prs:            views.NewPullRequestsView(term),
		copilot:        views.NewCopilotView(term, ag),
		workItemDetail: details.NewWorkItemDetailView(term, detailCfg),
		prDetail:       details.NewPRDetailView(term, detailCfg),
		currentView:    views.ViewDashboard,
	}

	app.boards.OnSelectItem(func(item *domain.WorkItem) {
		app.showWorkItemDetail(item)
	})

	app.prs.OnSelectItem(func(pr *domain.PullRequest) {
		app.showPRDetail(pr)
	})

	return app, nil
}

// Run starts the TUI.
func (a *App) Run() error {
	a.running = true

	a.term.EnableRawMode()
	defer a.term.DisableRawMode()
	a.term.HideCursor()
	defer a.term.ShowCursor()
	a.term.Clear()

	a.setStatus("Loading...")
	go a.refreshData()

	for a.running {
		a.render()
		key, err := a.term.ReadKey()
		if err != nil {
			continue
		}
		a.handleInput(key)
	}

	a.term.Clear()
	return nil
}

func (a *App) handleInput(key terminal.Key) {
	switch key.Type {
	case terminal.KeyCtrlC:
		a.running = false
		return
	case terminal.KeyEscape:
		if a.isDetailView() {
			a.currentView = a.previousView
			return
		}
		if a.currentView == views.ViewCopilot {
			a.switchToView(views.ViewDashboard)
			return
		}
		return
	}

	if a.getCurrentView().HandleKey(key) {
		return
	}

	switch key.Type {
	case terminal.KeyRune:
		switch key.Rune {
		case 'q', 'Q':
			a.running = false
		case '1':
			a.switchToView(views.ViewDashboard)
		case '2':
			a.switchToView(views.ViewBoards)
		case '3':
			a.switchToView(views.ViewPipelines)
		case '4':
			a.switchToView(views.ViewRepos)
		case '5':
			a.switchToView(views.ViewPullRequests)
		case '/', ':':
			a.switchToView(views.ViewCopilot)
		case 'r', 'R':
			a.setStatus("Refreshing...")
			go a.refreshData()
		case 'b':
			if a.isDetailView() {
				a.currentView = a.previousView
			}
		}
	case terminal.KeyTab:
		a.tabBar.Next()
		a.switchToTabView()
	}
}

func (a *App) getCurrentView() views.View {
	switch a.currentView {
	case views.ViewDashboard:
		return a.dashboard
	case views.ViewBoards:
		return a.boards
	case views.ViewPipelines:
		return a.pipelines
	case views.ViewRepos:
		return a.repos
	case views.ViewPullRequests:
		return a.prs
	case views.ViewCopilot:
		return a.copilot
	case views.ViewWorkItemDetail:
		return a.workItemDetail
	case views.ViewPRDetail:
		return a.prDetail
	default:
		return a.dashboard
	}
}

func (a *App) isDetailView() bool {
	return a.currentView == views.ViewWorkItemDetail || a.currentView == views.ViewPRDetail
}

func (a *App) switchToView(id views.ViewID) {
	if a.currentView == id {
		return
	}
	a.getCurrentView().OnExit()
	a.previousView = a.currentView
	a.currentView = id
	a.getCurrentView().OnEnter()

	switch id {
	case views.ViewDashboard:
		a.tabBar.SetActiveByID("dashboard")
	case views.ViewBoards:
		a.tabBar.SetActiveByID("boards")
	case views.ViewPipelines:
		a.tabBar.SetActiveByID("pipelines")
	case views.ViewRepos:
		a.tabBar.SetActiveByID("repos")
	case views.ViewPullRequests:
		a.tabBar.SetActiveByID("prs")
	case views.ViewCopilot:
		a.tabBar.SetActiveByID("copilot")
	}
}

func (a *App) switchToTabView() {
	tab := a.tabBar.ActiveTab()
	if tab == nil {
		return
	}
	switch tab.ID {
	case "dashboard":
		a.switchToView(views.ViewDashboard)
	case "boards":
		a.switchToView(views.ViewBoards)
	case "pipelines":
		a.switchToView(views.ViewPipelines)
	case "repos":
		a.switchToView(views.ViewRepos)
	case "prs":
		a.switchToView(views.ViewPullRequests)
	case "copilot":
		a.switchToView(views.ViewCopilot)
	}
}

func (a *App) showWorkItemDetail(item *domain.WorkItem) {
	a.workItemDetail.SetWorkItem(item)
	a.previousView = a.currentView
	a.currentView = views.ViewWorkItemDetail
}

func (a *App) showPRDetail(pr *domain.PullRequest) {
	a.prDetail.SetPullRequest(pr)
	a.previousView = a.currentView
	a.currentView = views.ViewPRDetail
}

func (a *App) refreshData() {
	a.loading = true
	a.mu.Lock()

	if items, err := a.client.GetMyWorkItems(); err == nil {
		a.workItems = items
		a.boards.SetWorkItems(items)
	}

	if builds, err := a.client.ListBuilds("", "", 20); err == nil {
		a.builds = builds
	}

	if pipelines, err := a.client.ListPipelines(); err == nil {
		a.pipelineList = pipelines
		a.pipelines.SetPipelines(pipelines)
	}

	if repos, err := a.client.ListRepositories(); err == nil {
		a.repoList = repos
		a.repos.SetRepositories(repos)
	}

	if prs, err := a.client.GetActivePullRequests(20); err == nil {
		a.prList = prs
		a.prs.SetPullRequests(prs)
	}

	a.dashboard.SetData(a.workItems, a.builds, a.prList)

	a.lastRefresh = time.Now()
	a.loading = false
	a.mu.Unlock()

	a.statusBar.SetLastRefresh(a.lastRefresh)
	a.setStatus("Data refreshed")
}

func (a *App) setStatus(msg string) {
	a.statusBar.SetMessage(msg)
}

func (a *App) render() {
	a.term.Clear()
	width := a.term.Width()
	height := a.term.Height()

	a.renderHeader(width)

	if !a.isDetailView() {
		a.tabBar.Render(3, 1, width)
	}

	contentStart := 5
	contentHeight := height - 7

	if a.loading {
		a.term.MoveTo(contentStart+contentHeight/2, width/2-10)
		fmt.Print(terminal.Style("⏳ Loading...", terminal.FgYellow, terminal.Bold))
	} else {
		a.mu.RLock()
		a.getCurrentView().Render(contentStart, width, contentHeight)
		a.mu.RUnlock()
	}

	a.updateHelpText()
	a.statusBar.Render(height-2, width)
}

func (a *App) renderHeader(width int) {
	a.term.MoveTo(1, 1)
	title := "╔═══ Azure Prod Ops ═══╗"
	padding := (width - len(title)) / 2
	fmt.Print(strings.Repeat(" ", padding))
	fmt.Print(terminal.Style(title, terminal.Bold, terminal.FgCyan))

	orgInfo := fmt.Sprintf("  %s/%s", a.config.Organization, a.config.Project)
	a.term.MoveTo(1, width-len(orgInfo)-1)
	fmt.Print(terminal.Style(orgInfo, terminal.Dim))
}

func (a *App) updateHelpText() {
	var help string
	switch {
	case a.currentView == views.ViewCopilot:
		help = " [Enter] Send │ [Esc] Back │ [Ctrl+C] Quit "
	case a.isDetailView():
		help = " [Esc/b] Back │ [q] Quit "
	case a.isFilterMode():
		help = " [Enter] Apply │ [Esc] Cancel │ Type to filter... "
	default:
		help = " [1-5] Tab │ [/] Copilot │ [↑↓/jk] Navigate │ [Enter] Details │ [f] Filter │ [r] Refresh │ [q] Quit "
	}
	a.statusBar.SetHelp(help)
}

func (a *App) isFilterMode() bool {
	switch v := a.getCurrentView().(type) {
	case *views.BoardsView:
		return v.IsFilterMode()
	case *views.PipelinesView:
		return v.IsFilterMode()
	case *views.ReposView:
		return v.IsFilterMode()
	case *views.PullRequestsView:
		return v.IsFilterMode()
	}
	return false
}
