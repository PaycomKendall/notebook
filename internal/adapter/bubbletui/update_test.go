package bubbletui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kendallowen/notebook/internal/todo"
)

func send(m *Model, k tea.KeyMsg) tea.Cmd {
	_, cmd := m.Update(k)
	return cmd
}

func TestTabCyclesFocus(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _, _ = s.AddTask("work", "alpha", nil, "") })
	if m.focus != focusLists {
		t.Fatalf("initial focus = %v, want lists", m.focus)
	}
	send(m, key("tab"))
	if m.focus != focusTasks {
		t.Errorf("after tab focus = %v, want tasks", m.focus)
	}
	send(m, key("tab"))
	send(m, key("tab")) // wraps detail -> lists
	if m.focus != focusLists {
		t.Errorf("focus after wrap = %v, want lists", m.focus)
	}
}

func TestArrowMovesTaskSelection(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) {
		_, _ = s.AddTask("work", "alpha", nil, "")
		_, _ = s.AddTask("work", "beta", nil, "")
	})
	m.focus = focusTasks
	send(m, key("j"))
	if m.taskIdx != 1 {
		t.Errorf("taskIdx after down = %d, want 1", m.taskIdx)
	}
	send(m, key("j")) // clamp at last
	if m.taskIdx != 1 {
		t.Errorf("taskIdx clamped = %d, want 1", m.taskIdx)
	}
}

func TestDToggles(t *testing.T) {
	m, svc := newTestModel(t, func(s *todo.Service) { _, _ = s.AddTask("work", "alpha", nil, "") })
	m.focus = focusTasks
	send(m, key("d"))
	l, _ := svc.GetList("work")
	if !l.Tasks[0].Done {
		t.Error("d did not toggle done")
	}
}

func TestAOpensAddForm(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _, _ = s.AddTask("work", "alpha", nil, "") })
	m.focus = focusTasks
	send(m, key("a"))
	if m.mode != modeAddTask {
		t.Errorf("a should open add form; mode=%v", m.mode)
	}
}

func TestListsPaneAOpensNewList(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _, _ = s.AddTask("work", "alpha", nil, "") })
	m.focus = focusLists
	send(m, key("a"))
	if m.mode != modeNewList {
		t.Errorf("a in lists pane should open new-list form; mode=%v", m.mode)
	}
}

func TestQuitReturnsQuitCmd(t *testing.T) {
	m, _ := newTestModel(t, nil)
	cmd := send(m, key("q"))
	if cmd == nil {
		t.Fatal("q should return a quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("q command should produce tea.QuitMsg")
	}
}

func TestFormKeyRoutesToForm(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _ = s.CreateList("work") })
	m.focus = focusLists
	send(m, key("a")) // open new-list form
	// typing 'q' must go to the form, not quit
	send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if m.mode != modeNewList {
		t.Errorf("typing in a form must not trigger global keys; mode=%v", m.mode)
	}
	if m.inputs[0].Value() != "q" {
		t.Errorf("form field should contain 'q'; got %q", m.inputs[0].Value())
	}
}
