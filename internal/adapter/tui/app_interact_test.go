package tui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestToggleKeyMarksTaskDone(t *testing.T) {
	app, svc := newTestApp(t)
	if _, err := svc.AddTask("work", "ship it", nil, ""); err != nil {
		t.Fatal(err)
	}

	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	screen.SetSize(100, 24)

	app.buildUI()
	app.refreshLists()
	app.refreshTasks()
	app.bindKeys()
	app.app.SetScreen(screen)
	go func() { _ = app.app.Run() }()
	defer app.app.Stop()

	// Focus the tasks pane and toggle the first task.
	app.app.QueueUpdateDraw(func() {
		app.focus(1) // tasks pane
		app.toggleSelected()
	})
	app.app.QueueUpdate(func() {}) // flush

	l, _ := svc.GetList("work")
	if !l.Tasks[0].Done {
		t.Errorf("expected task to be done after toggle; got %+v", l.Tasks[0])
	}
}

func TestFocusCyclesWithinRange(t *testing.T) {
	app, _ := newTestApp(t)
	app.buildUI()
	app.focus(0)
	app.cycleFocus(1)
	if app.focusIdx != 1 {
		t.Errorf("focusIdx = %d, want 1", app.focusIdx)
	}
	app.cycleFocus(1)
	app.cycleFocus(1) // wraps 2 -> 0
	if app.focusIdx != 0 {
		t.Errorf("focusIdx after wrap = %d, want 0", app.focusIdx)
	}
}
