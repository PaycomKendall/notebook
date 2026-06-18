# Bubble Tea TUI Adapter Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a second, full-parity TUI front-end built with Bubble Tea/Lipgloss/Bubbles, selectable at runtime (`nb tui --engine bubble`), with tview remaining the default.

**Architecture:** A new `internal/adapter/bubbletui` package implements the Elm architecture (`Model`/`Update`/`View`) over the same `internal/todo.Service`. Only `cmd/nb/main.go` imports adapters; the `tui` command resolves an engine string and the composition root dispatches.

**Tech Stack:** Go 1.24, `charmbracelet/bubbletea`, `charmbracelet/lipgloss`, `charmbracelet/bubbles` (textinput), Cobra (existing).

## Global Constraints

- **Module path:** `github.com/kendallowen/notebook`.
- **Hexagonal rule:** `internal/adapter/bubbletui` imports only `internal/todo` + Charm libs (no cobra/tview/jsonstore). Only `cmd/nb/main.go` imports adapters.
- **Engine selection:** `tui` command flag `-e/--engine`; resolution flag → `$NB_TUI` → `"tview"`; invalid value → error `invalid engine "%s" (want "tview" or "bubble")`.
- **`launchTUI` callback signature:** `func(engine string) error` (was `func() error`).
- **Default engine:** `tview` (unchanged behavior for `nb tui`).
- **Parity:** 3-pane nav, task toggle/add/edit/delete (tags+notes), list create/rename/delete, modal forms with Esc-cancel, context footer. All mutations go through the `Service`.
- **Form UX (Bubble Tea):** field-only forms — `Tab`/`↑↓` move between fields, `Enter` submits, `Esc` cancels (no Save/Cancel buttons). Validation failure (empty required field) keeps the form open; a Service error closes it and shows the message in `status`.
- **Detail pane:** static text (no scrolling) — a deliberate simplification.
- **TDD:** test-first; commit after each task. Run `gofmt -l` before each commit.

**Prerequisite:** Go 1.24 already installed; the tview adapter, cli, jsonstore, and domain are complete on `main`.

---

### Task 1: Dependencies + Model + data layer

**Files:**
- Create: `internal/adapter/bubbletui/model.go`
- Test: `internal/adapter/bubbletui/model_test.go`

**Interfaces:**
- Consumes: `todo.Service` (`ListNames`, `GetList`).
- Produces: `type Model`, `focusPane`/`mode` enums and constants; `func New(svc *todo.Service) *Model`; `func Run(svc *todo.Service) error`; methods `reloadLists()`, `reloadTasks()`, `currentListName() string`, `selectedTask() *todo.Task`, `Init() tea.Cmd`, `Update(tea.Msg) (tea.Model, tea.Cmd)` (stub), `View() string` (stub). Fields used by later tasks: `svc, listNames, listIdx, current, taskIdx, focus, mode, inputs, formField, formList, formTaskID, formOldName, confirmPrompt, confirmAction, status, width, height`.

- [ ] **Step 1: Add dependencies**

Run:
```bash
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/bubbles@latest
```
Expected: `go.mod`/`go.sum` updated.

- [ ] **Step 2: Write the failing test**

