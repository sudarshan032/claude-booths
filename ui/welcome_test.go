package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func plainWelcome(width, height int, withKeys bool) string {
	return ansi.Strip(renderWelcome(width, height, "No booths open yet.", withKeys))
}

func TestWelcomeFullSize(t *testing.T) {
	out := plainWelcome(100, 30, true)
	require.Contains(t, out, "claude booths")
	require.Contains(t, out, welcomeTagline)
	require.Contains(t, out, "▟█████▙", "logo")
	require.Contains(t, out, "new with prompt", "keys card")
	require.Contains(t, out, "No booths open yet.")
}

func TestWelcomeNeverWiderThanWidth(t *testing.T) {
	for _, w := range []int{12, 20, 30, 40, 60, 80, 100, 160} {
		for _, line := range strings.Split(renderWelcome(w, 40, "No booths open yet. Each booth is its own chat in its own git worktree.", true), "\n") {
			require.LessOrEqual(t, lipgloss.Width(line), w, "width %d: %q", w, ansi.Strip(line))
		}
	}
}

func TestWelcomeNarrowDropsLogo(t *testing.T) {
	out := plainWelcome(logoWidth-1, 40, true)
	require.NotContains(t, out, "▟█████▙")
	require.Contains(t, out, "claude booths")
}

func TestWelcomeShortDropsLogoThenKeys(t *testing.T) {
	full := renderWelcome(100, 0, "msg", true)
	h := lipgloss.Height(full)

	noLogo := plainWelcome(100, h-1, true)
	require.NotContains(t, noLogo, "▟█████▙")
	require.Contains(t, noLogo, "new booth")

	tiny := renderWelcome(100, 4, "msg", true)
	require.LessOrEqual(t, lipgloss.Height(tiny), 4)
	require.Contains(t, ansi.Strip(tiny), "claude booths")
}

func TestWelcomeKeysOnlyWhenAsked(t *testing.T) {
	require.NotContains(t, plainWelcome(100, 30, false), "new with prompt")
}
