package ui

import (
	"github.com/sudarshan032/claude-booths/ui/theme"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// welcomeKeys is the keys card on the empty screen.
var welcomeKeys = [][2]string{
	{"n", "new booth"},
	{"N", "new with prompt"},
	{"↵", "step inside"},
	{"ctrl+q", "step out"},
	{"t", "theme"},
	{"?", "help"},
}

const welcomeTagline = "Claude Code in parallel, one booth per task."

var (
	bannerLabelStyle, taglineStyle, welcomeKeyStyle, welcomeDescStyle lipgloss.Style
	keysCardStyle, welcomeMsgStyle                                    lipgloss.Style
)

func init() {
	theme.OnChange(func(p theme.Palette) {
		bannerLabelStyle = lipgloss.NewStyle().Foreground(p.Accent).Bold(true)
		taglineStyle = lipgloss.NewStyle().Foreground(p.Muted)
		welcomeKeyStyle = lipgloss.NewStyle().Foreground(p.Warning).Bold(true)
		welcomeDescStyle = lipgloss.NewStyle().Foreground(p.Muted)
		keysCardStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Faint).Padding(0, 2)
		welcomeMsgStyle = lipgloss.NewStyle().Foreground(p.Text)
	})
}

// renderBanner is the pixel wordmark with a small "claude" label above its left edge,
// large or compact.
func renderBanner(large bool) string {
	banner := renderPixelBanner(large)
	label := bannerLabelStyle.Render("claude")
	return lipgloss.JoinVertical(lipgloss.Left, label, banner)
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

// renderWelcome draws the screen shown when a pane has nothing live to show: the pixel
// wordmark (or the one-line prompt logo when it doesn't fit), the tagline, optionally the
// keys card, then message. It fits the result into width x height by shrinking the banner,
// then dropping it for the prompt logo, then dropping the keys card, then the tagline.
func renderWelcome(width, height int, message string, withKeys bool) string {
	if width <= 0 {
		return ""
	}
	msg := ""
	if message != "" {
		msg = welcomeMsgStyle.Width(min(width, 64)).Align(lipgloss.Center).Render(message)
	}

	// Logo candidates, biggest first; the last one always fits.
	var logos []string
	if width >= pixelLargeWidth {
		logos = append(logos, renderBanner(true))
	}
	if width >= pixelCompactWidth {
		logos = append(logos, renderBanner(false))
	}
	if width >= lipgloss.Width(promptLogoText) {
		logos = append(logos, PromptLogo(true))
	} else {
		logos = append(logos, PromptLogo(false))
	}
	if width < lipgloss.Width(PromptLogo(false)) {
		logos = append(logos, promptWordStyle.Render("booths"))
	}
	logo := logos[0]

	tagline, card := "", ""
	if width >= lipgloss.Width(welcomeTagline) {
		tagline = taglineStyle.Render(welcomeTagline)
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
		if tagline != "" {
			add(logo + "\n\n" + tagline)
		} else {
			add(logo)
		}
		add(card)
		add(msg)
		return lipgloss.JoinVertical(lipgloss.Center, parts...)
	}

	out := build()
	fits := func() bool { return height <= 0 || lipgloss.Height(out) <= height }
	for i := 1; i < len(logos) && !fits(); i++ {
		logo = logos[i]
		out = build()
	}
	for _, drop := range []*string{&card, &tagline} {
		if fits() {
			break
		}
		*drop = ""
		out = build()
	}
	return out
}