`internal/adapter/bubbletui/model_test.go`:
```go
package bubbletui

import (
	"testing"

	"github.com/kendallowen/notebook/internal/adapter/jsonstore"
	"github.com/kendallowen/notebook/internal/todo"
)

// newTestModel builds a Model over a temp-dir store, optionally seeded.
func newTestModel(t *testing.T, seed func(*todo.Service)) (*Model, *todo.Service) {
	t.Helper()
	store, err := jsonstore.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	svc := todo.NewService(store)
	if seed != nil {
		seed(svc)
	}
	return New(svc), svc
}

func TestNewLoadsListsAndTasks(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) {
		_, _ = s.AddTask("work", "alpha", nil, "")
		_, _ = s.AddTask("work", "beta", nil, "")
	})
	if len(m.listNames) != 1 || m.listNames[0] != "work" {
		t.Fatalf("listNames = %v, want [work]", m.listNames)
	}
	if m.current == nil || len(m.current.Tasks) != 2 {
		t.Fatalf("current tasks not loaded: %+v", m.current)
	}
	if got := m.selectedTask(); got == nil || got.Title != "alpha" {
		t.Errorf("selectedTask = %+v, want alpha", got)
	}
}

func TestReloadTasksClampsIndex(t *testing.T) {
	m, svc := newTestModel(t, func(s *todo.Service) {
		_, _ = s.AddTask("work", "alpha", nil, "")
		_, _ = s.AddTask("work", "beta", nil, "")
	})
	m.taskIdx = 1
	if err := svc.RemoveTask("work", 2); err != nil {
		t.Fatal(err)
	}
	m.reloadTasks()
	if m.taskIdx != 0 {
		t.Errorf("taskIdx = %d, want clamped to 0", m.taskIdx)
	}
	if m.selectedTask() == nil {
		t.Error("selectedTask should be valid after clamp")
	}
}
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `go test ./internal/adapter/bubbletui/`
Expected: FAIL — package/symbols undefined.

- [ ] **Step 4: Implement `model.go`**

`internal/adapter/bubbletui/model.go`:
```go
package bubbletui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kendallowen/notebook/internal/todo"
)

type focusPane int

const (
	focusLists focusPane = iota
	focusTasks
	focusDetail
)

type mode int

const (
	modeNormal mode = iota
	modeAddTask
	modeEditTask
	modeNewList
	modeRenameList
	modeConfirm
)

// Model is the Bubble Tea front-end over the shared Service.
type Model struct {
	svc *todo.Service

	listNames []string
	listIdx   int
	current   *todo.List
	taskIdx   int

	focus focusPane
	mode  mode

	inputs      []textinput.Model
	formField   int
	formList    string
	formTaskID  int
	formOldName string

	confirmPrompt string
	confirmAction func() error

	status        string
	width, height int
}

// New builds a Model and loads the initial lists + tasks.
func New(svc *todo.Service) *Model {
	m := &Model{svc: svc, width: 90, height: 24}
	m.reloadLists()
	m.reloadTasks()
	return m
}

func (m *Model) reloadLists() {
	names, err := m.svc.ListNames()
	if err != nil {
		m.status = err.Error()
		return
	}
	m.listNames = names
	if m.listIdx >= len(names) {
		m.listIdx = len(names) - 1
	}
	if m.listIdx < 0 {
		m.listIdx = 0
	}
}

func (m *Model) currentListName() string {
	if m.listIdx < 0 || m.listIdx >= len(m.listNames) {
		return ""
	}
	return m.listNames[m.listIdx]
}

func (m *Model) reloadTasks() {
	name := m.currentListName()
	if name == "" {
		m.current = nil
		m.taskIdx = 0
		return
	}
	l, err := m.svc.GetList(name)
	if err != nil {
		m.current = nil
		m.taskIdx = 0
		return
	}
	m.current = l
	if len(l.Tasks) == 0 {
		m.taskIdx = 0
		return
	}
	if m.taskIdx >= len(l.Tasks) {
		m.taskIdx = len(l.Tasks) - 1
	}
	if m.taskIdx < 0 {
		m.taskIdx = 0
	}
}

func (m *Model) selectedTask() *todo.Task {
	if m.current == nil || m.taskIdx < 0 || m.taskIdx >= len(m.current.Tasks) {
		return nil
	}
	return &m.current.Tasks[m.taskIdx]
}

// Init satisfies tea.Model; Bubble Tea sends the initial WindowSizeMsg.
func (m *Model) Init() tea.Cmd { return nil }

// Update is a stub until Task 4.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }

// View is a stub until Task 2.
func (m *Model) View() string { return "" }

