package tui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// The TUI must open with focus on the Tasks pane so arrow keys move through
// tasks immediately (the Lists pane often holds a single list).
func TestInitialFocusIsTasksPane(t *testing.T) {
	app, svc := newTestApp(t)
	if _, err := svc.AddTask("work", "one", nil, ""); err != nil {
		t.Fatal(err)
	}
	app.buildUI()
	app.refreshLists()
	app.refreshTasks()
	app.refreshDetail()
	app.bindKeys()
	if app.focusIdx != 1 {
		t.Errorf("initial focus = %d, want 1 (Tasks pane)", app.focusIdx)
	}
}

// A footer must show keybinding hints, and the hints must change with the
// focused pane (so users discover Tab, arrows, and the per-pane actions).
func TestFooterShowsContextualHints(t *testing.T) {
	app, _ := newTestApp(t)
	app.buildUI()

	app.focus(1) // Tasks pane
	tasksHint := app.footer.GetText(true)
	for _, want := range []string{"Tab", "add", "done", "delete", "quit"} {
		if !strings.Contains(tasksHint, want) {
			t.Errorf("tasks footer missing %q; got %q", want, tasksHint)
		}
	}

	app.focus(0) // Lists pane
	listsHint := app.footer.GetText(true)
	for _, want := range []string{"Tab", "new folder", "rename", "quit"} {
		if !strings.Contains(listsHint, want) {
			t.Errorf("lists footer missing %q; got %q", want, listsHint)
		}
	}
}

// Regression: the key handler must cycle focus on Tab and pass arrow/vim keys
// through to the focused pane (this is what makes navigation actually work).
func TestKeyHandlerCyclesFocusAndPassesArrows(t *testing.T) {
	app, svc := newTestApp(t)
	if _, err := svc.AddTask("work", "one", nil, ""); err != nil {
		t.Fatal(err)
	}
	app.buildUI()
	app.refreshLists()
	app.refreshTasks()
	app.bindKeys()

	cap := app.app.GetInputCapture()
	if cap == nil {
		t.Fatal("no input capture installed")
	}
	start := app.focusIdx

	if res := cap(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)); res != nil {
		t.Error("Tab should be consumed (return nil)")
	}
	if app.focusIdx == start {
		t.Errorf("Tab did not change focus (still %d)", app.focusIdx)
	}

	if res := cap(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)); res == nil {
		t.Error("Down arrow should pass through to the focused pane, not be consumed")
	}
	if res := cap(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone)); res == nil {
		t.Error("'j' should pass through to the focused pane, not be consumed")
	}
}
