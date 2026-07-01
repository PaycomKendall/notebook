package bubbletui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/kendallowen/notebook/internal/todo"
	"github.com/muesli/termenv"
)

// With a real color profile active, the notebook chrome must not corrupt
// pre-styled lines. Wrapping already-styled text in another lipgloss style (the
// underline rule) emits the embedded ESC bytes bare, so the terminal prints the
// rest of the sequence ("[1m") as literal text. The corruption signature is two
// consecutive ESC bytes. The note here carries an inline bold span so the body
// ruled-line path is exercised, not just the (now header-band) title.
func TestDetailNotebookNoStyleCorruption(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	m, _ := newTestModel(t, func(s *todo.Service) {
		_, _ = s.AddTask("work", "Module history", nil, "WIP: **Position**")
	})
	m.focus = focusDetail
	out := m.renderDetail()

	if strings.Contains(out, "\x1b\x1b") {
		t.Errorf("double-ESC corruption from styling already-styled text:\n%q", out)
	}
	if !strings.Contains(out, "Module history") {
		t.Errorf("task title missing from the page header band:\n%q", out)
	}
}

func TestNormalViewShowsPanesAndFooter(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) {
		_, _ = s.AddTask("work", "alpha", []string{"hr"}, "a note")
		_, _ = s.AddTask("work", "beta", nil, "")
	})
	m.focus = focusTasks
	out := m.View()

	for _, want := range []string{"Folders", "work", "Pages", "alpha", "beta", "[ ]", "Detail", "a note"} {
		if !strings.Contains(out, want) {
			t.Errorf("View missing %q\n%s", want, out)
		}
	}
	// selected task marker on the focused row
	if !strings.Contains(out, "❯") {
		t.Errorf("View missing selection marker\n%s", out)
	}
	// tasks-pane footer hints
	for _, want := range []string{"add", "done", "edit", "delete", "quit"} {
		if !strings.Contains(out, want) {
			t.Errorf("footer missing %q\n%s", want, out)
		}
	}
}

func TestFocusedPaneHasDistinctBorder(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _, _ = s.AddTask("work", "alpha", nil, "") })
	m.focus = focusTasks
	out := m.View()
	// The focused pane uses a thick border; the others stay rounded. Both
	// shapes must appear, so focus is visible regardless of palette.
	if !strings.Contains(out, "┏") {
		t.Errorf("focused pane should use a thick border (┏)\n%s", out)
	}
	if !strings.Contains(out, "╭") {
		t.Errorf("unfocused panes should keep the rounded border (╭)\n%s", out)
	}
}

func TestDetailShowsNotebookChrome(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) {
		_, _ = s.AddTask("work", "Groceries run", nil, "# Plan\n- step one")
	})
	m.focus = focusDetail
	out := m.renderDetail()
	if !strings.Contains(out, "◦") {
		t.Errorf("expected spiral binding gutter, got:\n%s", out)
	}
	if !strings.Contains(out, "Groceries run") {
		t.Errorf("expected task title in the header band, got:\n%s", out)
	}
	if !strings.Contains(out, "•") {
		t.Errorf("expected rendered bullet, got:\n%s", out)
	}
	if strings.Contains(out, "# Plan") {
		t.Errorf("markdown header marker should be stripped, got:\n%s", out)
	}
}

func TestDetailEmptyNoteStillLined(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) {
		_, _ = s.AddTask("work", "task", nil, "")
	})
	m.focus = focusDetail
	out := m.renderDetail()
	// Gutter glyph appears on filler rows even with no note text.
	if strings.Count(out, "◦") < 2 {
		t.Errorf("empty note should still render lined page, got:\n%s", out)
	}
}

func TestDetailPaneWiderThanTasks(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _ = s.CreateList("work") })
	m.width = 120
	m.height = 30
	_, tasks, detail := m.paneWidths()
	if detail <= tasks {
		t.Errorf("detail (%d) should be wider than tasks (%d)", detail, tasks)
	}
}

func TestListsFooterDiffersFromTasks(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _, _ = s.AddTask("work", "alpha", nil, "") })
	m.focus = focusLists
	out := m.View()
	for _, want := range []string{"new", "rename"} {
		if !strings.Contains(out, want) {
			t.Errorf("lists footer missing %q\n%s", want, out)
		}
	}
}