// Run starts the Bubble Tea program in the alternate screen.
func Run(svc *todo.Service) error {
	p := tea.NewProgram(New(svc), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./internal/adapter/bubbletui/`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
gofmt -l internal/adapter/bubbletui/
git add go.mod go.sum internal/adapter/bubbletui/
git commit -m "feat(bubbletui): deps, Model, data layer"
```

---

### Task 2: Styles + normal-mode View

**Files:**
- Create: `internal/adapter/bubbletui/styles.go`
- Create: `internal/adapter/bubbletui/view.go`
- Modify: `internal/adapter/bubbletui/model.go` (remove the stub `View`)
- Test: `internal/adapter/bubbletui/view_test.go`

**Interfaces:**
- Produces: Lipgloss styles in `styles.go`; `func (m *Model) View() string` (real, normal-mode); render helpers `renderLists`, `renderTasks`, `renderDetail`, `footer`, `paneWidths`, `paneHeight`, `hint`.

- [ ] **Step 1: Write the failing test**

`internal/adapter/bubbletui/view_test.go`:
```go
package bubbletui

import (
	"strings"
	"testing"

	"github.com/kendallowen/notebook/internal/todo"
)

func TestNormalViewShowsPanesAndFooter(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) {
		_, _ = s.AddTask("work", "alpha", []string{"hr"}, "a note")
		_, _ = s.AddTask("work", "beta", nil, "")
	})
	m.focus = focusTasks
	out := m.View()

	for _, want := range []string{"Lists", "work", "Tasks", "alpha", "beta", "[ ]", "Detail", "a note"} {
		if !strings.Contains(out, want) {
			t.Errorf("View missing %q\n%s", want, out)
		}
	}
	// selected task marker on the focused row
	if !strings.Contains(out, "❯") {
		t.Errorf("View missing selection marker\n%s", out)
	}
	// tasks-pane footer hints
	for _, want := range []string{"add", "done", "edit", "delete", "quit"} {
		if !strings.Contains(out, want) {
			t.Errorf("footer missing %q\n%s", want, out)
		}
	}
}

func TestListsFooterDiffersFromTasks(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _, _ = s.AddTask("work", "alpha", nil, "") })
	m.focus = focusLists
	out := m.View()
	for _, want := range []string{"new", "rename"} {
		if !strings.Contains(out, want) {
			t.Errorf("lists footer missing %q\n%s", want, out)
		}
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/adapter/bubbletui/`
Expected: FAIL — View returns "" (missing substrings) / duplicate `View`.

- [ ] **Step 3: Remove the stub `View` from `model.go`**

Delete these two lines from `internal/adapter/bubbletui/model.go`:
```go
// View is a stub until Task 2.
func (m *Model) View() string { return "" }
```

- [ ] **Step 4: Create `styles.go`**

`internal/adapter/bubbletui/styles.go`:
```go
package bubbletui

import "github.com/charmbracelet/lipgloss"

var (
	accent = lipgloss.Color("212")
	mauve  = lipgloss.Color("99")
	subtle = lipgloss.Color("245")
	selBg  = lipgloss.Color("57")
	white  = lipgloss.Color("231")
	warnFg = lipgloss.Color("203")

	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(accent)
	dimStyle   = lipgloss.NewStyle().Foreground(subtle)
	keyStyle   = lipgloss.NewStyle().Bold(true).Foreground(mauve)
	selStyle   = lipgloss.NewStyle().Foreground(white).Background(selBg).Bold(true)
	tagStyle   = lipgloss.NewStyle().Foreground(accent)
	warnStyle  = lipgloss.NewStyle().Foreground(warnFg)

	paneStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(mauve).Padding(0, 1)
	paneFocused  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent).Padding(0, 1)
	modalStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent).Padding(1, 2)
)
```

- [ ] **Step 5: Create `view.go`**

`internal/adapter/bubbletui/view.go`:
```go
package bubbletui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the current mode. Form/confirm modes are added in Task 3.
func (m *Model) View() string {
	return m.normalView()
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
	b.WriteString(titleStyle.Render("Lists") + "\n\n")
	for i, name := range m.listNames {
		if i == m.listIdx {
			b.WriteString(selStyle.Render("❯ "+name) + "\n")
		} else {
			b.WriteString("  " + name + "\n")
		}
	}
	style := paneStyle
	if m.focus == focusLists {
		style = paneFocused
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
	b.WriteString(titleStyle.Render(title) + "\n\n")
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
				b.WriteString(selStyle.Render("❯ "+line) + "\n")
			} else {
				b.WriteString("  " + line + "\n")
			}
		}
	}
	style := paneStyle
	if m.focus == focusTasks {
		style = paneFocused
	}
	return style.Width(tw).Height(m.paneHeight()).Render(strings.TrimRight(b.String(), "\n"))
}

func (m *Model) renderDetail() string {
	_, _, dw := m.paneWidths()
	var b strings.Builder
	b.WriteString(titleStyle.Render("Detail") + "\n\n")
	if t := m.selectedTask(); t != nil {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render(t.Title) + "\n")
		if len(t.Tags) > 0 {
			b.WriteString(tagStyle.Render("#"+strings.Join(t.Tags, " #")) + "\n")
		}
		b.WriteString("\n" + dimStyle.Render("Notes") + "\n" + t.Notes)
	}
	style := paneStyle
	if m.focus == focusDetail {
		style = paneFocused
	}
	return style.Width(dw).Height(m.paneHeight()).Render(strings.TrimRight(b.String(), "\n"))
}

// hint builds a footer line from key/label pairs.
func hint(pairs [][2]string) string {
	var b strings.Builder
	b.WriteString(" ")
	for _, p := range pairs {
		b.WriteString(keyStyle.Render(p[0]) + dimStyle.Render(" "+p[1]+"  "))
	}
	return strings.TrimRight(b.String(), " ")
}

func (m *Model) footer() string {
	if m.status != "" {
		return warnStyle.Render(" " + m.status)
	}
	switch m.focus {
	case focusLists:
		return hint([][2]string{{"tab", "pane"}, {"↑/↓", "move"}, {"a", "new"}, {"r", "rename"}, {"x", "delete"}, {"q", "quit"}})
	case focusTasks:
		return hint([][2]string{{"tab", "pane"}, {"↑/↓", "move"}, {"a", "add"}, {"d", "done"}, {"e", "edit"}, {"x", "delete"}, {"q", "quit"}})
	default:
		return hint([][2]string{{"tab", "pane"}, {"q", "quit"}})
	}
}
```

- [ ] **Step 6: Run the test to verify it passes**

Run: `go test ./internal/adapter/bubbletui/`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
gofmt -l internal/adapter/bubbletui/
git add internal/adapter/bubbletui/
git commit -m "feat(bubbletui): styles and normal-mode view"
```

---

### Task 3: Forms + confirm subsystem

**Files:**
- Create: `internal/adapter/bubbletui/forms.go`
- Modify: `internal/adapter/bubbletui/view.go` (`View` dispatches by mode)
- Test: `internal/adapter/bubbletui/forms_test.go`

**Interfaces:**
- Consumes: `Service` (`AddTask`, `EditTask`, `CreateList`, `RenameList`, `RemoveTask`, `DeleteList`).
- Produces: `newInput`, `refocusInputs`, `closeForm`, `openAddTask`, `openEditTask`, `openNewList`, `openRenameList`, `submitForm`, `updateForm(tea.KeyMsg)`, `confirmDeleteTask`, `confirmDeleteList`, `updateConfirm(tea.KeyMsg)`, `formView`, `confirmView`, `formTitle`. `View` now dispatches: form modes → `formView`, `modeConfirm` → `confirmView`, else `normalView`.

- [ ] **Step 1: Write the failing test**

`internal/adapter/bubbletui/forms_test.go`:
```go
package bubbletui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kendallowen/notebook/internal/todo"
)

func key(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEscape}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func typeStr(m *Model, s string) {
	for _, r := range s {
		m.updateForm(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
}

func TestAddTaskFormSubmits(t *testing.T) {
	m, svc := newTestModel(t, func(s *todo.Service) { _ = s.CreateList("work") })
	m.openAddTask()
	if m.mode != modeAddTask || len(m.inputs) != 3 {
		t.Fatalf("openAddTask state: mode=%v inputs=%d", m.mode, len(m.inputs))
	}
	typeStr(m, "buy milk")
	m.updateForm(key("tab")) // -> Tags
	typeStr(m, "store urgent")
	m.updateForm(key("enter"))

	if m.mode != modeNormal {
		t.Errorf("form should close on submit; mode=%v", m.mode)
	}
	l, _ := svc.GetList("work")
	if len(l.Tasks) != 1 || l.Tasks[0].Title != "buy milk" {
		t.Fatalf("task not added: %+v", l.Tasks)
	}
	if len(l.Tasks[0].Tags) != 2 {
		t.Errorf("tags = %v, want 2", l.Tasks[0].Tags)
	}
}

func TestEmptyTitleKeepsFormOpen(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _ = s.CreateList("work") })
	m.openAddTask()
	m.updateForm(key("enter")) // no title typed
	if m.mode != modeAddTask {
		t.Errorf("empty title should keep form open; mode=%v", m.mode)
	}
}

func TestEscCancelsForm(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _ = s.CreateList("work") })
	m.openAddTask()
	typeStr(m, "discard me")
	m.updateForm(key("esc"))
	if m.mode != modeNormal {
		t.Errorf("esc should cancel; mode=%v", m.mode)
	}
}

