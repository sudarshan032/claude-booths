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
	require.Contains(t, out, "claude")
	require.Contains(t, out, welcomeTagline)
	require.Contains(t, out, "████████", "large pixel banner")
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

func TestWelcomeBannerSizeFollowsWidth(t *testing.T) {
	large := plainWelcome(pixelLargeWidth, 40, false)
	require.Contains(t, large, "██", "large banner uses two cells per pixel")
	require.Equal(t, pixelRows+1, strings.Count(strings.Split(large, welcomeTagline)[0], "\n")-1, "label + 7 pixel rows")

	compact := plainWelcome(pixelLargeWidth-1, 40, false)
	require.Contains(t, compact, "▀", "compact banner uses half blocks")
	require.NotContains(t, compact, "████████")

	prompt := plainWelcome(pixelCompactWidth-1, 40, false)
	require.Contains(t, prompt, "B[ ][▮]THS", "too narrow for pixels: prompt logo")
	require.NotContains(t, prompt, "▀")
}

func TestWelcomeShortShrinksBannerThenDropsKeys(t *testing.T) {
	full := renderWelcome(100, 0, "msg", true)
	h := lipgloss.Height(full)

	smaller := plainWelcome(100, h-1, true)
	require.NotContains(t, smaller, "████████", "large banner gives way first")
	require.Contains(t, smaller, "new booth")

	tiny := renderWelcome(100, 4, "msg", true)
	require.LessOrEqual(t, lipgloss.Height(tiny), 4)
	require.Contains(t, ansi.Strip(tiny), "B[ ][▮]THS")
}

func TestPromptLogo(t *testing.T) {
	require.Equal(t, promptLogoText, ansi.Strip(PromptLogo(true)))
	require.Equal(t, "❯ B[ ][▮]THS", ansi.Strip(PromptLogo(false)))
}

func TestWelcomeKeysOnlyWhenAsked(t *testing.T) {
	require.NotContains(t, plainWelcome(100, 30, false), "new with prompt")
}
