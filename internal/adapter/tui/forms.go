package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// formHint is the keybinding hint shown in the footer of every input form.
const formHint = " Tab/↑↓: move  ·  Enter: newline in Notes  ·  Esc: cancel  (use buttons to submit)"

// showModalForm displays a bordered form with a hint footer, wiring Esc to
// cancel (restore the main view). All input forms go through this helper.
func (a *App) showModalForm(form *tview.Form, title string) {
	a.lastForm = form
	form.SetBorder(true).SetTitle(title)
	form.SetCancelFunc(func() { a.rebuildRoot() })
	hint := tview.NewTextView().SetText(formHint)
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(hint, 1, 0, false)
	a.app.SetRoot(layout, true)
}

// confirm shows a yes/no modal, running onYes if confirmed. Esc cancels.
func (a *App) confirm(prompt string, onYes func()) {
	prev := a.app.GetFocus()
	cancel := func() {
		a.rebuildRoot()
		a.app.SetFocus(prev)
	}
	modal := tview.NewModal().
		SetText(prompt + "\n\nEnter: select  ·  Esc: cancel").
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(_ int, label string) {
			a.rebuildRoot()
			if label == "Yes" {
				onYes()
			}
			a.app.SetFocus(prev)
		})
	modal.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEscape {
			cancel()
			return nil
		}
		return ev
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
		AddTextArea("Notes", "", 40, 6, 0, func(s string) { notes = s })
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
	a.showModalForm(form, " Add page ")
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
		AddTextArea("Notes", notes, 40, 6, 0, func(s string) { notes = s })
	form.AddButton("Save", func() {
		tp, np := title, notes
		_ = a.svc.EditTask(name, id, &tp, &np)
		a.rebuildRoot()
		a.refreshTasks()
		a.refreshDetail()
	})
	form.AddButton("Cancel", func() { a.rebuildRoot() })
	a.showModalForm(form, " Edit page ")
}

// newListForm creates a new list.
func (a *App) newListForm() {
	var name string
	form := tview.NewForm().
		AddInputField("Folder name", "", 30, nil, func(s string) { name = s })
	form.AddButton("Create", func() {
		if strings.TrimSpace(name) != "" {
			_ = a.svc.CreateList(name)
		}
		a.rebuildRoot()
		a.refreshLists()
	})
	form.AddButton("Cancel", func() { a.rebuildRoot() })
	a.showModalForm(form, " New folder ")
}

// deleteSelectedList removes the highlighted list after confirmation.
func (a *App) deleteSelectedList() {
	name := a.selectedListName()
	if name == "" {
		return
	}
	a.confirm(fmt.Sprintf("Delete folder %q and all its pages?", name), func() {
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
	a.showModalForm(form, " Rename folder ")
}

// moveTaskForm moves the selected task to another list (auto-created if new).
func (a *App) moveTaskForm() {
	t := a.selectedTask()
	if t == nil || a.current == nil {
		return
	}
	src, id := a.current.Name, t.ID
	var dest string
	form := tview.NewForm().
		AddInputField("Move to folder", "", 30, nil, func(s string) { dest = s })
	form.AddButton("Move", func() {
		if strings.TrimSpace(dest) != "" {
			_, _ = a.svc.MoveTask(src, id, dest)
		}
		a.rebuildRoot()
		a.refreshLists()
		a.refreshTasks()
		a.refreshDetail()
	})
	form.AddButton("Cancel", func() { a.rebuildRoot() })
	a.showModalForm(form, " Move page ")
}
