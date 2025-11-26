// Package details provides detail view implementations.
package details

import (
	"fmt"
	"strings"
	"time"

	"github.com/user/apo/internal/agent"
	"github.com/user/apo/internal/domain"
	"github.com/user/apo/internal/ui/terminal"
	"github.com/user/apo/internal/ui/views"
)

// DetailConfig holds configuration for detail views.
type DetailConfig struct {
	Organization string
	Project      string
}

// WorkItemDetailView shows work item details.
type WorkItemDetailView struct {
	views.BaseView
	workItem *domain.WorkItem
	config   DetailConfig
}

// NewWorkItemDetailView creates a work item detail view.
func NewWorkItemDetailView(term *terminal.Terminal, cfg DetailConfig) *WorkItemDetailView {
	return &WorkItemDetailView{
		BaseView: views.NewBaseView(term, views.ViewWorkItemDetail, "Work Item"),
		config:   cfg,
	}
}

// SetWorkItem sets the work item.
func (v *WorkItemDetailView) SetWorkItem(item *domain.WorkItem) {
	v.workItem = item
}

// Render renders the detail view.
func (v *WorkItemDetailView) Render(startRow, width, height int) {
	if v.workItem == nil {
		return
	}
	item := v.workItem
	term := v.Term()

	term.MoveTo(startRow, 2)
	icon := agent.GetWorkItemIcon(item.Type())
	fmt.Print(terminal.Style(fmt.Sprintf("%s %s #%d", icon, item.Type(), item.ID), terminal.Bold, terminal.FgCyan))

	term.MoveTo(startRow+2, 2)
	fmt.Print(terminal.Style(terminal.Truncate(item.Title(), width-4), terminal.Bold))

	term.MoveTo(startRow+4, 2)
	stateIcon := agent.GetStateIcon(item.State())
	fmt.Print(terminal.Style("State: ", terminal.Dim))
	fmt.Print(terminal.Style(stateIcon+" "+item.State(), terminal.Bold, terminal.FgYellow))

	row := startRow + 6
	term.MoveTo(row, 2)
	fmt.Print(terminal.Style("Assigned To: ", terminal.Dim))
	fmt.Print(terminal.Truncate(item.AssignedTo(), 30))

	term.MoveTo(row, width/2)
	fmt.Print(terminal.Style("Created: ", terminal.Dim))
	fmt.Print(formatDate(item.GetField("System.CreatedDate")))

	descRow := startRow + 9
	term.MoveTo(descRow, 2)
	fmt.Print(terminal.Style("─── Description ", terminal.Dim))
	fmt.Print(terminal.Style(strings.Repeat("─", width-20), terminal.Dim))

	desc := item.GetField("System.Description")
	if desc != "" {
		desc = stripHTML(desc)
		lines := wrapText(desc, width-6)
		for i, line := range lines {
			if descRow+1+i >= startRow+height-2 {
				break
			}
			term.MoveTo(descRow+1+i, 4)
			fmt.Print(line)
		}
	} else {
		term.MoveTo(descRow+1, 4)
		fmt.Print(terminal.Style("No description.", terminal.Dim))
	}

	term.MoveTo(startRow+height-2, 2)
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_workitems/edit/%d",
		v.config.Organization, v.config.Project, item.ID)
	fmt.Print(terminal.Style("URL: "+terminal.Truncate(url, width-10), terminal.Dim))
}

// HandleKey handles input.
func (v *WorkItemDetailView) HandleKey(key terminal.Key) bool { return false }

// PRDetailView shows PR details.
type PRDetailView struct {
	views.BaseView
	pr     *domain.PullRequest
	config DetailConfig
}

// NewPRDetailView creates a PR detail view.
func NewPRDetailView(term *terminal.Terminal, cfg DetailConfig) *PRDetailView {
	return &PRDetailView{
		BaseView: views.NewBaseView(term, views.ViewPRDetail, "Pull Request"),
		config:   cfg,
	}
}

// SetPullRequest sets the PR.
func (v *PRDetailView) SetPullRequest(pr *domain.PullRequest) {
	v.pr = pr
}

// Render renders the detail view.
func (v *PRDetailView) Render(startRow, width, height int) {
	if v.pr == nil {
		return
	}
	pr := v.pr
	term := v.Term()

	term.MoveTo(startRow, 2)
	icon := "🔀"
	if pr.IsDraft {
		icon = "📝"
	}
	fmt.Print(terminal.Style(fmt.Sprintf("%s Pull Request #%d", icon, pr.PullRequestID), terminal.Bold, terminal.FgCyan))

	term.MoveTo(startRow+2, 2)
	fmt.Print(terminal.Style(terminal.Truncate(pr.Title, width-4), terminal.Bold))

	term.MoveTo(startRow+4, 2)
	fmt.Print(terminal.Style("Status: ", terminal.Dim))
	statusStyle := terminal.FgGreen
	if pr.Status == "abandoned" {
		statusStyle = terminal.FgRed
	}
	fmt.Print(terminal.Style(strings.ToUpper(pr.Status), terminal.Bold, statusStyle))

	term.MoveTo(startRow+6, 2)
	fmt.Print(terminal.Style("Branch: ", terminal.Dim))
	fmt.Print(terminal.Style(pr.SourceBranch(), terminal.FgCyan))
	fmt.Print(terminal.Style(" → ", terminal.Dim))
	fmt.Print(terminal.Style(pr.TargetBranch(), terminal.FgGreen))

	term.MoveTo(startRow+8, 2)
	fmt.Print(terminal.Style("Created By: ", terminal.Dim))
	fmt.Print(pr.CreatedBy.DisplayName)

	term.MoveTo(startRow+8, width/2)
	fmt.Print(terminal.Style("Created: ", terminal.Dim))
	fmt.Print(pr.CreationDate.Format("Jan 2, 2006 15:04"))

	revRow := startRow + 10
	term.MoveTo(revRow, 2)
	fmt.Print(terminal.Style("─── Reviewers ", terminal.Dim))
	fmt.Print(terminal.Style(strings.Repeat("─", width-20), terminal.Dim))

	if len(pr.Reviewers) == 0 {
		term.MoveTo(revRow+1, 4)
		fmt.Print(terminal.Style("No reviewers", terminal.Dim))
	} else {
		for i, r := range pr.Reviewers {
			if revRow+1+i >= startRow+height-4 {
				break
			}
			term.MoveTo(revRow+1+i, 4)
			fmt.Printf("%s %s - %s", r.VoteIcon(), r.DisplayName, r.VoteStatus())
		}
	}

	term.MoveTo(startRow+height-2, 2)
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_git/%s/pullrequest/%d",
		v.config.Organization, v.config.Project, pr.Repository.Name, pr.PullRequestID)
	fmt.Print(terminal.Style("URL: "+terminal.Truncate(url, width-10), terminal.Dim))
}

// HandleKey handles input.
func (v *PRDetailView) HandleKey(key terminal.Key) bool { return false }

func formatDate(dateStr string) string {
	if dateStr == "" {
		return "-"
	}
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		if len(dateStr) > 10 {
			return dateStr[:10]
		}
		return dateStr
	}
	return t.Format("Jan 2, 2006")
}

func stripHTML(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
		} else if !inTag {
			result.WriteRune(r)
		}
	}
	text := result.String()
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	return strings.TrimSpace(text)
}

func wrapText(text string, width int) []string {
	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return lines
	}
	currentLine := words[0]
	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	return lines
}
