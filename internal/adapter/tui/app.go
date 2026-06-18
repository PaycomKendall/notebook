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

	focusIdx int
	panes    []tview.Primitive
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

// focus sets the focused pane by index (0=lists,1=tasks,2=detail).
func (a *App) focus(i int) {
	if a.panes == nil {
		a.panes = []tview.Primitive{a.lists, a.tasks, a.detail}
	}
	if i < 0 || i >= len(a.panes) {
		return
	}
	a.focusIdx = i
	a.app.SetFocus(a.panes[i])
}

// cycleFocus moves focus by delta with wraparound.
func (a *App) cycleFocus(delta int) {
	n := 3
	a.focus(((a.focusIdx+delta)%n + n) % n)
}

// toggleSelected flips the done state of the selected task and refreshes.
func (a *App) toggleSelected() {
	t := a.selectedTask()
	if t == nil || a.current == nil {
		return
	}
	if err := a.svc.ToggleTask(a.current.Name, t.ID); err != nil {
		return
	}
	idx := a.tasks.GetCurrentItem()
	a.refreshTasks()
	if idx < a.tasks.GetItemCount() {
		a.tasks.SetCurrentItem(idx)
	}
	a.refreshDetail()
}

// deleteSelected removes the selected task after confirmation.
func (a *App) deleteSelected() {
	t := a.selectedTask()
	if t == nil || a.current == nil {
		return
	}
	name, id := a.current.Name, t.ID
	a.confirm(fmt.Sprintf("Delete #%d %q?", id, t.Title), func() {
		_ = a.svc.RemoveTask(name, id)
		a.refreshTasks()
		a.refreshDetail()
	})
}

// bindKeys installs global and pane key handlers.
func (a *App) bindKeys() {
	a.panes = []tview.Primitive{a.lists, a.tasks, a.detail}
	a.app.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyTab:
			a.cycleFocus(1)
			return nil
		case tcell.KeyBacktab:
			a.cycleFocus(-1)
			return nil
		case tcell.KeyCtrlC:
			a.app.Stop()
			return nil
		}
		switch ev.Rune() {
		case 'q':
			a.app.Stop()
			return nil
		case 'a':
			if a.focusIdx == 0 {
				a.newListForm()
			} else {
				a.addTaskForm()
			}
			return nil
		case 'd':
			if a.focusIdx == 1 {
				a.toggleSelected()
			}
			return nil
		case 'x':
			if a.focusIdx == 0 {
				a.deleteSelectedList()
			} else if a.focusIdx == 1 {
				a.deleteSelected()
			}
			return nil
		case 'r':
			if a.focusIdx == 0 {
				a.renameListForm()
			}
			return nil
		case 'e':
			if a.focusIdx == 1 {
				a.editTaskForm()
			}
			return nil
		case 'n':
			if a.focusIdx == 1 {
				a.editTaskForm()
			}
			return nil
		}
		return ev
	})
	a.focus(0)
}
