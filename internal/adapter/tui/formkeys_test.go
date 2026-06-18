package tui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// While a form/modal is open it must own ALL key input: the global shortcuts
// must not fire, so Tab navigates between fields and letters type literally
// (previously `q` would quit the app and Tab was hijacked for pane switching).
func TestGlobalShortcutsDoNotFireWhileFormOpen(t *testing.T) {
	app, svc := newTestApp(t)
	if _, err := svc.AddTask("work", "task", nil, ""); err != nil {
		t.Fatal(err)
	}
	app.buildUI()
	app.refreshLists()
	app.refreshTasks()
	app.bindKeys()
	app.focus(1)       // Tasks pane
	app.editTaskForm() // open the edit form (now the root)

	cap := app.app.GetInputCapture()
	if cap == nil {
		t.Fatal("no input capture installed")
	}

	// Every key must pass through to the form (non-nil return = not consumed).
	for _, tc := range []struct {
		name string
		ev   *tcell.EventKey
	}{
		{"q", tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone)},
		{"a", tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone)},
		{"x", tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone)},
		{"d", tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone)},
		{"e", tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone)},
		{"Tab", tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)},
	} {
		if res := cap(tc.ev); res == nil {
			t.Errorf("key %q was consumed by the global capture while a form is open; "+
				"it must pass through to the form", tc.name)
		}
	}
}

// Sanity: in the main view (a pane focused) the global shortcuts STILL fire.
func TestGlobalShortcutsFireInMainView(t *testing.T) {
	app, svc := newTestApp(t)
	if _, err := svc.AddTask("work", "task", nil, ""); err != nil {
		t.Fatal(err)
	}
	app.buildUI()
	app.refreshLists()
	app.refreshTasks()
	app.bindKeys() // ends focused on a pane

	cap := app.app.GetInputCapture()
	if res := cap(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)); res != nil {
		t.Error("Tab should still cycle panes (be consumed) in the main view")
	}
}