func TestConfirmDeletesTask(t *testing.T) {
	m, svc := newTestModel(t, func(s *todo.Service) { _, _ = s.AddTask("work", "alpha", nil, "") })
	m.confirmDeleteTask()
	if m.mode != modeConfirm {
		t.Fatalf("mode = %v, want confirm", m.mode)
	}
	m.updateConfirm(key("y"))
	l, _ := svc.GetList("work")
	if len(l.Tasks) != 0 {
		t.Errorf("task not deleted: %+v", l.Tasks)
	}
}

func TestFormViewShowsFieldsAndHint(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _ = s.CreateList("work") })
	m.openAddTask()
	out := m.View()
	for _, want := range []string{"Add task", "Title", "Notes", "esc", "cancel"} {
		if !strings.Contains(out, want) {
			t.Errorf("formView missing %q\n%s", want, out)
		}
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/adapter/bubbletui/`
Expected: FAIL — `openAddTask`/`updateForm`/etc. undefined.

- [ ] **Step 3: Create `forms.go`**

`internal/adapter/bubbletui/forms.go`:
```go
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

// submitForm performs the Service call for the active form mode.
func (m *Model) submitForm() {
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
	}
	return ""
}

func (m *Model) formView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(m.formTitle()) + "\n\n")
	for i := range m.inputs {
		b.WriteString(m.inputs[i].View() + "\n")
	}
	b.WriteString("\n" + dimStyle.Render("tab/↑↓: move · enter: submit · esc: cancel"))
	if m.status != "" {
		b.WriteString("\n" + warnStyle.Render(m.status))
	}
	return modalStyle.Render(b.String())
}

