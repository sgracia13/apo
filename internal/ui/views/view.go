// Package views provides the application views.
package views

import "github.com/user/apo/internal/ui/terminal"

// ViewID identifies a view.
type ViewID string

const (
	ViewDashboard      ViewID = "dashboard"
	ViewBoards         ViewID = "boards"
	ViewPipelines      ViewID = "pipelines"
	ViewRepos          ViewID = "repos"
	ViewPullRequests   ViewID = "pullrequests"
	ViewCopilot        ViewID = "copilot"
	ViewWorkItemDetail ViewID = "workitem_detail"
	ViewPRDetail       ViewID = "pr_detail"
)

// View defines the interface for views.
type View interface {
	ID() ViewID
	Title() string
	Render(startRow, width, height int)
	HandleKey(key terminal.Key) bool
	OnEnter()
	OnExit()
}

// BaseView provides common functionality.
type BaseView struct {
	term  *terminal.Terminal
	id    ViewID
	title string
}

// NewBaseView creates a new base view.
func NewBaseView(term *terminal.Terminal, id ViewID, title string) BaseView {
	return BaseView{term: term, id: id, title: title}
}

// ID returns the view ID.
func (v *BaseView) ID() ViewID { return v.id }

// Title returns the view title.
func (v *BaseView) Title() string { return v.title }

// OnEnter is called when entering the view.
func (v *BaseView) OnEnter() {}

// OnExit is called when leaving the view.
func (v *BaseView) OnExit() {}

// Term returns the terminal.
func (v *BaseView) Term() *terminal.Terminal { return v.term }
