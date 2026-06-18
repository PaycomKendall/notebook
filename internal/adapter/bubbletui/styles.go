package bubbletui

import "github.com/charmbracelet/lipgloss"

// Styles holds the computed lipgloss styles for a Theme.
type Styles struct {
	title, dim, key, sel, tag, warn lipgloss.Style
	pane, paneFocused, modal        lipgloss.Style
}

// styles builds all styles from the theme's six colors.
func (t Theme) styles() Styles {
	return Styles{
		title:       lipgloss.NewStyle().Bold(true).Foreground(t.accent),
		dim:         lipgloss.NewStyle().Foreground(t.subtle),
		key:         lipgloss.NewStyle().Bold(true).Foreground(t.secondary),
		sel:         lipgloss.NewStyle().Foreground(t.selFg).Background(t.selBg).Bold(true),
		tag:         lipgloss.NewStyle().Foreground(t.accent),
		warn:        lipgloss.NewStyle().Foreground(t.warn),
		pane:        lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.secondary).Padding(0, 1),
		paneFocused: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.accent).Padding(0, 1),
		modal:       lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.accent).Padding(1, 2),
	}
}
