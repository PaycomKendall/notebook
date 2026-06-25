package tui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Pressing Esc in a form must cancel it and restore the main three-pane view.
func TestEscCancelsForm(t *testing.T) {
	app, _ := newTestApp(t)
	app.buildUI()
	app.refreshLists()
	app.refreshTasks()
	app.refreshDetail()
	app.bindKeys()

	form := tview.NewForm().AddInputField("X", "", 10, nil, nil)
	app.showModalForm(form, " Test ")
	if app.paneFocused() {
		t.Fatal("a form should be open here (panes must not be focused)")
	}

	// Esc through the form's own input handler triggers its cancel func.
	form.InputHandler()(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone), func(tview.Primitive) {})

	if !app.paneFocused() {
		t.Error("Esc should cancel the form and restore the main (pane) view")
	}
}

// A form must show a hint footer mentioning Tab and Esc so navigation/cancel
// are discoverable while the main hint bar is hidden.
func TestFormShowsHintFooter(t *testing.T) {
	app, _ := newTestApp(t)
	app.buildUI()
	app.refreshLists()
	app.refreshTasks()
	app.refreshDetail()
	app.bindKeys()

	form := tview.NewForm().AddInputField("X", "", 10, nil, nil)
	app.showModalForm(form, " Test ")

	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	screen.SetSize(80, 12)
	app.app.SetScreen(screen)
	go func() { _ = app.app.Run() }()
	defer app.app.Stop()
	app.app.Draw()

	if !screenContains(screen, "Esc") || !screenContains(screen, "Tab") {
		t.Error("form should display a hint footer mentioning Tab and Esc")
	}
}

// Verifies a multi-line note round-trips through the edit form's Save handler.
func TestEditFormAcceptsMultilineNotes(t *testing.T) {
	app, svc := newTestApp(t)
	if _, err := svc.AddTask("inbox", "task", nil, "old"); err != nil {
		t.Fatal(err)
	}
	app.buildUI()
	app.refreshLists()
	app.refreshTasks()
	app.editTaskForm() // builds the form and records app.lastForm

	item := app.lastForm.GetFormItemByLabel("Notes")
	ta, ok := item.(*tview.TextArea)
	if !ok {
		t.Fatalf("Notes item is %T, want *tview.TextArea", item)
	}
	ta.SetText("line one\nline two", true)

	// Trigger the Save button (index 0) via its input handler.
	app.lastForm.GetButton(0).InputHandler()(
		tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone),
		func(tview.Primitive) {},
	)

	l, _ := svc.GetList("inbox")
	if l.Tasks[0].Notes != "line one\nline two" {
		t.Errorf("notes = %q, want multi-line", l.Tasks[0].Notes)
	}
}
