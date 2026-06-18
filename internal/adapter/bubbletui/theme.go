package bubbletui

import "github.com/charmbracelet/lipgloss"

// Theme is the six-color palette every style derives from.
type Theme struct {
	accent    lipgloss.TerminalColor // focused border, tags, titles
	secondary lipgloss.TerminalColor // unfocused borders, help keys
	subtle    lipgloss.TerminalColor // dim/help text
	selBg     lipgloss.TerminalColor // selected-row background
	selFg     lipgloss.TerminalColor // selected-row text
	warn      lipgloss.TerminalColor // status/errors
}

// themeDefault is the original Charm look, adaptive to light/dark terminals.
var themeDefault = Theme{
	accent:    lipgloss.AdaptiveColor{Light: "205", Dark: "212"},
	secondary: lipgloss.AdaptiveColor{Light: "63", Dark: "99"},
	subtle:    lipgloss.AdaptiveColor{Light: "244", Dark: "245"},
	selBg:     lipgloss.AdaptiveColor{Light: "189", Dark: "57"},
	selFg:     lipgloss.AdaptiveColor{Light: "236", Dark: "231"},
	warn:      lipgloss.AdaptiveColor{Light: "160", Dark: "203"},
}
