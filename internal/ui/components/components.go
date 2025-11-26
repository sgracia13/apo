// Package components provides reusable UI components.
package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/user/apo/internal/ui/terminal"
)

// Tab represents a navigation tab.
type Tab struct {
	ID, Name, Key, Icon string
}

// TabBar is a horizontal tab navigation component.
type TabBar struct {
	term   *terminal.Terminal
	tabs   []Tab
	active int
}

// NewTabBar creates a new tab bar.
func NewTabBar(term *terminal.Terminal, tabs []Tab) *TabBar {
	return &TabBar{term: term, tabs: tabs}
}

// SetActive sets the active tab by index.
func (t *TabBar) SetActive(index int) {
	if index >= 0 && index < len(t.tabs) {
		t.active = index
	}
}

// SetActiveByID sets the active tab by ID.
func (t *TabBar) SetActiveByID(id string) {
	for i, tab := range t.tabs {
		if tab.ID == id {
			t.active = i
			return
		}
	}
}

// ActiveTab returns the active tab.
func (t *TabBar) ActiveTab() *Tab {
	if t.active >= 0 && t.active < len(t.tabs) {
		return &t.tabs[t.active]
	}
	return nil
}

// Next moves to the next tab.
func (t *TabBar) Next() {
	t.active = (t.active + 1) % len(t.tabs)
}

// Render renders the tab bar.
func (t *TabBar) Render(row, startCol, width int) {
	t.term.MoveTo(row, startCol)
	var sb strings.Builder
	sb.WriteString(" ")

	for i, tab := range t.tabs {
		text := fmt.Sprintf(" %s %s [%s] ", tab.Icon, tab.Name, tab.Key)
		if i == t.active {
			sb.WriteString(terminal.Style(text, terminal.Bold, terminal.Reverse))
		} else {
			sb.WriteString(terminal.Style(text, terminal.Dim))
		}
		if i < len(t.tabs)-1 {
			sb.WriteString("│")
		}
	}
	fmt.Print(sb.String())

	t.term.MoveTo(row+1, startCol)
	fmt.Print(terminal.Style(strings.Repeat("─", width-startCol), terminal.Dim))
}

// StatusBar displays status and help.
type StatusBar struct {
	term        *terminal.Terminal
	message     string
	messageTime time.Time
	lastRefresh time.Time
	helpText    string
}

// NewStatusBar creates a new status bar.
func NewStatusBar(term *terminal.Terminal) *StatusBar {
	return &StatusBar{term: term}
}

// SetMessage sets a status message.
func (s *StatusBar) SetMessage(msg string) {
	s.message = msg
	s.messageTime = time.Now()
}

// SetLastRefresh sets the last refresh time.
func (s *StatusBar) SetLastRefresh(t time.Time) {
	s.lastRefresh = t
}

// SetHelp sets the help text.
func (s *StatusBar) SetHelp(text string) {
	s.helpText = text
}

// Render renders the status bar.
func (s *StatusBar) Render(row, width int) {
	s.term.MoveTo(row, 1)
	for i := 0; i < width; i++ {
		fmt.Print(terminal.Style("─", terminal.Dim))
	}

	if time.Since(s.messageTime) < 3*time.Second && s.message != "" {
		s.term.MoveTo(row, 3)
		fmt.Print(terminal.Style(" "+s.message+" ", terminal.FgYellow))
	}

	if !s.lastRefresh.IsZero() {
		refresh := fmt.Sprintf("Last refresh: %s", s.lastRefresh.Format("15:04:05"))
		s.term.MoveTo(row, width-len(refresh)-2)
		fmt.Print(terminal.Style(refresh, terminal.Dim))
	}

	s.term.MoveTo(row+1, 1)
	fmt.Print(terminal.Style(terminal.Pad(s.helpText, width), terminal.BgBlue, terminal.FgWhite))
}

// Input is a text input component.
type Input struct {
	term   *terminal.Terminal
	value  string
	prompt string
	active bool
}

// NewInput creates a new input.
func NewInput(term *terminal.Terminal, prompt string) *Input {
	return &Input{term: term, prompt: prompt}
}

// Value returns the input value.
func (i *Input) Value() string { return i.value }

// Clear clears the input.
func (i *Input) Clear() { i.value = "" }

// Activate activates input.
func (i *Input) Activate() { i.active = true }

// Deactivate deactivates input.
func (i *Input) Deactivate() { i.active = false }

// InsertChar inserts a character.
func (i *Input) InsertChar(c rune) { i.value += string(c) }

// Backspace removes last character.
func (i *Input) Backspace() {
	if len(i.value) > 0 {
		i.value = i.value[:len(i.value)-1]
	}
}

// Render renders the input.
func (i *Input) Render(row, col, width int) {
	i.term.MoveTo(row, col)
	fmt.Print(terminal.Pad("", width))
	i.term.MoveTo(row, col)
	fmt.Print(terminal.Style(i.prompt, terminal.FgGreen, terminal.Bold))
	fmt.Print(i.value)
	if i.active {
		fmt.Print(terminal.Style("█", terminal.Blink))
	}
}

// ListItem represents an item in a list.
type ListItem struct {
	ID, Icon, Label, Sublabel string
	Data                      interface{}
}

