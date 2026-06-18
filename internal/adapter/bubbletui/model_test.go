package bubbletui

import (
	"testing"

	"github.com/kendallowen/notebook/internal/adapter/jsonstore"
	"github.com/kendallowen/notebook/internal/todo"
)

// newTestModel builds a Model over a temp-dir store, optionally seeded.
func newTestModel(t *testing.T, seed func(*todo.Service)) (*Model, *todo.Service) {
	t.Helper()
	store, err := jsonstore.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	svc := todo.NewService(store)
	if seed != nil {
		seed(svc)
	}
	return New(svc), svc
}

func TestNewLoadsListsAndTasks(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) {
		_, _ = s.AddTask("work", "alpha", nil, "")
		_, _ = s.AddTask("work", "beta", nil, "")
	})
	if len(m.listNames) != 1 || m.listNames[0] != "work" {
		t.Fatalf("listNames = %v, want [work]", m.listNames)
	}
	if m.current == nil || len(m.current.Tasks) != 2 {
		t.Fatalf("current tasks not loaded: %+v", m.current)
	}
	if got := m.selectedTask(); got == nil || got.Title != "alpha" {
		t.Errorf("selectedTask = %+v, want alpha", got)
	}
}

func TestReloadTasksClampsIndex(t *testing.T) {
	m, svc := newTestModel(t, func(s *todo.Service) {
		_, _ = s.AddTask("work", "alpha", nil, "")
		_, _ = s.AddTask("work", "beta", nil, "")
	})
	m.taskIdx = 1
	if err := svc.RemoveTask("work", 2); err != nil {
		t.Fatal(err)
	}
	m.reloadTasks()
	if m.taskIdx != 0 {
		t.Errorf("taskIdx = %d, want clamped to 0", m.taskIdx)
	}
	if m.selectedTask() == nil {
		t.Error("selectedTask should be valid after clamp")
	}
}
