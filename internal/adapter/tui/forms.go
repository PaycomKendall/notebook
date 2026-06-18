package tui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

// confirm shows a yes/no modal, running onYes if confirmed.
func (a *App) confirm(prompt string, onYes func()) {
	prev := a.app.GetFocus()
	modal := tview.NewModal().
		SetText(prompt).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(_ int, label string) {
			a.rebuildRoot()
			if label == "Yes" {
				onYes()
			}
			a.app.SetFocus(prev)
		})
	a.app.SetRoot(modal, true)
}

// rebuildRoot restores the main three-pane flex as the root.
func (a *App) rebuildRoot() {
	flex := tview.NewFlex().
		AddItem(a.lists, 24, 0, true).
		AddItem(a.tasks, 0, 2, false).
		AddItem(a.detail, 0, 2, false)
	a.app.SetRoot(flex, true)
	a.focus(a.focusIdx)
}

// addTaskForm collects a new task and adds it to the current list.
func (a *App) addTaskForm() {
	list := "inbox"
	if a.current != nil {
		list = a.current.Name
	} else if n := a.selectedListName(); n != "" {
		list = n
	}
	var title, tags, notes string
	form := tview.NewForm().
		AddInputField("Title", "", 40, nil, func(s string) { title = s }).
		AddInputField("Tags (space-sep)", "", 40, nil, func(s string) { tags = s }).
		AddInputField("Notes", "", 40, nil, func(s string) { notes = s })
	form.AddButton("Add", func() {
		if strings.TrimSpace(title) != "" {
			_, _ = a.svc.AddTask(list, title, strings.Fields(tags), notes)
		}
		a.rebuildRoot()
		a.refreshLists()
		a.refreshTasks()
		a.refreshDetail()
	})
	form.AddButton("Cancel", func() { a.rebuildRoot() })
	form.SetBorder(true).SetTitle(" Add task ")
	a.app.SetRoot(form, true)
}

// editTaskForm edits the selected task's title and notes.
func (a *App) editTaskForm() {
	t := a.selectedTask()
	if t == nil || a.current == nil {
		return
	}
	name, id := a.current.Name, t.ID
	title, notes := t.Title, t.Notes
	form := tview.NewForm().
		AddInputField("Title", title, 40, nil, func(s string) { title = s }).
		AddInputField("Notes", notes, 40, nil, func(s string) { notes = s })
	form.AddButton("Save", func() {
		tp, np := title, notes
		_ = a.svc.EditTask(name, id, &tp, &np)
		a.rebuildRoot()
		a.refreshTasks()
		a.refreshDetail()
	})
	form.AddButton("Cancel", func() { a.rebuildRoot() })
	form.SetBorder(true).SetTitle(" Edit task ")
	a.app.SetRoot(form, true)
}

// newListForm creates a new list.
func (a *App) newListForm() {
	var name string
	form := tview.NewForm().
		AddInputField("List name", "", 30, nil, func(s string) { name = s })
	form.AddButton("Create", func() {
		if strings.TrimSpace(name) != "" {
			_ = a.svc.CreateList(name)
		}
		a.rebuildRoot()
		a.refreshLists()
	})
	form.AddButton("Cancel", func() { a.rebuildRoot() })
	form.SetBorder(true).SetTitle(" New list ")
	a.app.SetRoot(form, true)
}

// deleteSelectedList removes the highlighted list after confirmation.
func (a *App) deleteSelectedList() {
	name := a.selectedListName()
	if name == "" {
		return
	}
	a.confirm(fmt.Sprintf("Delete list %q and all its tasks?", name), func() {
		_ = a.svc.DeleteList(name)
		a.refreshLists()
		a.refreshTasks()
		a.refreshDetail()
	})
}

// renameListForm renames the highlighted list.
func (a *App) renameListForm() {
	old := a.selectedListName()
	if old == "" {
		return
	}
	name := old
	form := tview.NewForm().
		AddInputField("New name", old, 30, nil, func(s string) { name = s })
	form.AddButton("Rename", func() {
		if strings.TrimSpace(name) != "" && name != old {
			_ = a.svc.RenameList(old, name)
		}
		a.rebuildRoot()
		a.refreshLists()
		a.refreshTasks()
		a.refreshDetail()
	})
	form.AddButton("Cancel", func() { a.rebuildRoot() })
	form.SetBorder(true).SetTitle(" Rename list ")
	a.app.SetRoot(form, true)
}
