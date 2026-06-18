package bubbletui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the current mode.
func (m *Model) View() string {
	switch m.mode {
	case modeAddTask, modeEditTask, modeNewList, modeRenameList:
		return m.formView()
	case modeConfirm:
		return m.confirmView()
	default:
		return m.normalView()
	}
}

func (m *Model) paneWidths() (lists, tasks, detail int) {
	lists = 14
	avail := m.width - (lists + 4) - 8
	if avail < 36 {
		avail = 36
	}
	tasks = avail * 3 / 5
	detail = avail - tasks
	return
}

func (m *Model) paneHeight() int {
	h := m.height - 4
	if h < 4 {
		h = 4
	}
	return h
}

func (m *Model) normalView() string {
	row := lipgloss.JoinHorizontal(lipgloss.Top,
		m.renderLists(), m.renderTasks(), m.renderDetail())
	return row + "\n" + m.footer()
}

func (m *Model) renderLists() string {
	lw, _, _ := m.paneWidths()
	var b strings.Builder
	b.WriteString(m.styles.title.Render("Lists") + "\n\n")
	for i, name := range m.listNames {
		if i == m.listIdx {
			b.WriteString(m.styles.sel.Render("❯ "+name) + "\n")
		} else {
			b.WriteString("  " + name + "\n")
		}
	}
	style := m.styles.pane
	if m.focus == focusLists {
		style = m.styles.paneFocused
	}
	return style.Width(lw).Height(m.paneHeight()).Render(strings.TrimRight(b.String(), "\n"))
}

func (m *Model) renderTasks() string {
	_, tw, _ := m.paneWidths()
	title := "Tasks"
	if n := m.currentListName(); n != "" {
		title = "Tasks · " + n
	}
	var b strings.Builder
	b.WriteString(m.styles.title.Render(title) + "\n\n")
	if m.current != nil {
		for i, task := range m.current.Tasks {
			box := "[ ]"
			if task.Done {
				box = "[x]"
			}
			line := fmt.Sprintf("%s #%d %s", box, task.ID, task.Title)
			if len(task.Tags) > 0 {
				line += "  #" + strings.Join(task.Tags, " #")
			}
			if i == m.taskIdx {
				b.WriteString(m.styles.sel.Render("❯ "+line) + "\n")
			} else {
				b.WriteString("  " + line + "\n")
			}
		}
	}
	style := m.styles.pane
	if m.focus == focusTasks {
		style = m.styles.paneFocused
	}
	return style.Width(tw).Height(m.paneHeight()).Render(strings.TrimRight(b.String(), "\n"))
}

func (m *Model) renderDetail() string {
	_, _, dw := m.paneWidths()
	var b strings.Builder
	b.WriteString(m.styles.title.Render("Detail") + "\n\n")
	if t := m.selectedTask(); t != nil {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render(t.Title) + "\n")
		if len(t.Tags) > 0 {
			b.WriteString(m.styles.tag.Render("#"+strings.Join(t.Tags, " #")) + "\n")
		}
		b.WriteString("\n" + m.styles.dim.Render("Notes") + "\n" + t.Notes)
	}
	style := m.styles.pane
	if m.focus == focusDetail {
		style = m.styles.paneFocused
	}
	return style.Width(dw).Height(m.paneHeight()).Render(strings.TrimRight(b.String(), "\n"))
}

// hint builds a footer line from key/label pairs using the model's styles.
func (m *Model) hint(pairs [][2]string) string {
	var b strings.Builder
	b.WriteString(" ")
	for _, p := range pairs {
		b.WriteString(m.styles.key.Render(p[0]) + m.styles.dim.Render(" "+p[1]+"  "))
	}
	return strings.TrimRight(b.String(), " ")
}

func (m *Model) footer() string {
	if m.status != "" {
		return m.styles.warn.Render(" " + m.status)
	}
	switch m.focus {
	case focusLists:
		return m.hint([][2]string{{"tab", "pane"}, {"↑/↓", "move"}, {"a", "new"}, {"r", "rename"}, {"x", "delete"}, {"q", "quit"}})
	case focusTasks:
		return m.hint([][2]string{{"tab", "pane"}, {"↑/↓", "move"}, {"a", "add"}, {"d", "done"}, {"e", "edit"}, {"x", "delete"}, {"q", "quit"}})
	default:
		return m.hint([][2]string{{"tab", "pane"}, {"q", "quit"}})
	}
}
