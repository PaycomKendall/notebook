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
