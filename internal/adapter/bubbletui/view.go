package bubbletui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the current mode.
func (m *Model) View() string {
	switch m.mode {
	case modeAddTask, modeEditTask, modeNewList, modeRenameList, modeMoveTask:
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
	focused := m.focus == focusLists
	var b strings.Builder
	b.WriteString(m.titleFor("Lists", focused) + "\n\n")
	for i, name := range m.listNames {
		if i == m.listIdx {
			b.WriteString(m.selRow("❯ "+name, focused) + "\n")
		} else {
			b.WriteString("  " + name + "\n")
		}
	}
	return m.paneStyle(focused).Width(lw).Height(m.paneHeight()).Render(strings.TrimRight(b.String(), "\n"))
}

func (m *Model) renderTasks() string {
	_, tw, _ := m.paneWidths()
	focused := m.focus == focusTasks
	title := "Tasks"
	if n := m.currentListName(); n != "" {
		title = "Tasks · " + n
	}
	var b strings.Builder
	b.WriteString(m.titleFor(title, focused) + "\n\n")
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
				b.WriteString(m.selRow("❯ "+line, focused) + "\n")
			} else {
				b.WriteString("  " + line + "\n")
			}
		}
	}
	return m.paneStyle(focused).Width(tw).Height(m.paneHeight()).Render(strings.TrimRight(b.String(), "\n"))
}

func (m *Model) renderDetail() string {
	_, _, dw := m.paneWidths()
	focused := m.focus == focusDetail
	var b strings.Builder
	b.WriteString(m.titleFor("Detail", focused) + "\n\n")
	if t := m.selectedTask(); t != nil {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render(t.Title) + "\n")
		if len(t.Tags) > 0 {
			b.WriteString(m.styles.tag.Render("#"+strings.Join(t.Tags, " #")) + "\n")
		}
		b.WriteString("\n" + m.styles.dim.Render("Notes") + "\n" + t.Notes)
	}
	return m.paneStyle(focused).Width(dw).Height(m.paneHeight()).Render(strings.TrimRight(b.String(), "\n"))
}

// titleFor renders a pane title as a filled chip when the pane is focused.
func (m *Model) titleFor(s string, focused bool) string {
	if focused {
		return m.styles.titleFocused.Render(s)
	}
	return m.styles.title.Render(s)
}

// selRow renders the selected row bright when focused, dimmed otherwise.
func (m *Model) selRow(s string, focused bool) string {
	if focused {
		return m.styles.sel.Render(s)
	}
	return m.styles.selDim.Render(s)
}

// paneStyle picks the thick focused border or the plain rounded one.
func (m *Model) paneStyle(focused bool) lipgloss.Style {
	if focused {
		return m.styles.paneFocused
	}
	return m.styles.pane
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
		return m.hint([][2]string{{"tab", "pane"}, {"↑/↓", "move"}, {"a", "add"}, {"d", "done"}, {"e", "edit"}, {"m", "→list"}, {"x", "delete"}, {"q", "quit"}})
	default:
		return m.hint([][2]string{{"tab", "pane"}, {"q", "quit"}})
	}
}
