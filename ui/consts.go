package ui

import (
	"claude-booths/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

var bannerArt = lipgloss.JoinVertical(lipgloss.Center, `
░█████╗░██╗░░░░░░█████╗░██╗░░░██╗██████╗░███████╗
██╔══██╗██║░░░░░██╔══██╗██║░░░██║██╔══██╗██╔════╝
██║░░╚═╝██║░░░░░███████║██║░░░██║██║░░██║█████╗░░
██║░░██╗██║░░░░░██╔══██║██║░░░██║██║░░██║██╔══╝░░
╚█████╔╝███████╗██║░░██║╚██████╔╝██████╔╝███████╗
░╚════╝░╚══════╝╚═╝░░╚═╝░╚═════╝░╚═════╝░╚══════╝
`, `
██████╗░░█████╗░░█████╗░████████╗██╗░░██╗░██████╗
██╔══██╗██╔══██╗██╔══██╗╚══██╔══╝██║░░██║██╔════╝
██████╔╝██║░░██║██║░░██║░░░██║░░░███████║╚█████╗░
██╔══██╗██║░░██║██║░░██║░░░██║░░░██╔══██║░╚═══██╗
██████╔╝╚█████╔╝╚█████╔╝░░░██║░░░██║░░██║██████╔╝
╚═════╝░░╚════╝░░╚════╝░░░░╚═╝░░░╚═╝░░╚═╝╚═════╝░
`)

// FallBackText is the banner shown when there's nothing to preview, in the theme's accent.
var FallBackText string

func init() {
	theme.OnChange(func(p theme.Palette) { FallBackText = lipgloss.NewStyle().Foreground(p.Accent).Render(bannerArt) })
}
