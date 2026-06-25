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

func TestNotesFieldIsMultiline(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _ = s.CreateList("work") })
	m.openAddTask()
	// Notes is the 3rd field (index 2).
	if !m.inputs[2].multiline() {
		t.Fatalf("Notes field should be multiline")
	}
}

func TestEnterInNotesInsertsNewlineNotSubmit(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _ = s.CreateList("work") })
	m.openAddTask()
	typeStr(m, "title")
	m.formField = 2 // Notes
	m.refocusInputs()
	typeStr(m, "line one")
	m.updateForm(key("enter")) // should NOT submit; should add a newline
	if m.mode != modeAddTask {
		t.Fatalf("enter in multiline Notes must not submit; mode=%v", m.mode)
	}
	typeStr(m, "line two")
	if !strings.Contains(m.inputs[2].Value(), "\n") {
		t.Errorf("notes should contain a newline, got %q", m.inputs[2].Value())
	}
}

func TestCtrlSSubmitsFromNotes(t *testing.T) {
	m, svc := newTestModel(t, func(s *todo.Service) { _ = s.CreateList("work") })
	m.openAddTask()
	typeStr(m, "buy milk")
	m.formField = 2
	m.refocusInputs()
	typeStr(m, "2%")
	m.updateForm(tea.KeyMsg{Type: tea.KeyCtrlS})
	if m.mode != modeNormal {
		t.Fatalf("ctrl+s should submit; mode=%v", m.mode)
	}
	l, _ := svc.GetList("work")
	if len(l.Tasks) != 1 || l.Tasks[0].Notes != "2%" {
		t.Fatalf("task not added with notes: %+v", l.Tasks)
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
