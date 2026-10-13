// Package theme holds the color palettes of the TUI. Styles elsewhere are built from
// Current() inside an OnChange hook, so switching themes at runtime restyles everything.
package theme

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Palette assigns a color to each role in the UI. Backgrounds are left to the terminal.
type Palette struct {
	Name string

	Accent   lipgloss.TerminalColor // borders, tabs, selection highlight, title badge
	OnAccent lipgloss.TerminalColor // text drawn on Accent

	Text  lipgloss.TerminalColor // body text
	Muted lipgloss.TerminalColor // secondary text, hints
	Faint lipgloss.TerminalColor // separators, disabled items

	SelectedBg lipgloss.TerminalColor // selected session in the list
	SelectedFg lipgloss.TerminalColor

	Success lipgloss.TerminalColor // ready dot, added lines
	Danger  lipgloss.TerminalColor // errors, removed lines, confirm dialogs
	Warning lipgloss.TerminalColor // paused hint, help keys
	Info    lipgloss.TerminalColor // diff hunks, help headers

	Menu lipgloss.TerminalColor // highlighted menu action
}

// DefaultName is used when the config names no theme or an unknown one.
const DefaultName = "claude-style"

var builtins = map[string]Palette{
	"claude-style": {
		Accent: lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}, OnAccent: lipgloss.Color("#FFFFDF"),
		Text: lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#DDDDDD"}, Muted: lipgloss.Color("#808080"), Faint: lipgloss.AdaptiveColor{Light: "#DDDADA", Dark: "#3C3C3C"},
		SelectedBg: lipgloss.Color("#DDE4F0"), SelectedFg: lipgloss.Color("#1A1A1A"),
		Success: lipgloss.Color("#51BD73"), Danger: lipgloss.Color("#DE613E"), Warning: lipgloss.Color("#FFCC00"), Info: lipgloss.Color("#36CFC9"),
		Menu: lipgloss.Color("#AF87FF"),
	},
	"monokai": {
		Accent: lipgloss.Color("#F92672"), OnAccent: lipgloss.Color("#272822"),
		Text: lipgloss.Color("#F8F8F2"), Muted: lipgloss.Color("#A59F85"), Faint: lipgloss.Color("#49483E"),
		SelectedBg: lipgloss.Color("#49483E"), SelectedFg: lipgloss.Color("#F8F8F2"),
		Success: lipgloss.Color("#A6E22E"), Danger: lipgloss.Color("#F92672"), Warning: lipgloss.Color("#E6DB74"), Info: lipgloss.Color("#66D9EF"),
		Menu: lipgloss.Color("#FD971F"),
	},
	"vscode": {
		Accent: lipgloss.Color("#007ACC"), OnAccent: lipgloss.Color("#FFFFFF"),
		Text: lipgloss.Color("#D4D4D4"), Muted: lipgloss.Color("#858585"), Faint: lipgloss.Color("#3C3C3C"),
		SelectedBg: lipgloss.Color("#04395E"), SelectedFg: lipgloss.Color("#FFFFFF"),
		Success: lipgloss.Color("#4EC9B0"), Danger: lipgloss.Color("#F44747"), Warning: lipgloss.Color("#DCDCAA"), Info: lipgloss.Color("#9CDCFE"),
		Menu: lipgloss.Color("#C586C0"),
	},
	"ocean-green": {
		Accent: lipgloss.Color("#2BB3A3"), OnAccent: lipgloss.Color("#0B1F1D"),
		Text: lipgloss.Color("#D8F3EF"), Muted: lipgloss.Color("#7FA9A3"), Faint: lipgloss.Color("#24413D"),
		SelectedBg: lipgloss.Color("#164A44"), SelectedFg: lipgloss.Color("#E6FFFB"),
		Success: lipgloss.Color("#5BE49B"), Danger: lipgloss.Color("#FF6B6B"), Warning: lipgloss.Color("#F7D774"), Info: lipgloss.Color("#4FC3F7"),
		Menu: lipgloss.Color("#7CF5D8"),
	},
	"dracula": {
		Accent: lipgloss.Color("#BD93F9"), OnAccent: lipgloss.Color("#282A36"),
		Text: lipgloss.Color("#F8F8F2"), Muted: lipgloss.Color("#6272A4"), Faint: lipgloss.Color("#44475A"),
		SelectedBg: lipgloss.Color("#44475A"), SelectedFg: lipgloss.Color("#F8F8F2"),
		Success: lipgloss.Color("#50FA7B"), Danger: lipgloss.Color("#FF5555"), Warning: lipgloss.Color("#F1FA8C"), Info: lipgloss.Color("#8BE9FD"),
		Menu: lipgloss.Color("#FF79C6"),
	},
	"nord": {
		Accent: lipgloss.Color("#88C0D0"), OnAccent: lipgloss.Color("#2E3440"),
		Text: lipgloss.Color("#ECEFF4"), Muted: lipgloss.Color("#7B88A1"), Faint: lipgloss.Color("#434C5E"),
		SelectedBg: lipgloss.Color("#3B4252"), SelectedFg: lipgloss.Color("#ECEFF4"),
		Success: lipgloss.Color("#A3BE8C"), Danger: lipgloss.Color("#BF616A"), Warning: lipgloss.Color("#EBCB8B"), Info: lipgloss.Color("#81A1C1"),
		Menu: lipgloss.Color("#B48EAD"),
	},
	"gruvbox": {
		Accent: lipgloss.Color("#FE8019"), OnAccent: lipgloss.Color("#282828"),
		Text: lipgloss.Color("#EBDBB2"), Muted: lipgloss.Color("#A89984"), Faint: lipgloss.Color("#504945"),
		SelectedBg: lipgloss.Color("#504945"), SelectedFg: lipgloss.Color("#FBF1C7"),
		Success: lipgloss.Color("#B8BB26"), Danger: lipgloss.Color("#FB4934"), Warning: lipgloss.Color("#FABD2F"), Info: lipgloss.Color("#83A598"),
		Menu: lipgloss.Color("#D3869B"),
	},
	"solarized": {
		Accent: lipgloss.Color("#268BD2"), OnAccent: lipgloss.Color("#FDF6E3"),
		Text: lipgloss.Color("#93A1A1"), Muted: lipgloss.Color("#657B83"), Faint: lipgloss.Color("#073642"),
		SelectedBg: lipgloss.Color("#073642"), SelectedFg: lipgloss.Color("#EEE8D5"),
		Success: lipgloss.Color("#859900"), Danger: lipgloss.Color("#DC322F"), Warning: lipgloss.Color("#B58900"), Info: lipgloss.Color("#2AA198"),
		Menu: lipgloss.Color("#D33682"),
	},
}