// List is a scrollable list component.
type List struct {
	term          *terminal.Terminal
	title         string
	items         []ListItem
	filtered      []int
	selected      int
	scroll        int
	height        int
	filterMode    bool
	filterQuery   string
}

// NewList creates a new list.
func NewList(term *terminal.Terminal, title string) *List {
	return &List{term: term, title: title}
}

// SetItems sets the list items.
func (l *List) SetItems(items []ListItem) {
	l.items = items
	l.ClearFilter()
}

// SelectedIndex returns the selected index.
func (l *List) SelectedIndex() int { return l.selected }

// SelectedItem returns the selected item.
func (l *List) SelectedItem() *ListItem {
	indices := l.activeIndices()
	if l.selected >= 0 && l.selected < len(indices) {
		idx := indices[l.selected]
		if idx < len(l.items) {
			return &l.items[idx]
		}
	}
	return nil
}

// MoveUp moves selection up.
func (l *List) MoveUp() {
	if l.selected > 0 {
		l.selected--
		l.adjustScroll()
	}
}

// MoveDown moves selection down.
func (l *List) MoveDown() {
	if l.selected < len(l.activeIndices())-1 {
		l.selected++
		l.adjustScroll()
	}
}

// MoveToTop jumps to top.
func (l *List) MoveToTop() {
	l.selected = 0
	l.scroll = 0
}

// MoveToBottom jumps to bottom.
func (l *List) MoveToBottom() {
	indices := l.activeIndices()
	if len(indices) > 0 {
		l.selected = len(indices) - 1
		l.adjustScroll()
	}
}

// SetFilter sets the filter query.
func (l *List) SetFilter(query string) {
	l.filterQuery = query
	l.applyFilter()
}

// ClearFilter clears the filter.
func (l *List) ClearFilter() {
	l.filterQuery = ""
	l.filtered = nil
	l.selected = 0
	l.scroll = 0
}

// ToggleFilterMode toggles filter mode.
func (l *List) ToggleFilterMode() {
	l.filterMode = !l.filterMode
	if !l.filterMode {
		l.ClearFilter()
	}
}

// IsFilterMode returns filter mode state.
func (l *List) IsFilterMode() bool { return l.filterMode }

// FilterQuery returns the filter query.
func (l *List) FilterQuery() string { return l.filterQuery }

func (l *List) activeIndices() []int {
	if len(l.filtered) > 0 {
		return l.filtered
	}
	indices := make([]int, len(l.items))
	for i := range l.items {
		indices[i] = i
	}
	return indices
}

func (l *List) applyFilter() {
	if l.filterQuery == "" {
		l.filtered = nil
		return
	}
	query := strings.ToLower(l.filterQuery)
	l.filtered = []int{}
	for i, item := range l.items {
		if strings.Contains(strings.ToLower(item.Label), query) {
			l.filtered = append(l.filtered, i)
		}
	}
	l.selected = 0
	l.scroll = 0
}

func (l *List) adjustScroll() {
	if l.height <= 0 {
		return
	}
	if l.selected < l.scroll {
		l.scroll = l.selected
	}
	if l.selected >= l.scroll+l.height {
		l.scroll = l.selected - l.height + 1
	}
}

// Render renders the list.
func (l *List) Render(startRow, startCol, width, height int) {
	l.height = height - 3

	l.term.MoveTo(startRow, startCol)
	titleText := l.title
	if len(l.filtered) > 0 {
		titleText = fmt.Sprintf("%s (filtered: %d/%d)", l.title, len(l.filtered), len(l.items))
	} else {
		titleText = fmt.Sprintf("%s (%d)", l.title, len(l.items))
	}
	fmt.Print(terminal.Style(titleText, terminal.Bold, terminal.FgYellow))

	if l.filterMode {
		l.term.MoveTo(startRow, width-30)
		fmt.Print(terminal.Style("🔍 ", terminal.Dim))
		fmt.Print(l.filterQuery)
		fmt.Print(terminal.Style("█", terminal.Blink))
	}

	l.term.MoveTo(startRow+1, startCol)
	fmt.Print(terminal.Style(strings.Repeat("─", width-startCol-1), terminal.Dim))

	indices := l.activeIndices()
	row := startRow + 2

	visibleEnd := l.scroll + l.height
	if visibleEnd > len(indices) {
		visibleEnd = len(indices)
	}

	for i := l.scroll; i < visibleEnd; i++ {
		idx := indices[i]
		if idx >= len(l.items) {
			continue
		}
		item := l.items[idx]

		l.term.MoveTo(row, startCol)
		line := item.Icon + " " + terminal.Truncate(item.Label, width-startCol-10)

		if i == l.selected {
			fmt.Print(terminal.Style(terminal.Pad(line, width-startCol-1), terminal.Reverse))
		} else {
			fmt.Print(line)
		}
		row++
	}

	if len(l.items) == 0 {
		l.term.MoveTo(row, startCol+2)
		fmt.Print(terminal.Style("No items", terminal.Dim))
	}

	if l.scroll > 0 {
		l.term.MoveTo(startRow+2, width-2)
		fmt.Print(terminal.Style("▲", terminal.FgYellow))
	}
	if visibleEnd < len(indices) {
		l.term.MoveTo(startRow+l.height+1, width-2)
		fmt.Print(terminal.Style("▼", terminal.FgYellow))
	}
}
