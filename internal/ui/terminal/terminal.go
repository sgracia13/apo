// Package terminal provides low-level terminal control.
package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ANSI codes
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"
	Blink     = "\033[5m"
	Reverse   = "\033[7m"

	FgBlack   = "\033[30m"
	FgRed     = "\033[31m"
	FgGreen   = "\033[32m"
	FgYellow  = "\033[33m"
	FgBlue    = "\033[34m"
	FgMagenta = "\033[35m"
	FgCyan    = "\033[36m"
	FgWhite   = "\033[37m"

	BgBlue  = "\033[44m"
	BgWhite = "\033[47m"
)

// KeyType represents the type of key pressed.
type KeyType int

const (
	KeyRune KeyType = iota
	KeyEnter
	KeyEscape
	KeyBackspace
	KeyTab
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyCtrlC
)

// Key represents a key press.
type Key struct {
	Type KeyType
	Rune rune
}

// Terminal provides terminal control.
type Terminal struct {
	width, height int
}

// New creates a new Terminal.
func New() *Terminal {
	t := &Terminal{}
	t.UpdateSize()
	return t
}

// UpdateSize refreshes terminal dimensions.
func (t *Terminal) UpdateSize() {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err == nil {
		fmt.Sscanf(string(out), "%d %d", &t.height, &t.width)
	}
	if t.width == 0 {
		t.width = 120
	}
	if t.height == 0 {
		t.height = 40
	}
}

// Width returns terminal width.
func (t *Terminal) Width() int  { return t.width }

// Height returns terminal height.
func (t *Terminal) Height() int { return t.height }

// Clear clears the screen.
func (t *Terminal) Clear() { fmt.Print("\033[2J\033[H") }

// MoveTo moves cursor to position (1-indexed).
func (t *Terminal) MoveTo(row, col int) { fmt.Printf("\033[%d;%dH", row, col) }

// HideCursor hides the cursor.
func (t *Terminal) HideCursor() { fmt.Print("\033[?25l") }

// ShowCursor shows the cursor.
func (t *Terminal) ShowCursor() { fmt.Print("\033[?25h") }

// EnableRawMode enables raw terminal input.
func (t *Terminal) EnableRawMode() error {
	cmd := exec.Command("stty", "raw", "-echo")
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// DisableRawMode restores normal terminal input.
func (t *Terminal) DisableRawMode() error {
	cmd := exec.Command("stty", "-raw", "echo")
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// ReadKey reads a single key.
func (t *Terminal) ReadKey() (Key, error) {
	buf := make([]byte, 3)
	n, err := os.Stdin.Read(buf)
	if err != nil {
		return Key{}, err
	}

	if n == 1 {
		switch buf[0] {
		case 3:
			return Key{Type: KeyCtrlC}, nil
		case 9:
			return Key{Type: KeyTab}, nil
		case 13:
			return Key{Type: KeyEnter}, nil
		case 27:
			return Key{Type: KeyEscape}, nil
		case 127:
			return Key{Type: KeyBackspace}, nil
		default:
			return Key{Type: KeyRune, Rune: rune(buf[0])}, nil
		}
	}

	if n >= 3 && buf[0] == 27 && buf[1] == '[' {
		switch buf[2] {
		case 'A':
			return Key{Type: KeyUp}, nil
		case 'B':
			return Key{Type: KeyDown}, nil
		case 'C':
			return Key{Type: KeyRight}, nil
		case 'D':
			return Key{Type: KeyLeft}, nil
		}
	}

	return Key{Type: KeyRune, Rune: rune(buf[0])}, nil
}

// Style applies style codes to text.
func Style(text string, codes ...string) string {
	if len(codes) == 0 {
		return text
	}
	prefix := strings.Join(codes, "")
	return prefix + text + Reset
}

// Pad pads a string to width.
func Pad(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

// Truncate truncates with ellipsis.
func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