var (
	current = withName(DefaultName)
	hooks   []func(Palette)
)

func withName(name string) Palette {
	p := builtins[name]
	p.Name = name
	return p
}

// Current returns the active palette.
func Current() Palette { return current }

// Names lists the built-in themes, default first, the rest alphabetically.
func Names() []string {
	names := make([]string, 0, len(builtins))
	for n := range builtins {
		if n != DefaultName {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	return append([]string{DefaultName}, names...)
}

// OnChange registers f to rebuild styles from a palette and runs it once right away.
// Hooks run on the caller's goroutine; the TUI only switches themes from Update.
func OnChange(f func(Palette)) {
	hooks = append(hooks, f)
	f(current)
}

// Set activates a theme by name (case-insensitive; "" means the default).
func Set(name string) error {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		name = DefaultName
	}
	if _, ok := builtins[name]; !ok {
		return fmt.Errorf("unknown theme %q (available: %s)", name, strings.Join(Names(), ", "))
	}
	current = withName(name)
	for _, f := range hooks {
		f(current)
	}
	return nil
}

// Next returns the theme after the current one in Names(), wrapping around.
func Next() string {
	names := Names()
	for i, n := range names {
		if n == current.Name {
			return names[(i+1)%len(names)]
		}
	}
	return DefaultName
}

// Hex returns c as #RRGGBB for tools outside lipgloss (tmux), resolving adaptive colors
// for a dark background. ok is false for colors without a fixed RGB value.
func Hex(c lipgloss.TerminalColor) (hex string, ok bool) {
	switch v := c.(type) {
	case lipgloss.AdaptiveColor:
		c = lipgloss.Color(v.Dark)
	case lipgloss.Color:
	default:
		return "", false
	}
	s := string(c.(lipgloss.Color))
	if len(s) == 7 && s[0] == '#' {
		return s, true
	}
	return "", false
}
