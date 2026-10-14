package ui

import (
	"claude-booths/ui/theme"
	"errors"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

type ErrBox struct {
	height, width int
	err           error
	info          bool // err is a notice, not a failure
}

var errStyle, infoStyle lipgloss.Style

func init() {
	theme.OnChange(func(p theme.Palette) {
		errStyle = lipgloss.NewStyle().Foreground(p.Danger)
		infoStyle = lipgloss.NewStyle().Foreground(p.Success)
	})
}

func NewErrBox() *ErrBox {
	return &ErrBox{}
}

func (e *ErrBox) SetError(err error) {
	e.err, e.info = err, false
}

// SetInfo shows a notice in the same slot, styled as success instead of error.
func (e *ErrBox) SetInfo(msg string) {
	e.err, e.info = errors.New(msg), true
}

func (e *ErrBox) Clear() {
	e.err = nil
}

func (e *ErrBox) SetSize(width, height int) {
	e.width = width
	e.height = height
}

func (e *ErrBox) String() string {
	var err string
	if e.err != nil {
		err = e.err.Error()
		lines := strings.Split(err, "\n")
		err = strings.Join(lines, "//")
		if runewidth.StringWidth(err) > e.width-3 && e.width-3 >= 0 {
			err = runewidth.Truncate(err, e.width-3, "...")
		}
	}
	style := errStyle
	if e.info {
		style = infoStyle
	}
	return lipgloss.Place(e.width, e.height, lipgloss.Center, lipgloss.Center, style.Render(err))
}
