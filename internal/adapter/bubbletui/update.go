package bubbletui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Update routes messages: window sizing, then keys by mode.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		switch m.mode {
		case modeConfirm:
			return m.updateConfirm(msg)
		case modeNormal:
			return m.updateNormal(msg)
		default:
			return m.updateForm(msg)
		}
	}
	return m, nil
}

func (m *Model) cycleFocus(delta int) {
	n := 3
	m.focus = focusPane(((int(m.focus)+delta)%n + n) % n)
}

func (m *Model) moveSelection(delta int) {
	switch m.focus {
	case focusLists:
		m.listIdx += delta
		if m.listIdx < 0 {
			m.listIdx = 0
		}
		if m.listIdx >= len(m.listNames) {
			m.listIdx = len(m.listNames) - 1
		}
		m.taskIdx = 0
		m.reloadTasks()
	case focusTasks:
		if m.current == nil {
			return
		}
		m.taskIdx += delta
		if m.taskIdx < 0 {
			m.taskIdx = 0
		}
		if m.taskIdx >= len(m.current.Tasks) {
			m.taskIdx = len(m.current.Tasks) - 1
		}
	}
}

func (m *Model) toggleSelected() {
	t := m.selectedTask()
	if t == nil || m.current == nil {
		return
	}
	if err := m.svc.ToggleTask(m.current.Name, t.ID); err != nil {
		m.status = err.Error()
		return
	}
	m.reloadTasks()
}

func (m *Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.status = "" // clear stale status on any normal-mode key
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "tab":
		m.cycleFocus(1)
	case "shift+tab":
		m.cycleFocus(-1)
	case "up", "k":
		m.moveSelection(-1)
	case "down", "j":
		m.moveSelection(1)
	case "a":
		if m.focus == focusLists {
			m.openNewList()
		} else {
			m.openAddTask()
		}
		return m, textinput.Blink
	case "d":
		if m.focus == focusTasks {
			m.toggleSelected()
		}
	case "e", "n":
		if m.focus == focusTasks {
			m.openEditTask()
			return m, textinput.Blink
		}
	case "m":
		if m.focus == focusTasks {
			m.openMoveTask()
			return m, textinput.Blink
		}
	case "r":
		if m.focus == focusLists {
			m.openRenameList()
			return m, textinput.Blink
		}
	case "x":
		if m.focus == focusTasks {
			m.confirmDeleteTask()
		} else if m.focus == focusLists {
			m.confirmDeleteList()
		}
	}
	return m, nil
}
