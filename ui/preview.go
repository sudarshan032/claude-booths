package ui

import (
	"claude-style/session"
	"claude-style/ui/theme"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

var previewPaneStyle lipgloss.Style

func init() {
	theme.OnChange(func(p theme.Palette) { previewPaneStyle = lipgloss.NewStyle().Foreground(p.Text) })
}

type PreviewPane struct {
	width  int
	height int

	previewState previewState
	isScrolling  bool
	viewport     viewport.Model
}

type previewState struct {
	// fallback is true if the preview pane is displaying fallback text
	fallback bool
	// text is the text displayed in the preview pane
	text string
}

func NewPreviewPane() *PreviewPane {
	return &PreviewPane{
		viewport: viewport.New(0, 0),
	}
}

func (p *PreviewPane) SetSize(width, maxHeight int) {
	p.width = width
	p.height = maxHeight
	p.viewport.Width = width
	p.viewport.Height = max(maxHeight-1, 0) // one row for the scroll-mode header
}

// scrollStep is how many lines one wheel notch or shift+arrow moves.
const scrollStep = 3

// enterScroll loads the pane's full scrollback into the viewport, positioned at the bottom.
func (p *PreviewPane) enterScroll(instance *session.Instance) error {
	content, err := instance.PreviewFullHistory()
	if err != nil {
		return err
	}
	p.viewport.SetContent(strings.TrimRight(content, "\n"))
	p.viewport.GotoBottom()
	p.isScrolling = true
	return nil
}

// scrollHeader is the first row in scroll mode: where you are and how to get out.
func (p *PreviewPane) scrollHeader() string {
	pal := theme.Current()
	label := lipgloss.NewStyle().Background(pal.Accent).Foreground(pal.OnAccent).Bold(true).Render(" SCROLL ")
	hint := lipgloss.NewStyle().Foreground(pal.Muted).Render(fmt.Sprintf(
		" %3.0f%%  wheel / shift+↑↓ to scroll · Esc to go back to live", p.viewport.ScrollPercent()*100))
	return label + hint
}

// setFallbackState sets the preview state with fallback text and a message
func (p *PreviewPane) setFallbackState(message string) {
	p.previewState = previewState{
		fallback: true,
		text:     lipgloss.JoinVertical(lipgloss.Center, FallBackText, "", message),
	}
}

// Updates the preview pane content with the tmux pane content
func (p *PreviewPane) UpdateContent(instance *session.Instance) error {
	switch {
	case instance == nil:
		p.setFallbackState("No agents running yet. Spin up a new instance with 'n' to get started!")
		return nil
	case instance.Status == session.Loading:
		p.setFallbackState("Setting up workspace...")
		return nil
	case instance.Status == session.Paused:
		p.setFallbackState(lipgloss.JoinVertical(lipgloss.Center,
			"Session is paused. Press 'r' to resume.",
			"",
			lipgloss.NewStyle().
				Foreground(theme.Current().Warning).
				Render(fmt.Sprintf(
					"The instance can be checked out at '%s' (copied to your clipboard)",
					instance.Branch,
				)),
		))
		return nil
	}

	var content string
	var err error

	if !p.isScrolling {
		// In normal mode, use the usual preview
		content, err = instance.Preview()
		if err != nil {
			return err
		}

		// Always update the preview state with content, even if empty
		// This ensures that newly created instances will display their content immediately
		if len(content) == 0 && !instance.Started() {
			p.setFallbackState("Please enter a name for the instance.")
		} else {
			// Update the preview state with the current content
			p.previewState = previewState{
				fallback: false,
				text:     content,
			}
		}
	}

	return nil
}

// Returns the preview pane content as a string.
func (p *PreviewPane) String() string {
	if p.width == 0 || p.height == 0 {
		return strings.Repeat("\n", p.height)
	}

	if p.previewState.fallback {
		// Calculate available height for fallback text
		availableHeight := p.height - 3 - 4 // 2 for borders, 1 for margin, 1 for padding

		// Count the number of lines in the fallback text
		fallbackLines := len(strings.Split(p.previewState.text, "\n"))

		// Calculate padding needed above and below to center the content
		totalPadding := availableHeight - fallbackLines
		topPadding := 0
		bottomPadding := 0
		if totalPadding > 0 {
			topPadding = totalPadding / 2
			bottomPadding = totalPadding - topPadding // accounts for odd numbers
		}

		// Build the centered content
		var lines []string
		if topPadding > 0 {
			lines = append(lines, strings.Repeat("\n", topPadding))
		}
		lines = append(lines, p.previewState.text)
		if bottomPadding > 0 {
			lines = append(lines, strings.Repeat("\n", bottomPadding))
		}

		// Center both vertically and horizontally
		return previewPaneStyle.
			Width(p.width).
			Align(lipgloss.Center).
			Render(strings.Join(lines, ""))
	}

	// Scroll mode: frozen snapshot of the full history, with a header saying so
	if p.isScrolling {
		return lipgloss.JoinVertical(lipgloss.Left, p.scrollHeader(), p.viewport.View())
	}

	// Normal mode display
	// Calculate available height accounting for border and margin
	availableHeight := p.height

	// capture-pane ends every row with \n; trim it so a full pane fits exactly instead of
	// losing its last row (the agent's input box / status line) to a "..." marker.
	lines := strings.Split(strings.TrimRight(p.previewState.text, "\n"), "\n")

	// If taller than the pane, keep the bottom: that's where the agent's live output and prompt are
	if availableHeight > 0 {
		if len(lines) > availableHeight {
			lines = lines[len(lines)-availableHeight:]
		} else {
			// Pad with empty lines to fill available height
			padding := availableHeight - len(lines)
			lines = append(lines, make([]string, padding)...)
		}
	}

	content := strings.Join(lines, "\n")
	rendered := previewPaneStyle.Width(p.width).Render(content)
	return rendered
}

// ScrollUp scrolls up, entering scroll mode (at the bottom of the history) first if needed.
func (p *PreviewPane) ScrollUp(instance *session.Instance) error {
	if instance == nil || instance.Status == session.Paused {
		return nil
	}
	if !p.isScrolling {
		if err := p.enterScroll(instance); err != nil {
			return err
		}
	}
	p.viewport.LineUp(scrollStep)
	return nil
}

// ScrollDown scrolls down; scrolling past the bottom returns to the live preview.
func (p *PreviewPane) ScrollDown(instance *session.Instance) error {
	if instance == nil || instance.Status == session.Paused || !p.isScrolling {
		return nil // nothing below the live view
	}
	if p.viewport.AtBottom() {
		return p.ResetToNormalMode(instance)
	}
	p.viewport.LineDown(scrollStep)
	return nil
}

// ResetToNormalMode exits scroll mode and returns to normal mode
func (p *PreviewPane) ResetToNormalMode(instance *session.Instance) error {
	if instance == nil || instance.Status == session.Paused {
		return nil
	}

	if p.isScrolling {
		p.isScrolling = false
		// Reset viewport
		p.viewport.SetContent("")
		p.viewport.GotoTop()

		// Immediately update content instead of waiting for next UpdateContent call
		content, err := instance.Preview()
		if err != nil {
			return err
		}
		p.previewState.text = content
	}

	return nil
}