func (m *Model) confirmView() string {
	body := titleStyle.Render("Confirm") + "\n\n" + m.confirmPrompt + "\n\n" +
		dimStyle.Render("y/enter: yes · n/esc: no")
	return modalStyle.Render(body)
}
```

- [ ] **Step 4: Update `View` in `view.go` to dispatch by mode**

In `internal/adapter/bubbletui/view.go`, replace:
```go
// View renders the current mode. Form/confirm modes are added in Task 3.
func (m *Model) View() string {
	return m.normalView()
}
```
with:
```go
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
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./internal/adapter/bubbletui/`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
gofmt -l internal/adapter/bubbletui/
git add internal/adapter/bubbletui/
git commit -m "feat(bubbletui): forms and confirm modal subsystem"
```

---

### Task 4: Normal-mode Update (navigation + actions)

**Files:**
- Create: `internal/adapter/bubbletui/update.go`
- Modify: `internal/adapter/bubbletui/model.go` (remove the stub `Update`)
- Test: `internal/adapter/bubbletui/update_test.go`

**Interfaces:**
- Consumes: all Task 3 helpers; `Service.ToggleTask`.
- Produces: real `func (m *Model) Update(tea.Msg) (tea.Model, tea.Cmd)` routing by mode; `updateNormal(tea.KeyMsg)`; `toggleSelected()`; `cycleFocus(int)`; `moveSelection(int)`.

