package tui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestMoveKeyOpensForm(t *testing.T) {
	app, svc := newTestApp(t)
	if _, err := svc.AddTask("work", "alpha", nil, ""); err != nil {
		t.Fatal(err)
	}
	app.buildUI()
	app.refreshLists()
	app.refreshTasks()
	app.bindKeys()
	app.focus(1) // Tasks pane

	cap := app.app.GetInputCapture()
	cap(tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModNone))
	if app.paneFocused() {
		t.Error("'m' in the Tasks pane should open the move form (panes no longer focused)")
	}
}
