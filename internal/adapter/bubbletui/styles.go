package bubbletui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/kendallowen/notebook/internal/markdown"
)

// Styles holds the computed lipgloss styles for a Theme.
type Styles struct {
	title, titleFocused, dim, key, sel, selDim, tag, warn lipgloss.Style
	pane, paneFocused, modal                              lipgloss.Style
}

// styles builds all styles from the theme's six colors.
func (t Theme) styles() Styles {
	return Styles{
		title: lipgloss.NewStyle().Bold(true).Foreground(t.accent),
		// Focused pane title reads as a filled "tab" chip so the active pane is
		// obvious even when accent/secondary border colors are close.
		titleFocused: lipgloss.NewStyle().Bold(true).Foreground(t.selFg).Background(t.accent).Padding(0, 1),
		dim:          lipgloss.NewStyle().Foreground(t.subtle),
		key:          lipgloss.NewStyle().Bold(true).Foreground(t.secondary),
		sel:          lipgloss.NewStyle().Foreground(t.selFg).Background(t.selBg).Bold(true),
		// selDim marks the selected row in an unfocused pane: keep the cursor
		// but drop the background so only the focused pane has a bright row.
		selDim: lipgloss.NewStyle().Foreground(t.subtle),
		tag:    lipgloss.NewStyle().Foreground(t.accent),
		warn:   lipgloss.NewStyle().Foreground(t.warn),
		pane:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.secondary).Padding(0, 1),
		// Focused pane uses a thick border shape (not just a recolor) so focus
		// survives themes where the two border colors look alike.
		paneFocused: lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderForeground(t.accent).Padding(0, 1),
		modal:       lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.accent).Padding(1, 2),
	}
}

// mdStyles maps the active theme to markdown element styles for the Detail pane.
func (m *Model) mdStyles() markdown.Styles {
	return markdown.Styles{
		H1:     lipgloss.NewStyle().Bold(true).Foreground(m.themeAccent()),
		H2:     lipgloss.NewStyle().Bold(true).Foreground(m.themeAccent()),
		H3:     lipgloss.NewStyle().Bold(true).Foreground(m.themeAccent()),
		Bold:   lipgloss.NewStyle().Bold(true),
		Italic: lipgloss.NewStyle().Italic(true),
		Code:   m.styles.tag,
		Bullet: m.styles.tag,
	}
}

func (m *Model) themeAccent() lipgloss.TerminalColor { return m.theme.accent }
