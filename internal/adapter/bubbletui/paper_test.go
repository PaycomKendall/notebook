package bubbletui

import (
	"regexp"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/kendallowen/notebook/internal/todo"
	"github.com/muesli/termenv"
)

// TestPaperThemeHasNoBackgroundHoles guards the "full notebook" look: on a
// paper theme (Theme.bg set) every cell of a rendered frame must sit on the
// page background. A style reset (ESC[0m) followed by visible spaces before any
// background is re-set exposes the terminal's own background — a hole in the
// paper. This regressed once via unstyled width-padding in the markdown
// renderer and the notebook page header, so lock it down.
func TestPaperThemeHasNoBackgroundHoles(t *testing.T) {
	// lipgloss strips color when it can't detect a color terminal (as in CI),
	// which would make this check vacuous. Force a 256-color profile for the
	// test and restore the previous one so we don't affect other tests.
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })

	m, _ := newTestModel(t, func(s *todo.Service) {
		_, _ = s.AddTask("work", "alpha task", []string{"urgent"},
			"# Heading\n\nSome **bold** note text.\n\n- bullet one\n- bullet two")
		_, _ = s.AddTask("work", "beta task", nil, "plain body")
	})
	m.theme = themeNotebook
	m.styles = themeNotebook.styles()
	m.width, m.height = 100, 24

	out := m.normalView()

	// A reset, then one or more spaces, then another reset with no background
	// set in between is a paper hole.
	hole := regexp.MustCompile(`\x1b\[0?m {1,}\x1b\[0?m`)
	if loc := hole.FindStringIndex(out); loc != nil {
		start := loc[0] - 48
		if start < 0 {
			start = 0
		}
		t.Errorf("paper-leak hole (unstyled spaces between resets) near: %q", out[start:loc[1]])
	}
}
