package bubbletui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// formField is a single field in a modal form. It abstracts over single-line
// (textinput) and multi-line (textarea) widgets so the form loop stays generic.
type formField interface {
	Focus() tea.Cmd
	Blur()
	Update(tea.Msg) tea.Cmd
	View() string
	Value() string
	SetValue(string)
	label() string // a heading printed above the widget; "" if the widget shows its own
	multiline() bool
}

// lineField wraps a single-line textinput. Its label lives in the input's
// Prompt, so label() returns "".
type lineField struct{ ti textinput.Model }

func (f *lineField) Focus() tea.Cmd { return f.ti.Focus() }
func (f *lineField) Blur()          { f.ti.Blur() }
func (f *lineField) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	f.ti, cmd = f.ti.Update(msg)
	return cmd
}
func (f *lineField) View() string      { return f.ti.View() }
func (f *lineField) Value() string     { return f.ti.Value() }
func (f *lineField) SetValue(s string) { f.ti.SetValue(s) }
func (f *lineField) label() string     { return "" }
func (f *lineField) multiline() bool   { return false }

func newInput(label, value string) formField {
	ti := textinput.New()
	ti.Prompt = label + ": "
	ti.SetValue(value)
	ti.CharLimit = 200
	ti.Width = 36
	return &lineField{ti: ti}
}

// areaField wraps a multi-line textarea. Its label is printed above it by
// formView (textarea has no label/Prompt concept like textinput).
type areaField struct {
	ta  textarea.Model
	lbl string
}

func (f *areaField) Focus() tea.Cmd { return f.ta.Focus() }
func (f *areaField) Blur()          { f.ta.Blur() }
func (f *areaField) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	f.ta, cmd = f.ta.Update(msg)
	return cmd
}
func (f *areaField) View() string      { return f.ta.View() }
func (f *areaField) Value() string     { return f.ta.Value() }
func (f *areaField) SetValue(s string) { f.ta.SetValue(s) }
func (f *areaField) label() string     { return f.lbl }
func (f *areaField) multiline() bool   { return true }

func newArea(label, value string) formField {
	ta := textarea.New()
	ta.ShowLineNumbers = false
	ta.CharLimit = 2000
	ta.SetWidth(38)
	ta.SetHeight(5)
	ta.SetValue(value)
	return &areaField{ta: ta, lbl: label}
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
	m.inputs = []formField{
		newInput("Title", ""),
		newInput("Tags (space-separated)", ""),
		newArea("Notes", ""),
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
	m.inputs = []formField{
		newInput("Title", t.Title),
		newArea("Notes", t.Notes),
	}
	m.formField = 0
	m.refocusInputs()
}

func (m *Model) openNewList() {
	m.mode = modeNewList
	m.inputs = []formField{newInput("Folder name", "")}
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
	m.inputs = []formField{newInput("New name", old)}
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
	m.inputs = []formField{newInput("Move to folder", "")}
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
	cur := m.inputs[m.formField]
	switch msg.Type {
	case tea.KeyEsc:
		m.closeForm()
		return m, nil
	case tea.KeyCtrlS:
		m.submitForm()
		return m, nil
	case tea.KeyEnter:
		if !cur.multiline() {
			m.submitForm()
			return m, nil
		}
		// multiline: fall through to delegate (inserts a newline)
	case tea.KeyTab:
		m.formField = (m.formField + 1) % len(m.inputs)
		m.refocusInputs()
		return m, nil
	case tea.KeyShiftTab:
		m.formField = (m.formField - 1 + len(m.inputs)) % len(m.inputs)
		m.refocusInputs()
		return m, nil
	case tea.KeyDown, tea.KeyUp:
		if !cur.multiline() {
			step := 1
			if msg.Type == tea.KeyUp {
				step = -1
			}
			m.formField = (m.formField + step + len(m.inputs)) % len(m.inputs)
			m.refocusInputs()
			return m, nil
		}
		// multiline: fall through to delegate (cursor moves between lines)
	}
	cmd := cur.Update(msg)
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
	m.confirmPrompt = fmt.Sprintf("Delete folder %q and all its pages?", name)
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
		return "Add page"
	case modeEditTask:
		return "Edit page"
	case modeNewList:
		return "New folder"
	case modeRenameList:
		return "Rename folder"
	case modeMoveTask:
		return "Move page"
	}
	return ""
}

func (m *Model) formView() string {
	var b strings.Builder
	b.WriteString(m.styles.title.Render(m.formTitle()) + "\n\n")
	for i := range m.inputs {
		if lbl := m.inputs[i].label(); lbl != "" {
			b.WriteString(m.styles.dim.Render(lbl) + "\n")
		}
		b.WriteString(m.inputs[i].View() + "\n")
	}
	b.WriteString("\n" + m.styles.dim.Render("tab: move · ctrl+s: submit · esc: cancel"))
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