- [ ] **Step 1: Write the failing test**

`internal/adapter/bubbletui/update_test.go`:
```go
package bubbletui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kendallowen/notebook/internal/todo"
)

func send(m *Model, k tea.KeyMsg) tea.Cmd {
	_, cmd := m.Update(k)
	return cmd
}

func TestTabCyclesFocus(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _, _ = s.AddTask("work", "alpha", nil, "") })
	if m.focus != focusLists {
		t.Fatalf("initial focus = %v, want lists", m.focus)
	}
	send(m, key("tab"))
	if m.focus != focusTasks {
		t.Errorf("after tab focus = %v, want tasks", m.focus)
	}
	send(m, key("tab"))
	send(m, key("tab")) // wraps detail -> lists
	if m.focus != focusLists {
		t.Errorf("focus after wrap = %v, want lists", m.focus)
	}
}

func TestArrowMovesTaskSelection(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) {
		_, _ = s.AddTask("work", "alpha", nil, "")
		_, _ = s.AddTask("work", "beta", nil, "")
	})
	m.focus = focusTasks
	send(m, key("j"))
	if m.taskIdx != 1 {
		t.Errorf("taskIdx after down = %d, want 1", m.taskIdx)
	}
	send(m, key("j")) // clamp at last
	if m.taskIdx != 1 {
		t.Errorf("taskIdx clamped = %d, want 1", m.taskIdx)
	}
}

func TestDToggles(t *testing.T) {
	m, svc := newTestModel(t, func(s *todo.Service) { _, _ = s.AddTask("work", "alpha", nil, "") })
	m.focus = focusTasks
	send(m, key("d"))
	l, _ := svc.GetList("work")
	if !l.Tasks[0].Done {
		t.Error("d did not toggle done")
	}
}

func TestAOpensAddForm(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _, _ = s.AddTask("work", "alpha", nil, "") })
	m.focus = focusTasks
	send(m, key("a"))
	if m.mode != modeAddTask {
		t.Errorf("a should open add form; mode=%v", m.mode)
	}
}

func TestListsPaneAOpensNewList(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _, _ = s.AddTask("work", "alpha", nil, "") })
	m.focus = focusLists
	send(m, key("a"))
	if m.mode != modeNewList {
		t.Errorf("a in lists pane should open new-list form; mode=%v", m.mode)
	}
}

func TestQuitReturnsQuitCmd(t *testing.T) {
	m, _ := newTestModel(t, nil)
	cmd := send(m, key("q"))
	if cmd == nil {
		t.Fatal("q should return a quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("q command should produce tea.QuitMsg")
	}
}

func TestFormKeyRoutesToForm(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _ = s.CreateList("work") })
	m.focus = focusLists
	send(m, key("a")) // open new-list form
	// typing 'q' must go to the form, not quit
	send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if m.mode != modeNewList {
		t.Errorf("typing in a form must not trigger global keys; mode=%v", m.mode)
	}
	if m.inputs[0].Value() != "q" {
		t.Errorf("form field should contain 'q'; got %q", m.inputs[0].Value())
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/adapter/bubbletui/`
Expected: FAIL — navigation does nothing (stub Update) / duplicate `Update`.

- [ ] **Step 3: Remove the stub `Update` from `model.go`**

Delete these two lines from `internal/adapter/bubbletui/model.go`:
```go
// Update is a stub until Task 4.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
```

- [ ] **Step 4: Create `update.go`**

