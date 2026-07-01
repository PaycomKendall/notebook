package bubbletui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/kendallowen/notebook/internal/markdown"
)

// Styles holds the computed lipgloss styles for a Theme.
type Styles struct {
	title, titleFocused, dim, key, sel, selDim, tag, warn lipgloss.Style
	pane, paneFocused, modal, page                        lipgloss.Style
}

// styles builds all styles from the theme's palette.
//
// For "paper" themes (t.bg set) every style that renders text onto the page
// also carries the page background: a lipgloss style that sets only a
// foreground emits a reset at its end, which would otherwise expose the
// terminal's own background for those cells and leave holes in the paper.
func (t Theme) styles() Styles {
	// base carries the page background (nothing for terminal-native themes) so
	// every derived style inherits it.
	base := lipgloss.NewStyle()
	if t.bg != nil {
		base = base.Background(t.bg)
	}

	pane := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.secondary).Padding(0, 1)
	// Focused pane uses a thick border shape (not just a recolor) so focus
	// survives themes where the two border colors look alike.
	paneFocused := lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderForeground(t.accent).Padding(0, 1)
	modal := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.accent).Padding(1, 2)
	if t.bg != nil {
		// Paint the page: fill pane interiors and border cells with paper, and
		// render otherwise-unstyled body text (task/list rows) as ink.
		pane = pane.Background(t.bg).BorderBackground(t.bg).Foreground(t.fg)
		paneFocused = paneFocused.Background(t.bg).BorderBackground(t.bg).Foreground(t.fg)
		modal = modal.Background(t.bg).BorderBackground(t.bg).Foreground(t.fg)
	}

	return Styles{
		title: base.Bold(true).Foreground(t.accent),
		// Focused pane title reads as a filled "tab" chip so the active pane is
		// obvious even when accent/secondary border colors are close.
		titleFocused: lipgloss.NewStyle().Bold(true).Foreground(t.selFg).Background(t.accent).Padding(0, 1),
		dim:          base.Foreground(t.subtle),
		key:          base.Bold(true).Foreground(t.secondary),
		sel:          lipgloss.NewStyle().Foreground(t.selFg).Background(t.selBg).Bold(true),
		// selDim marks the selected row in an unfocused pane: keep the cursor
		// but drop the background so only the focused pane has a bright row.
		selDim:      base.Foreground(t.subtle),
		tag:         base.Foreground(t.accent),
		warn:        base.Foreground(t.warn),
		pane:        pane,
		paneFocused: paneFocused,
		modal:       modal,
		// page fills the whole terminal (footer + margins) with paper so the
		// notebook look extends edge to edge; empty for terminal-native themes.
		page: base,
	}
}

// mdStyles maps the active theme to markdown element styles for the Detail pane.
// Each starts from the page background (empty for terminal-native themes) so
// notes render on paper without reset holes.
func (m *Model) mdStyles() markdown.Styles {
	base := lipgloss.NewStyle()
	if m.theme.bg != nil {
		base = base.Background(m.theme.bg).Foreground(m.theme.fg)
	}
	return markdown.Styles{
		H1:     base.Bold(true).Foreground(m.themeAccent()),
		H2:     base.Bold(true).Foreground(m.themeAccent()),
		H3:     base.Bold(true).Foreground(m.themeAccent()),
		Bold:   base.Bold(true),
		Italic: base.Italic(true),
		Code:   m.styles.tag,
		Bullet: m.styles.tag,
		Base:   base,
	}
}

func (m *Model) themeAccent() lipgloss.TerminalColor { return m.theme.accent }
