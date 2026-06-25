package bubbletui

import (
	"strings"
	"testing"

	"github.com/kendallowen/notebook/internal/todo"
)

func TestNormalViewShowsPanesAndFooter(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) {
		_, _ = s.AddTask("work", "alpha", []string{"hr"}, "a note")
		_, _ = s.AddTask("work", "beta", nil, "")
	})
	m.focus = focusTasks
	out := m.View()

	for _, want := range []string{"Lists", "work", "Tasks", "alpha", "beta", "[ ]", "Detail", "a note"} {
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
		_, _ = s.AddTask("work", "task", nil, "# Plan\n- step one")
	})
	m.focus = focusDetail
	out := m.renderDetail()
	if !strings.Contains(out, "◦") {
		t.Errorf("expected spiral binding gutter, got:\n%s", out)
	}
	if !strings.Contains(out, "N O T E B O O K") {
		t.Errorf("expected header band, got:\n%s", out)
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
