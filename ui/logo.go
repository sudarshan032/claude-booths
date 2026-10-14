package ui

import (
	"claude-booths/ui/theme"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// The pixel wordmark: "booths" in a 7-row bitmap where the two o's are booths, the
// second one lit. Pixel values: 1 letter, 2 booth outline, 3 lit booth, 4 window in the
// lit booth (left blank so the terminal background shows through), 0 empty.
var pixelGlyphs = []string{
	"1000/1000/1110/1001/1001/1001/1110",        // b
	"00000/00000/02220/20002/20002/20002/22222", // o, empty booth
	"00000/00000/03330/33333/33433/33333/33333", // O, lit booth
	"0100/0100/1111/0100/0100/0100/0011",        // t
	"1000/1000/1110/1001/1001/1001/1001",        // h
	"0000/0000/0111/1000/0110/0001/1110",        // s
}

const pixelRows = 7

// pixelGrid lays the glyphs side by side with one empty column between them.
var pixelGrid = func() [][]byte {
	grid := make([][]byte, pixelRows)
	for gi, g := range pixelGlyphs {
		for y, row := range strings.Split(g, "/") {
			if gi > 0 {
				grid[y] = append(grid[y], '0')
			}
			grid[y] = append(grid[y], row...)
		}
	}
	return grid
}()

var pixelCols = len(pixelGrid[0])

// Widths of the two banner sizes: large uses two cells per pixel (square-ish pixels),
// compact packs two pixel rows into one text row with half blocks.
var (
	pixelLargeWidth   = pixelCols * 2
	pixelCompactWidth = pixelCols
)

var (
	pixelStyles                                                        map[byte]lipgloss.Style
	pixelColors                                                        map[byte]lipgloss.TerminalColor
	promptMarkStyle, promptDimStyle, promptWordStyle, promptBoothStyle lipgloss.Style
	promptLampStyle                                                    lipgloss.Style
)

func init() {
	theme.OnChange(func(p theme.Palette) {
		pixelColors = map[byte]lipgloss.TerminalColor{'1': p.Text, '2': p.Accent, '3': p.Warning}
		pixelStyles = map[byte]lipgloss.Style{}
		for k, c := range pixelColors {
			pixelStyles[k] = lipgloss.NewStyle().Foreground(c)
		}
		promptMarkStyle = lipgloss.NewStyle().Foreground(p.Warning).Bold(true)
		promptDimStyle = lipgloss.NewStyle().Foreground(p.Muted)
		promptWordStyle = lipgloss.NewStyle().Foreground(p.Text).Bold(true)
		promptBoothStyle = lipgloss.NewStyle().Foreground(p.Accent).Bold(true)
		promptLampStyle = lipgloss.NewStyle().Foreground(p.Warning)
	})
}

// renderPixelBanner draws the pixel wordmark, two cells per pixel when large, half
// blocks otherwise.
func renderPixelBanner(large bool) string {
	var rows []string
	if large {
		for _, line := range pixelGrid {
			var b strings.Builder
			for _, v := range line {
				if st, ok := pixelStyles[v]; ok {
					b.WriteString(st.Render("██"))
				} else {
					b.WriteString("  ")
				}
			}
			rows = append(rows, b.String())
		}
		return strings.Join(rows, "\n")
	}
	for y := 0; y < pixelRows; y += 2 {
		var b strings.Builder
		for x := 0; x < pixelCols; x++ {
			top, bottom := pixelGrid[y][x], byte('0')
			if y+1 < pixelRows {
				bottom = pixelGrid[y+1][x]
			}
			tc, topOn := pixelColors[top]
			bc, bottomOn := pixelColors[bottom]
			switch {
			case topOn && bottomOn && top == bottom:
				b.WriteString(pixelStyles[top].Render("█"))
			case topOn && bottomOn:
				b.WriteString(lipgloss.NewStyle().Foreground(tc).Background(bc).Render("▀"))
			case topOn:
				b.WriteString(pixelStyles[top].Render("▀"))
			case bottomOn:
				b.WriteString(pixelStyles[bottom].Render("▄"))
			default:
				b.WriteByte(' ')
			}
		}
		rows = append(rows, b.String())
	}
	return strings.Join(rows, "\n")
}

// promptLogoText is the plain form of PromptLogo, for width checks.
const promptLogoText = "❯ claude/b[ ][▮]ths"

// PromptLogo is the one-line logo for small places: the list title, help, narrow panes.
// Pass withClaude=false to drop the "claude/" part where space is tight.
func PromptLogo(withClaude bool) string {
	s := promptMarkStyle.Render("❯") + " "
	if withClaude {
		s += promptDimStyle.Render("claude/")
	}
	return s + promptWordStyle.Render("b") + promptBoothStyle.Render("[ ][") + promptLampStyle.Render("▮") +
		promptBoothStyle.Render("]") + promptWordStyle.Render("ths")
}
