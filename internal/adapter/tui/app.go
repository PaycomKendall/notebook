package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kendallowen/notebook/internal/todo"
	"github.com/rivo/tview"
)

// App is the tview front-end over the shared Service.
type App struct {
	svc *todo.Service
	app *tview.Application

	lists  *tview.List
	tasks  *tview.List
	detail *tview.TextView

	listNames []string
	current   *todo.List // currently displayed list
}

// New creates a TUI App.
func New(svc *todo.Service) *App {
	return &App{svc: svc, app: tview.NewApplication()}
}

// buildUI constructs the three-pane layout.
func (a *App) buildUI() {
	a.lists = tview.NewList().ShowSecondaryText(false)
	a.lists.SetBorder(true).SetTitle(" Lists ")

	a.tasks = tview.NewList().ShowSecondaryText(false)
	a.tasks.SetBorder(true).SetTitle(" Tasks ")

	a.detail = tview.NewTextView().SetDynamicColors(true)
	a.detail.SetBorder(true).SetTitle(" Detail ")

	a.lists.SetChangedFunc(func(int, string, string, rune) {
		a.refreshTasks()
		a.refreshDetail()
	})
	a.tasks.SetChangedFunc(func(int, string, string, rune) {
		a.refreshDetail()
	})

	flex := tview.NewFlex().
		AddItem(a.lists, 24, 0, true).
		AddItem(a.tasks, 0, 2, false).
		AddItem(a.detail, 0, 2, false)

	a.app.SetRoot(flex, true)
}

// refreshLists reloads the Lists pane from the service.
func (a *App) refreshLists() {
	names, err := a.svc.ListNames()
	if err != nil {
		return
	}
	a.listNames = names
	a.lists.Clear()
	for _, name := range names {
		a.lists.AddItem(name, "", 0, nil)
	}
}

// selectedListName returns the highlighted list name (or "").
func (a *App) selectedListName() string {
	i := a.lists.GetCurrentItem()
	if i < 0 || i >= len(a.listNames) {
		return ""
	}
	return a.listNames[i]
}

// refreshTasks reloads the Tasks pane for the selected list.
func (a *App) refreshTasks() {
	a.tasks.Clear()
	name := a.selectedListName()
	if name == "" {
		a.current = nil
		return
	}
	l, err := a.svc.GetList(name)
	if err != nil {
		a.current = nil
		return
	}
	a.current = l
	for _, task := range l.Tasks {
		box := "[ ]"
		if task.Done {
			box = "[x]"
		}
		label := fmt.Sprintf("%s #%d %s", box, task.ID, task.Title)
		if len(task.Tags) > 0 {
			label += "  #" + strings.Join(task.Tags, " #")
		}
		a.tasks.AddItem(label, "", 0, nil)
	}
}

// selectedTask returns the highlighted task pointer (or nil).
func (a *App) selectedTask() *todo.Task {
	if a.current == nil {
		return nil
	}
	i := a.tasks.GetCurrentItem()
	if i < 0 || i >= len(a.current.Tasks) {
		return nil
	}
	return &a.current.Tasks[i]
}

// refreshDetail updates the Detail pane for the selected task.
func (a *App) refreshDetail() {
	t := a.selectedTask()
	if t == nil {
		a.detail.SetText("")
		return
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", t.Title)
	if len(t.Tags) > 0 {
		fmt.Fprintf(&b, "#%s\n\n", strings.Join(t.Tags, " #"))
	}
	fmt.Fprintf(&b, "Notes:\n%s\n", t.Notes)
	a.detail.SetText(b.String())
}

// Run builds the UI and starts the event loop.
func (a *App) Run() error {
	a.buildUI()
	a.refreshLists()
	a.refreshTasks()
	a.refreshDetail()
	a.bindKeys()
	return a.app.Run()
}

// bindKeys is implemented in Task 15; defined here so Run compiles.
func (a *App) bindKeys() {}

var _ = tcell.KeyTab