`internal/adapter/bubbletui/update.go`:
```go
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
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./internal/adapter/bubbletui/`
Expected: PASS.

- [ ] **Step 6: Run the full suite + vet**

Run: `go vet ./... && go test ./...`
Expected: PASS (the bubbletui package builds and all tests pass; other packages unaffected).

- [ ] **Step 7: Commit**

```bash
gofmt -l internal/adapter/bubbletui/
git add internal/adapter/bubbletui/
git commit -m "feat(bubbletui): normal-mode update (navigation + actions)"
```

---

### Task 5: CLI engine selection + composition root

**Files:**
- Modify: `internal/adapter/cli/cli.go` (signature, `--engine` flag, `resolveEngine`)
- Modify: `internal/adapter/cli/cli_test.go` (`newTestCmd` launchTUI signature)
- Modify: `internal/adapter/cli/cli_help_test.go` (`runRoot` launchTUI signature)
- Modify: `cmd/nb/main.go` (dispatch by engine)
- Test: `internal/adapter/cli/cli_engine_test.go`

**Interfaces:**
- Produces: `NewRootCmd(svc *todo.Service, launchTUI func(engine string) error) *cobra.Command`; `resolveEngine(flag string) (string, error)`.
- Consumes: `bubbletui.Run` (in main).

- [ ] **Step 1: Write the failing test**

`internal/adapter/cli/cli_engine_test.go`:
```go
package cli

import (
	"bytes"
	"testing"

	"github.com/kendallowen/notebook/internal/adapter/jsonstore"
	"github.com/kendallowen/notebook/internal/todo"
)

func runEngine(t *testing.T, env string, args ...string) (engine string, err error) {
	t.Helper()
	t.Setenv("NB_TUI", env)
	store, e := jsonstore.New(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	svc := todo.NewService(store)
	got := ""
	cmd := NewRootCmd(svc, func(eng string) error { got = eng; return nil })
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	return got, cmd.Execute()
}

func TestEngineResolution(t *testing.T) {
	if e, err := runEngine(t, "", "tui"); err != nil || e != "tview" {
		t.Errorf("default = %q (err %v), want tview", e, err)
	}
	if e, err := runEngine(t, "", "tui", "--engine", "bubble"); err != nil || e != "bubble" {
		t.Errorf("flag = %q (err %v), want bubble", e, err)
	}
	if e, err := runEngine(t, "bubble", "tui"); err != nil || e != "bubble" {
		t.Errorf("env = %q (err %v), want bubble", e, err)
	}
	if e, err := runEngine(t, "bubble", "tui", "-e", "tview"); err != nil || e != "tview" {
		t.Errorf("flag-overrides-env = %q (err %v), want tview", e, err)
	}
	if _, err := runEngine(t, "", "tui", "--engine", "nope"); err == nil {
		t.Error("invalid engine should error")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/adapter/cli/`
Expected: FAIL — `NewRootCmd` signature mismatch / `--engine` unknown.

- [ ] **Step 3: Update `cli.go`**

In `internal/adapter/cli/cli.go`: add `"fmt"` to imports, change the signature, add `resolveEngine`, and give the `tui` command the `--engine` flag. Replace the function up to the `tui` command with:
```go
// resolveEngine applies the engine rule: flag, then $NB_TUI, then "tview".
func resolveEngine(flag string) (string, error) {
	e := flag
	if e == "" {
		e = os.Getenv("NB_TUI")
	}
	if e == "" {
		e = "tview"
	}
	switch e {
	case "tview", "bubble":
		return e, nil
	default:
		return "", fmt.Errorf("invalid engine %q (want \"tview\" or \"bubble\")", e)
	}
}

// NewRootCmd builds the command tree. launchTUI runs the interactive UI for
// the chosen engine; bare `nb` prints help.
func NewRootCmd(svc *todo.Service, launchTUI func(engine string) error) *cobra.Command {
	root := &cobra.Command{
		Use:           "nb",
		Short:         "notebook — a CLI + TUI task tracker",
		Long:          banner + "\n\nnotebook — a CLI + TUI task tracker.\nRun `nb tui` for the interactive interface.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	var engine string
	tui := &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := resolveEngine(engine)
			if err != nil {
				return err
			}
			return launchTUI(eng)
		},
	}
	tui.Flags().StringVarP(&engine, "engine", "e", "", `TUI engine: "tview" (default) or "bubble"; or $NB_TUI`)
```
(Leave the rest of `NewRootCmd` — the `root.AddCommand(...)` lines and `return root` — unchanged.)

