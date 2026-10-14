package ui

import (
	"claude-booths/ui/theme"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// The booth logo: three arches on a floor, the middle one lit. Each row is the left booth,
// the lit middle booth and the right booth, so the parts can take different colors.
var logoRows = [][3]string{
	{"       ", " ▄▄▄▄▄ ", "       "},
	{" ╭───╮ ", "▟█████▙", " ╭───╮ "},
	{" │ ▬ │ ", "██   ██", " │ ▬ │ "},
	{" │   │ ", "███████", " │   │ "},
}

const logoFloor = "━━━━━━━━━━━━━━━━━━━━━"

var logoWidth = lipgloss.Width(logoFloor)

// welcomeKeys is the keys card on the empty screen.
var welcomeKeys = [][2]string{
	{"n", "new booth"},
	{"N", "new with prompt"},
	{"↵", "step inside"},
	{"ctrl+q", "step out"},
	{"t", "theme"},
	{"?", "help"},
}

const welcomeTagline = "Many chats. One terminal."

var (
	logoBoothStyle, logoLampStyle, wordClaudeStyle, wordBoothsStyle lipgloss.Style
	taglineStyle, welcomeKeyStyle, welcomeDescStyle, keysCardStyle  lipgloss.Style
	welcomeMsgStyle                                                 lipgloss.Style
)

func init() {
	theme.OnChange(func(p theme.Palette) {
		logoBoothStyle = lipgloss.NewStyle().Foreground(p.Accent)
		logoLampStyle = lipgloss.NewStyle().Foreground(p.Warning)
		wordClaudeStyle = lipgloss.NewStyle().Foreground(p.Text).Bold(true)
		wordBoothsStyle = lipgloss.NewStyle().Foreground(p.Accent).Bold(true)
		taglineStyle = lipgloss.NewStyle().Foreground(p.Muted)
		welcomeKeyStyle = lipgloss.NewStyle().Foreground(p.Warning).Bold(true)
		welcomeDescStyle = lipgloss.NewStyle().Foreground(p.Muted)
		keysCardStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Faint).Padding(0, 2)
		welcomeMsgStyle = lipgloss.NewStyle().Foreground(p.Text)
	})
}

func renderLogo() string {
	rows := make([]string, 0, len(logoRows)+1)
	for _, r := range logoRows {
		rows = append(rows, logoBoothStyle.Render(r[0])+logoLampStyle.Render(r[1])+logoBoothStyle.Render(r[2]))
	}
	rows = append(rows, logoBoothStyle.Render(logoFloor))
	return strings.Join(rows, "\n")
}

// renderKeysCard lays the keys out in two columns, or one when cols is 1; each column
// aligns its own keys.
func renderKeysCard(cols int) string {
	type colW struct{ key, desc int }
	widths := make([]colW, cols)
	for i, k := range welcomeKeys {
		c := &widths[i%cols]
		c.key = max(c.key, lipgloss.Width(k[0]))
		c.desc = max(c.desc, lipgloss.Width(k[1]))
	}
	var rows []string
	for i := 0; i < len(welcomeKeys); i += cols {
		var cells []string
		for c := 0; c < cols && i+c < len(welcomeKeys); c++ {
			k, w := welcomeKeys[i+c], widths[c]
			cells = append(cells, welcomeKeyStyle.Render(k[0]+strings.Repeat(" ", w.key-lipgloss.Width(k[0])))+"  "+
				welcomeDescStyle.Render(k[1]+strings.Repeat(" ", w.desc-lipgloss.Width(k[1]))))
		}
		rows = append(rows, strings.Join(cells, "     "))
	}
	return keysCardStyle.Render(strings.Join(rows, "\n"))
}

// renderWelcome draws the screen shown when a pane has nothing live to show: the logo,
// the wordmark, the tagline, optionally the keys card, then message. It fits the result into
// width x height by dropping the logo first, then the keys card, then the tagline.
func renderWelcome(width, height int, message string, withKeys bool) string {
	if width <= 0 {
		return ""
	}
	wordmark := wordClaudeStyle.Render("claude") + " " + wordBoothsStyle.Render("booths")
	msg := ""
	if message != "" {
		msg = welcomeMsgStyle.Width(min(width, 64)).Align(lipgloss.Center).Render(message)
	}

	if width < len("claude booths") {
		wordmark = wordBoothsStyle.Render("booths")
	}
	logo, tagline, card := "", "", ""
	if width >= lipgloss.Width(welcomeTagline) {
		tagline = taglineStyle.Render(welcomeTagline)
	}
	if width >= logoWidth {
		logo = renderLogo()
	}
	if withKeys {
		for _, cols := range []int{2, 1} {
			if c := renderKeysCard(cols); lipgloss.Width(c) <= width {
				card = c
				break
			}
		}
	}

	build := func() string {
		var parts []string
		add := func(s string) {
			if s == "" {
				return
			}
			if len(parts) > 0 {
				parts = append(parts, "")
			}
			parts = append(parts, s)
		}
		if logo != "" {
			add(logo)
		}
		if tagline != "" {
			add(wordmark + "\n" + tagline)
		} else {
			add(wordmark)
		}
		add(card)
		add(msg)
		return lipgloss.JoinVertical(lipgloss.Center, parts...)
	}

	out := build()
	for _, drop := range []*string{&logo, &card, &tagline} {
		if height <= 0 || lipgloss.Height(out) <= height {
			break
		}
		*drop = ""
		out = build()
	}
	return out
}
