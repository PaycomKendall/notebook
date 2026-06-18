package bubbletui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func newInput(label, value string) textinput.Model {
	ti := textinput.New()
	ti.Prompt = label + ": "
	ti.SetValue(value)
	ti.CharLimit = 200
	ti.Width = 36
	return ti
}

func (m *Model) refocusInputs() {
	for i := range m.inputs {
		if i == m.formField {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

func (m *Model) closeForm() {
	m.mode = modeNormal
	m.inputs = nil
	m.formField = 0
}

func (m *Model) openAddTask() {
	m.formList = m.currentListName()
	if m.formList == "" {
		m.formList = "inbox"
	}
	m.mode = modeAddTask
	m.inputs = []textinput.Model{
		newInput("Title", ""),
		newInput("Tags (space-separated)", ""),
		newInput("Notes", ""),
	}
	m.formField = 0
	m.refocusInputs()
}

func (m *Model) openEditTask() {
	t := m.selectedTask()
	if t == nil || m.current == nil {
		return
	}
	m.formList = m.current.Name
	m.formTaskID = t.ID
	m.mode = modeEditTask
	m.inputs = []textinput.Model{
		newInput("Title", t.Title),
		newInput("Notes", t.Notes),
	}
	m.formField = 0
	m.refocusInputs()
}

func (m *Model) openNewList() {
	m.mode = modeNewList
	m.inputs = []textinput.Model{newInput("List name", "")}
	m.formField = 0
	m.refocusInputs()
}

func (m *Model) openRenameList() {
	old := m.currentListName()
	if old == "" {
		return
	}
	m.formOldName = old
	m.mode = modeRenameList
	m.inputs = []textinput.Model{newInput("New name", old)}
	m.formField = 0
	m.refocusInputs()
}

func (m *Model) openMoveTask() {
	t := m.selectedTask()
	if t == nil || m.current == nil {
		return
	}
	m.formList = m.current.Name
	m.formTaskID = t.ID
	m.mode = modeMoveTask
	m.inputs = []textinput.Model{newInput("Move to list", "")}
	m.formField = 0
	m.refocusInputs()
}

// submitForm performs the Service call for the active form mode.
func (m *Model) submitForm() {
	m.status = ""
	switch m.mode {
	case modeAddTask:
		title := strings.TrimSpace(m.inputs[0].Value())
		if title == "" {
			return
		}
		tags := strings.Fields(m.inputs[1].Value())
		notes := m.inputs[2].Value()
		if _, err := m.svc.AddTask(m.formList, title, tags, notes); err != nil {
			m.status = err.Error()
		}
		m.closeForm()
		m.reloadLists()
		m.reloadTasks()
	case modeEditTask:
		title := strings.TrimSpace(m.inputs[0].Value())
		if title == "" {
			return
		}
		notes := m.inputs[1].Value()
		if err := m.svc.EditTask(m.formList, m.formTaskID, &title, &notes); err != nil {
			m.status = err.Error()
		}
		m.closeForm()
		m.reloadTasks()
	case modeNewList:
		name := strings.TrimSpace(m.inputs[0].Value())
		if name == "" {
			return
		}
		if err := m.svc.CreateList(name); err != nil {
			m.status = err.Error()
		}
		m.closeForm()
		m.reloadLists()
		m.reloadTasks()
	case modeRenameList:
		name := strings.TrimSpace(m.inputs[0].Value())
		if name == "" {
			return
		}
		if err := m.svc.RenameList(m.formOldName, name); err != nil {
			m.status = err.Error()
		}
		m.closeForm()
		m.reloadLists()
		m.reloadTasks()
	case modeMoveTask:
		dest := strings.TrimSpace(m.inputs[0].Value())
		if dest == "" {
			return
		}
		if _, err := m.svc.MoveTask(m.formList, m.formTaskID, dest); err != nil {
			m.status = err.Error()
		}
		m.closeForm()
		m.reloadLists()
		m.reloadTasks()
	}
}

// updateForm routes a key to the active form.
func (m *Model) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.closeForm()
		return m, nil
	case "enter":
		m.submitForm()
		return m, nil
	case "tab", "down":
		m.formField = (m.formField + 1) % len(m.inputs)
		m.refocusInputs()
		return m, nil
	case "shift+tab", "up":
		m.formField = (m.formField - 1 + len(m.inputs)) % len(m.inputs)
		m.refocusInputs()
		return m, nil
	}
	var cmd tea.Cmd
	m.inputs[m.formField], cmd = m.inputs[m.formField].Update(msg)
	return m, cmd
}

func (m *Model) confirmDeleteTask() {
	t := m.selectedTask()
	if t == nil || m.current == nil {
		return
	}
	name, id := m.current.Name, t.ID
	m.confirmPrompt = fmt.Sprintf("Delete #%d %q?", id, t.Title)
	m.confirmAction = func() error { return m.svc.RemoveTask(name, id) }
	m.mode = modeConfirm
}

func (m *Model) confirmDeleteList() {
	name := m.currentListName()
	if name == "" {
		return
	}
	m.confirmPrompt = fmt.Sprintf("Delete list %q and all its tasks?", name)
	m.confirmAction = func() error { return m.svc.DeleteList(name) }
	m.mode = modeConfirm
}

// updateConfirm handles the yes/no modal.
func (m *Model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		if m.confirmAction != nil {
			if err := m.confirmAction(); err != nil {
				m.status = err.Error()
			}
		}
		m.mode = modeNormal
		m.confirmAction = nil
		m.reloadLists()
		m.reloadTasks()
	case "n", "esc":
		m.mode = modeNormal
		m.confirmAction = nil
	}
	return m, nil
}

func (m *Model) formTitle() string {
	switch m.mode {
	case modeAddTask:
		return "Add task"
	case modeEditTask:
		return "Edit task"
	case modeNewList:
		return "New list"
	case modeRenameList:
		return "Rename list"
	case modeMoveTask:
		return "Move task"
	}
	return ""
}

func (m *Model) formView() string {
	var b strings.Builder
	b.WriteString(m.styles.title.Render(m.formTitle()) + "\n\n")
	for i := range m.inputs {
		b.WriteString(m.inputs[i].View() + "\n")
	}
	b.WriteString("\n" + m.styles.dim.Render("tab/↑↓: move · enter: submit · esc: cancel"))
	if m.status != "" {
		b.WriteString("\n" + m.styles.warn.Render(m.status))
	}
	return m.styles.modal.Render(b.String())
}

func (m *Model) confirmView() string {
	body := m.styles.title.Render("Confirm") + "\n\n" + m.confirmPrompt + "\n\n" +
		m.styles.dim.Render("y/enter: yes · n/esc: no")
	return m.styles.modal.Render(body)
}
