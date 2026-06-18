package bubbletui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kendallowen/notebook/internal/todo"
)

func key(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEscape}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func typeStr(m *Model, s string) {
	for _, r := range s {
		m.updateForm(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
}

func TestAddTaskFormSubmits(t *testing.T) {
	m, svc := newTestModel(t, func(s *todo.Service) { _ = s.CreateList("work") })
	m.openAddTask()
	if m.mode != modeAddTask || len(m.inputs) != 3 {
		t.Fatalf("openAddTask state: mode=%v inputs=%d", m.mode, len(m.inputs))
	}
	typeStr(m, "buy milk")
	m.updateForm(key("tab")) // -> Tags
	typeStr(m, "store urgent")
	m.updateForm(key("enter"))

	if m.mode != modeNormal {
		t.Errorf("form should close on submit; mode=%v", m.mode)
	}
	l, _ := svc.GetList("work")
	if len(l.Tasks) != 1 || l.Tasks[0].Title != "buy milk" {
		t.Fatalf("task not added: %+v", l.Tasks)
	}
	if len(l.Tasks[0].Tags) != 2 {
		t.Errorf("tags = %v, want 2", l.Tasks[0].Tags)
	}
}

func TestEmptyTitleKeepsFormOpen(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _ = s.CreateList("work") })
	m.openAddTask()
	m.updateForm(key("enter")) // no title typed
	if m.mode != modeAddTask {
		t.Errorf("empty title should keep form open; mode=%v", m.mode)
	}
}

func TestEscCancelsForm(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _ = s.CreateList("work") })
	m.openAddTask()
	typeStr(m, "discard me")
	m.updateForm(key("esc"))
	if m.mode != modeNormal {
		t.Errorf("esc should cancel; mode=%v", m.mode)
	}
}

func TestConfirmDeletesTask(t *testing.T) {
	m, svc := newTestModel(t, func(s *todo.Service) { _, _ = s.AddTask("work", "alpha", nil, "") })
	m.confirmDeleteTask()
	if m.mode != modeConfirm {
		t.Fatalf("mode = %v, want confirm", m.mode)
	}
	m.updateConfirm(key("y"))
	l, _ := svc.GetList("work")
	if len(l.Tasks) != 0 {
		t.Errorf("task not deleted: %+v", l.Tasks)
	}
}

func TestValidationClearsStaleStatus(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _ = s.CreateList("work") })
	m.status = "old error"
	m.openAddTask()
	m.updateForm(key("enter")) // empty title -> validation early return
	if m.status != "" {
		t.Errorf("stale status should be cleared on submit; got %q", m.status)
	}
	if m.mode != modeAddTask {
		t.Errorf("form should stay open on empty title; mode=%v", m.mode)
	}
}

func TestFormViewShowsFieldsAndHint(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _ = s.CreateList("work") })
	m.openAddTask()
	out := m.View()
	for _, want := range []string{"Add task", "Title", "Notes", "esc", "cancel"} {
		if !strings.Contains(out, want) {
			t.Errorf("formView missing %q\n%s", want, out)
		}
	}
}
