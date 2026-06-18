package bubbletui

import (
	"testing"

	"github.com/kendallowen/notebook/internal/todo"
)

func TestMoveKeyOpensMoveForm(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _, _ = s.AddTask("work", "alpha", nil, "") })
	m.focus = focusTasks
	send(m, key("m"))
	if m.mode != modeMoveTask {
		t.Errorf("m should open the move form; mode=%v", m.mode)
	}
}

func TestMoveFormMovesTask(t *testing.T) {
	m, svc := newTestModel(t, func(s *todo.Service) { _, _ = s.AddTask("work", "alpha", nil, "") })
	m.focus = focusTasks
	m.openMoveTask()
	typeStr(m, "archive")
	m.updateForm(key("enter"))
	if m.mode != modeNormal {
		t.Errorf("form should close after submit; mode=%v", m.mode)
	}
	work, _ := svc.GetList("work")
	if len(work.Tasks) != 0 {
		t.Errorf("task still in source: %+v", work.Tasks)
	}
	arch, err := svc.GetList("archive")
	if err != nil || len(arch.Tasks) != 1 {
		t.Fatalf("dest not populated: err=%v list=%+v", err, arch)
	}
}
