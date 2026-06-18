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
