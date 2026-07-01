package bubbletui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kendallowen/notebook/internal/markdown"
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
	tasks = avail * 2 / 5
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
	content := row + "\n" + m.footer()
	// Paper themes flood the whole terminal (footer + margins below/right of the
	// panes) with the page background so the notebook look runs edge to edge.
	// Safe because every inner style also carries the page bg, so any style reset
	// only ever exposes paper-on-paper. Terminal-native themes are unchanged.
	if m.theme.bg != nil && m.width > 0 && m.height > 0 {
		return m.styles.page.Width(m.width).Height(m.height).Render(content)
	}
	return content
}

func (m *Model) renderLists() string {
	lw, _, _ := m.paneWidths()
	focused := m.focus == focusLists
	var b strings.Builder
	b.WriteString(m.titleFor("Folders", focused) + "\n\n")
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
	title := "Pages"
	if n := m.currentListName(); n != "" {
		title = "Pages · " + n
	}
	var b strings.Builder
	b.WriteString(m.titleFor(title, focused) + "\n\n")
	if m.current != nil {
		for i, task := range m.current.Tasks {
			box := "[ ]"
			if task.Done {
				box = "[x]"
			}
			line := fmt.Sprintf("%s %s", box, task.Title)
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

const (
	ndGutter = "◦ " // spiral binding hole + space
	ndMargin = "│ " // margin rule + space
)

func (m *Model) renderDetail() string {
	_, _, dw := m.paneWidths()
	focused := m.focus == focusDetail
	// inner content width: pane width minus paneStyle's horizontal padding (1+1)
	// minus the gutter and margin prefixes.
	contentW := dw - 2 - lipgloss.Width(ndGutter) - lipgloss.Width(ndMargin)
	if contentW < 8 {
		contentW = 8
	}

	title := ""
	var lines []string
	if t := m.selectedTask(); t != nil {
		title = t.Title
		if len(t.Tags) > 0 {
			lines = append(lines, m.styles.tag.Render("#"+strings.Join(t.Tags, " #")))
		}
		lines = append(lines, "")
		lines = append(lines, strings.Split(markdown.Render(t.Notes, contentW, m.mdStyles()), "\n")...)
	}

	body := m.titleFor("Detail", focused) + "\n\n" + m.notebookPage(title, lines, contentW)
	return m.paneStyle(focused).Width(dw).Height(m.paneHeight()).Render(strings.TrimRight(body, "\n"))
}

// notebookPage decorates content lines as a notebook page: a header band showing
// the task title, a separator rule, then guttered + margined + ruled rows
// filling the pane height.
func (m *Model) notebookPage(title string, lines []string, contentW int) string {
	gutter := m.styles.dim.Render(ndGutter)
	margin := m.styles.tag.Render(ndMargin)
	rule := lipgloss.NewStyle().Underline(true).Foreground(m.theme.subtle)
	if m.theme.bg != nil {
		rule = rule.Background(m.theme.bg)
	}

	if title == "" {
		title = "Notebook"
	}
	var b strings.Builder
	// Pad the header band to contentW like the ruled rows below it: the title is
	// pre-styled (bold), so append separately-styled padding rather than letting
	// the pane's block padding fill the gap — on paper themes that block padding
	// arrives without the page background and punches a hole in the paper.
	head := m.styles.title.Render(title)
	if pad := contentW - lipgloss.Width(title); pad > 0 {
		head += m.styles.dim.Render(strings.Repeat(" ", pad))
	}
	b.WriteString(gutter + margin + head + "\n")
	b.WriteString(gutter + margin + m.styles.dim.Render(strings.Repeat("─", contentW)) + "\n")

	// rows = pane height minus the Detail title (1), the blank line (1),
	// the header band (1) and the separator (1).
	rows := m.paneHeight() - 4
	if rows < 1 {
		rows = 1
	}
	for i := 0; i < rows; i++ {
		text := ""
		if i < len(lines) {
			text = lines[i]
		}
		b.WriteString(gutter + margin + ruledLine(text, contentW, rule) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// ruledLine renders one notebook row, padded to contentW and underlined to look
// like a ruled line. If text already contains ANSI styling (the bold title, a
// tag, or an inline bold/code span), only the trailing padding is underlined:
// wrapping already-styled text in another lipgloss style corrupts the embedded
// escapes — lipgloss emits the ESC byte bare, so the terminal prints the rest of
// the sequence ("[1m") as literal text. Plain rows are underlined whole.
func ruledLine(text string, contentW int, rule lipgloss.Style) string {
	pad := contentW - lipgloss.Width(text)
	if pad < 0 {
		pad = 0
	}
	spaces := strings.Repeat(" ", pad)
	if strings.ContainsRune(text, '\x1b') {
		return text + rule.Render(spaces)
	}
	return rule.Render(text + spaces)
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
		return m.hint([][2]string{{"tab", "pane"}, {"↑/↓", "move"}, {"a", "add"}, {"d", "done"}, {"e", "edit"}, {"m", "→folder"}, {"x", "delete"}, {"q", "quit"}})
	default:
		return m.hint([][2]string{{"tab", "pane"}, {"q", "quit"}})
	}
}