- [ ] **Step 4: Update the two test harnesses**

In `internal/adapter/cli/cli_test.go`, change the `newTestCmd` launchTUI argument:
```go
		cmd := NewRootCmd(svc, func() error { return nil })
```
to:
```go
		cmd := NewRootCmd(svc, func(string) error { return nil })
```

In `internal/adapter/cli/cli_help_test.go`, change the `runRoot` launchTUI argument:
```go
	cmd := NewRootCmd(svc, func() error { launched = true; return nil })
```
to:
```go
	cmd := NewRootCmd(svc, func(string) error { launched = true; return nil })
```

- [ ] **Step 5: Update `cmd/nb/main.go`**

In `cmd/nb/main.go`, add the import `"github.com/kendallowen/notebook/internal/adapter/bubbletui"` and replace the `launchTUI` definition:
```go
	launchTUI := func() error {
		return tui.New(svc).Run()
	}
```
with:
```go
	launchTUI := func(engine string) error {
		if engine == "bubble" {
			return bubbletui.Run(svc)
		}
		return tui.New(svc).Run()
	}
```

- [ ] **Step 6: Run the tests to verify they pass**

Run: `go test ./... && go vet ./...`
Expected: PASS, vet clean.

- [ ] **Step 7: Commit**

```bash
gofmt -l ./internal/adapter/cli/ ./cmd/nb/
git add internal/adapter/cli/ cmd/nb/
git commit -m "feat(cli): --engine selection wiring bubbletui adapter"
```

---

### Task 6: README + final verification

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Update the README**

In `README.md`, under `## CLI`, change the `nb tui` line and add an engine note. Replace:
```markdown
    nb tui                   # launch the interactive TUI
```
with:
```markdown
    nb tui                   # launch the interactive TUI (tview by default)
    nb tui --engine bubble   # launch the Bubble Tea TUI (or NB_TUI=bubble)
```

- [ ] **Step 2: Build, vet, and run the full suite**

Run:
```bash
go build ./...
go vet ./...
go test ./...
```
Expected: all PASS, no vet output.

- [ ] **Step 3: Non-interactive smoke of engine resolution**

Run:
```bash
go install ./cmd/nb
NB_DIR="$(mktemp -d)" nb add "from cli" -l demo
nb tui --engine nope ; echo "exit=$?"
```
Expected: `add` prints `Added [demo] #1: from cli`; `nb tui --engine nope` prints the `invalid engine "nope"` error and exits non-zero. (Do not launch the interactive TUIs headlessly — they need a real terminal; interactive verification is left to the human.)

- [ ] **Step 4: Commit**

```bash
git add README.md
git commit -m "docs: document --engine bubble TUI option"
```

---

## Notes for the implementer

- All TUI logic stays in `internal/adapter/bubbletui` and goes through the `Service` — never touch the filesystem or other adapters directly.
- Bubble Tea tests are deterministic: build a `*Model`, send `tea.KeyMsg` values to `Update`/`updateForm`/`updateConfirm`, and assert state or `View()` substrings. No event loop or terminal needed. Lipgloss strips color when output isn't a TTY (as in `go test`), so `View()` substring assertions are stable.
- Keep test task/list titles short (e.g. `alpha`) so they aren't truncated by pane-width rendering.
- Interactive TUIs (both engines) require a real terminal; the headless agent cannot drive them. Rely on the unit tests + leave a live `nb tui --engine bubble` check to the human.
