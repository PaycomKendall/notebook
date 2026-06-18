# Bubble Tea Themes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add five selectable color themes (default, nord, dracula, gruvbox, mono) to the Bubble Tea TUI adapter, chosen via `nb tui --theme <name>` / `$NB_THEME`.

**Architecture:** Move the bubbletui adapter's Lipgloss styling from package-global vars onto the `Model` (a `Styles` struct built from a `Theme`). A `Theme` is six colors; `default`/`mono` use `AdaptiveColor` (light/dark), the named themes use fixed palettes. The `tui` command resolves a raw theme string and passes it through `launchTUI(engine, theme)`; `bubbletui.Run` validates it.

**Tech Stack:** Go 1.24, `charmbracelet/bubbletea`/`lipgloss`/`bubbles`, Cobra.

## Global Constraints

- **Module path:** `github.com/kendallowen/notebook`.
- **Hexagonal rule:** `internal/adapter/bubbletui` imports only `internal/todo` + Charm libs. `cli` does NOT import bubbletui — theme validation lives in bubbletui. Only `cmd/nb/main.go` imports adapters.
- **Themes apply to the bubble engine only.** tview ignores `--theme`.
- **Selection:** `tui` flag `--theme` (long only); resolution flag → `$NB_THEME` → `"default"`. Invalid → `invalid theme "%s" (want default, nord, dracula, gruvbox, mono)`, returned by `bubbletui.Run` before the program starts.
- **`launchTUI` signature:** `func(engine, theme string) error`.
- **Theme model:** `Theme` = 6 `lipgloss.TerminalColor` fields (`accent, secondary, subtle, selBg, selFg, warn`); `(Theme).styles()` builds a `Styles` struct; `Model.styles Styles`; `New(svc, theme Theme)`.
- **Adaptive:** `default` and `mono` use `lipgloss.AdaptiveColor{Light,Dark}`; `nord`/`dracula`/`gruvbox` fixed.
- **TDD:** test-first; commit after each task; `gofmt -l` before each commit.

**Prerequisite:** The bubbletui adapter, cli, and engine wiring are complete on `main`. Go 1.24 installed.

---

### Task 1: Theme/Styles refactor (styling onto the Model, `default` theme only)

**Files:**
- Create: `internal/adapter/bubbletui/theme.go`
- Rewrite: `internal/adapter/bubbletui/styles.go`
- Modify: `internal/adapter/bubbletui/model.go` (add `styles` field; `New` takes a `Theme`; `Run` passes `themeDefault`)
- Rewrite: `internal/adapter/bubbletui/view.go` (use `m.styles`; `hint` becomes a method)
- Modify: `internal/adapter/bubbletui/forms.go` (`formView`/`confirmView` use `m.styles`)
- Modify: `internal/adapter/bubbletui/model_test.go` (`newTestModel` passes `themeDefault`)
- Test: `internal/adapter/bubbletui/theme_test.go`

**Interfaces:**
- Produces: `type Theme struct { accent, secondary, subtle, selBg, selFg, warn lipgloss.TerminalColor }`; `var themeDefault Theme`; `type Styles struct { title, dim, key, sel, tag, warn, pane, paneFocused, modal lipgloss.Style }`; `func (t Theme) styles() Styles`; `func New(svc *todo.Service, theme Theme) *Model` (now sets `m.styles`); `func (m *Model) hint(pairs [][2]string) string`.

- [ ] **Step 1: Write the failing test**

`internal/adapter/bubbletui/theme_test.go`:
```go
package bubbletui

import "testing"

func TestStylesDeriveFromTheme(t *testing.T) {
	m, _ := newTestModel(t, nil)
	if got := m.styles.title.GetForeground(); got != themeDefault.accent {
		t.Errorf("title foreground = %v, want theme accent %v", got, themeDefault.accent)
	}
	if got := m.styles.sel.GetBackground(); got != themeDefault.selBg {
		t.Errorf("sel background = %v, want theme selBg %v", got, themeDefault.selBg)
	}
	if got := m.styles.paneFocused.GetBorderTopForeground(); got != themeDefault.accent {
		t.Errorf("focused border = %v, want accent %v", got, themeDefault.accent)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/adapter/bubbletui/`
Expected: FAIL — `m.styles`/`themeDefault` undefined and `New` arity mismatch.

- [ ] **Step 3: Create `theme.go`**

`internal/adapter/bubbletui/theme.go`:
```go
package bubbletui

import "github.com/charmbracelet/lipgloss"

// Theme is the six-color palette every style derives from.
type Theme struct {
	accent    lipgloss.TerminalColor // focused border, tags, titles
	secondary lipgloss.TerminalColor // unfocused borders, help keys
	subtle    lipgloss.TerminalColor // dim/help text
	selBg     lipgloss.TerminalColor // selected-row background
	selFg     lipgloss.TerminalColor // selected-row text
	warn      lipgloss.TerminalColor // status/errors
}

// themeDefault is the original Charm look, adaptive to light/dark terminals.
var themeDefault = Theme{
	accent:    lipgloss.AdaptiveColor{Light: "205", Dark: "212"},
	secondary: lipgloss.AdaptiveColor{Light: "63", Dark: "99"},
	subtle:    lipgloss.AdaptiveColor{Light: "244", Dark: "245"},
	selBg:     lipgloss.AdaptiveColor{Light: "189", Dark: "57"},
	selFg:     lipgloss.AdaptiveColor{Light: "236", Dark: "231"},
	warn:      lipgloss.AdaptiveColor{Light: "160", Dark: "203"},
}
```

- [ ] **Step 4: Rewrite `styles.go`**

Replace the entire contents of `internal/adapter/bubbletui/styles.go` with:
```go
package bubbletui

import "github.com/charmbracelet/lipgloss"

// Styles holds the computed lipgloss styles for a Theme.
type Styles struct {
	title, dim, key, sel, tag, warn lipgloss.Style
	pane, paneFocused, modal        lipgloss.Style
}

// styles builds all styles from the theme's six colors.
func (t Theme) styles() Styles {
	return Styles{
		title: lipgloss.NewStyle().Bold(true).Foreground(t.accent),
		dim:   lipgloss.NewStyle().Foreground(t.subtle),
		key:   lipgloss.NewStyle().Bold(true).Foreground(t.secondary),
		sel:   lipgloss.NewStyle().Foreground(t.selFg).Background(t.selBg).Bold(true),
		tag:   lipgloss.NewStyle().Foreground(t.accent),
		warn:  lipgloss.NewStyle().Foreground(t.warn),
		pane:        lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.secondary).Padding(0, 1),
		paneFocused: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.accent).Padding(0, 1),
		modal:       lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.accent).Padding(1, 2),
	}
}
```

- [ ] **Step 5: Update `model.go`**

Add a `styles` field to the struct — change:
```go
	status        string
	width, height int
}
```
to:
```go
	status        string
	width, height int

	styles Styles
}
```
Change `New` to take a theme and build styles — replace:
```go
// New builds a Model and loads the initial lists + tasks.
func New(svc *todo.Service) *Model {
	m := &Model{svc: svc, width: 90, height: 24, focus: focusTasks}
	m.reloadLists()
	m.reloadTasks()
	return m
}
```
with:
```go
// New builds a Model with the given theme and loads the initial lists + tasks.
func New(svc *todo.Service, theme Theme) *Model {
	m := &Model{svc: svc, width: 90, height: 24, focus: focusTasks, styles: theme.styles()}
	m.reloadLists()
	m.reloadTasks()
	return m
}
```
Update `Run`'s `New` call (theme wiring lands in Task 3) — replace:
```go
	p := tea.NewProgram(New(svc), tea.WithAltScreen())
```
with:
```go
	p := tea.NewProgram(New(svc, themeDefault), tea.WithAltScreen())
```

- [ ] **Step 6: Rewrite `view.go`**

Replace the entire contents of `internal/adapter/bubbletui/view.go` with:
```go
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
```

- [ ] **Step 7: Update `forms.go`**

In `internal/adapter/bubbletui/forms.go`, replace `formView` and `confirmView` with:
```go
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
```

- [ ] **Step 8: Update `newTestModel` in `model_test.go`**

In `internal/adapter/bubbletui/model_test.go`, change:
```go
	return New(svc), svc
```
to:
```go
	return New(svc, themeDefault), svc
```

- [ ] **Step 9: Run the tests to verify they pass**

Run: `go test ./internal/adapter/bubbletui/`
Expected: PASS (the new theme test plus all existing view/forms/update tests, which still pass since color is stripped in headless tests).

- [ ] **Step 10: Commit**

```bash
gofmt -l internal/adapter/bubbletui/
git add internal/adapter/bubbletui/
git commit -m "refactor(bubbletui): styles on the Model, built from a Theme (default)"
```

---

### Task 2: Theme presets + resolveTheme

**Files:**
- Modify: `internal/adapter/bubbletui/theme.go` (add 4 presets, `themes` map, `resolveTheme`)
- Test: `internal/adapter/bubbletui/theme_test.go` (append)

**Interfaces:**
- Produces: `var themeNord, themeDracula, themeGruvbox, themeMono Theme`; `var themes map[string]Theme`; `func resolveTheme(name string) (Theme, error)`.

- [ ] **Step 1: Write the failing test**

Append to `internal/adapter/bubbletui/theme_test.go`:
```go
func TestResolveThemeDefaultsAndKnown(t *testing.T) {
	got, err := resolveTheme("")
	if err != nil {
		t.Fatalf("empty: %v", err)
	}
	if got != themeDefault {
		t.Error("empty name should resolve to the default theme")
	}
	for _, name := range []string{"default", "nord", "dracula", "gruvbox", "mono"} {
		if _, err := resolveTheme(name); err != nil {
			t.Errorf("resolveTheme(%q) = %v", name, err)
		}
	}
}

func TestResolveThemeInvalid(t *testing.T) {
	if _, err := resolveTheme("nope"); err == nil {
		t.Error("invalid theme name should return an error")
	}
}

func TestThemesAreDistinct(t *testing.T) {
	if themes["nord"].accent == themes["dracula"].accent {
		t.Error("nord and dracula accents should differ")
	}
	if themes["gruvbox"].accent == themes["mono"].accent {
		t.Error("gruvbox and mono accents should differ")
	}
}

func TestStylesFollowChosenTheme(t *testing.T) {
	m, _ := newTestModel(t, nil)
	m2 := New(m.svc, themes["nord"])
	if got := m2.styles.title.GetForeground(); got != themeNord.accent {
		t.Errorf("nord title foreground = %v, want %v", got, themeNord.accent)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/adapter/bubbletui/`
Expected: FAIL — `resolveTheme`/`themes`/`themeNord` undefined.

- [ ] **Step 3: Add presets + resolution to `theme.go`**

Add `"fmt"` to the imports of `internal/adapter/bubbletui/theme.go` (so the import block becomes):
```go
import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)
```
Append to `internal/adapter/bubbletui/theme.go`:
```go
// themeNord — cool arctic blues/grays (fixed).
var themeNord = Theme{
	accent:    lipgloss.Color("110"),
	secondary: lipgloss.Color("109"),
	subtle:    lipgloss.Color("102"),
	selBg:     lipgloss.Color("24"),
	selFg:     lipgloss.Color("189"),
	warn:      lipgloss.Color("167"),
}

// themeDracula — dark with vivid purple/pink (fixed).
var themeDracula = Theme{
	accent:    lipgloss.Color("212"),
	secondary: lipgloss.Color("141"),
	subtle:    lipgloss.Color("103"),
	selBg:     lipgloss.Color("60"),
	selFg:     lipgloss.Color("231"),
	warn:      lipgloss.Color("203"),
}

// themeGruvbox — warm retro earth tones (fixed).
var themeGruvbox = Theme{
	accent:    lipgloss.Color("208"),
	secondary: lipgloss.Color("214"),
	subtle:    lipgloss.Color("245"),
	selBg:     lipgloss.Color("237"),
	selFg:     lipgloss.Color("223"),
	warn:      lipgloss.Color("167"),
}

// themeMono — grayscale + a red warn, adaptive to light/dark.
var themeMono = Theme{
	accent:    lipgloss.AdaptiveColor{Light: "238", Dark: "252"},
	secondary: lipgloss.AdaptiveColor{Light: "244", Dark: "245"},
	subtle:    lipgloss.AdaptiveColor{Light: "248", Dark: "240"},
	selBg:     lipgloss.AdaptiveColor{Light: "252", Dark: "238"},
	selFg:     lipgloss.AdaptiveColor{Light: "232", Dark: "255"},
	warn:      lipgloss.AdaptiveColor{Light: "160", Dark: "203"},
}

var themes = map[string]Theme{
	"default": themeDefault,
	"nord":    themeNord,
	"dracula": themeDracula,
	"gruvbox": themeGruvbox,
	"mono":    themeMono,
}

// resolveTheme maps a name to a Theme; "" -> default; unknown -> error.
func resolveTheme(name string) (Theme, error) {
	if name == "" {
		name = "default"
	}
	if t, ok := themes[name]; ok {
		return t, nil
	}
	return Theme{}, fmt.Errorf("invalid theme %q (want default, nord, dracula, gruvbox, mono)", name)
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/adapter/bubbletui/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
gofmt -l internal/adapter/bubbletui/
git add internal/adapter/bubbletui/
git commit -m "feat(bubbletui): nord/dracula/gruvbox/mono presets + resolveTheme"
```

---

### Task 3: Wiring — `--theme` flag through to `bubbletui.Run`

**Files:**
- Modify: `internal/adapter/bubbletui/model.go` (`Run` takes a theme name, resolves it)
- Modify: `internal/adapter/cli/cli.go` (`launchTUI` signature, `--theme` flag, raw theme resolution)
- Modify: `cmd/nb/main.go` (pass theme to `bubbletui.Run`)
- Modify: `internal/adapter/cli/cli_test.go`, `cli_help_test.go`, `cli_robust_test.go`, `cli_engine_test.go` (callback signature)
- Test: `internal/adapter/cli/cli_theme_test.go`

**Interfaces:**
- Produces: `func Run(svc *todo.Service, themeName string) error`; `NewRootCmd(svc *todo.Service, launchTUI func(engine, theme string) error) *cobra.Command`.

- [ ] **Step 1: Write the failing test**

`internal/adapter/cli/cli_theme_test.go`:
```go
package cli

import (
	"bytes"
	"testing"

	"github.com/kendallowen/notebook/internal/adapter/jsonstore"
	"github.com/kendallowen/notebook/internal/todo"
)

func runTheme(t *testing.T, env string, args ...string) (theme string, err error) {
	t.Helper()
	t.Setenv("NB_THEME", env)
	store, e := jsonstore.New(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	svc := todo.NewService(store)
	got := ""
	cmd := NewRootCmd(svc, func(_, th string) error { got = th; return nil })
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	return got, cmd.Execute()
}

func TestThemeResolution(t *testing.T) {
	if th, err := runTheme(t, "", "tui"); err != nil || th != "default" {
		t.Errorf("default = %q (err %v), want default", th, err)
	}
	if th, err := runTheme(t, "", "tui", "--theme", "nord"); err != nil || th != "nord" {
		t.Errorf("flag = %q (err %v), want nord", th, err)
	}
	if th, err := runTheme(t, "dracula", "tui"); err != nil || th != "dracula" {
		t.Errorf("env = %q (err %v), want dracula", th, err)
	}
	if th, err := runTheme(t, "dracula", "tui", "--theme", "mono"); err != nil || th != "mono" {
		t.Errorf("flag-overrides-env = %q (err %v), want mono", th, err)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/adapter/cli/`
Expected: FAIL — `NewRootCmd` callback arity mismatch / `--theme` unknown.

- [ ] **Step 3: Update `bubbletui.Run` in `model.go`**

In `internal/adapter/bubbletui/model.go`, replace `Run`:
```go
// Run starts the Bubble Tea program in the alternate screen.
func Run(svc *todo.Service) error {
	p := tea.NewProgram(New(svc, themeDefault), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
```
with:
```go
// Run resolves the theme name and starts the Bubble Tea program (alt screen).
func Run(svc *todo.Service, themeName string) error {
	theme, err := resolveTheme(themeName)
	if err != nil {
		return err
	}
	p := tea.NewProgram(New(svc, theme), tea.WithAltScreen())
	_, err = p.Run()
	return err
}
```

- [ ] **Step 4: Update `cli.go`**

In `internal/adapter/cli/cli.go`, change the `NewRootCmd` signature and the `tui` command. Replace:
```go
func NewRootCmd(svc *todo.Service, launchTUI func(engine string) error) *cobra.Command {
```
with:
```go
func NewRootCmd(svc *todo.Service, launchTUI func(engine, theme string) error) *cobra.Command {
```
Then replace the `tui` command block (the `var engine string` declaration, the `tui := &cobra.Command{...}` with its RunE, and the `tui.Flags()...` line) with:
```go
	var engine, theme string
	tui := &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := resolveEngine(engine)
			if err != nil {
				return err
			}
			th := theme
			if th == "" {
				th = os.Getenv("NB_THEME")
			}
			if th == "" {
				th = "default"
			}
			return launchTUI(eng, th)
		},
	}
	tui.Flags().StringVarP(&engine, "engine", "e", "", `TUI engine: "tview" (default) or "bubble"; or $NB_TUI`)
	tui.Flags().StringVar(&theme, "theme", "", `bubble theme: default, nord, dracula, gruvbox, mono; or $NB_THEME`)
```
(Leave `resolveEngine`, the `root.AddCommand(...)` lines, and `return root` unchanged.)

- [ ] **Step 5: Update `cmd/nb/main.go`**

In `cmd/nb/main.go`, replace the `launchTUI` definition:
```go
	launchTUI := func(engine string) error {
		if engine == "bubble" {
			return bubbletui.Run(svc)
		}
		return tui.New(svc).Run()
	}
```
with:
```go
	launchTUI := func(engine, theme string) error {
		if engine == "bubble" {
			return bubbletui.Run(svc, theme)
		}
		return tui.New(svc).Run()
	}
```

- [ ] **Step 6: Update the existing cli test harnesses**

These three call `NewRootCmd` with the old callback — update each to the two-arg form:
- `internal/adapter/cli/cli_test.go`: change `func(string) error { return nil }` to `func(string, string) error { return nil }`.
- `internal/adapter/cli/cli_help_test.go`: change `func(string) error { launched = true; return nil }` to `func(string, string) error { launched = true; return nil }`.
- `internal/adapter/cli/cli_robust_test.go`: change `func(string) error { return nil }` to `func(string, string) error { return nil }`.
- `internal/adapter/cli/cli_engine_test.go`: in `runEngine`, change `func(eng string) error { got = eng; return nil }` to `func(eng, _ string) error { got = eng; return nil }`.

- [ ] **Step 7: Run the tests to verify they pass**

Run: `go test ./... && go vet ./...`
Expected: PASS, vet clean.

- [ ] **Step 8: Commit**

```bash
gofmt -l ./internal/adapter/ ./cmd/nb/
git add internal/adapter/ cmd/nb/
git commit -m "feat(cli): --theme flag wiring bubbletui themes"
```

---

### Task 4: README + final verification

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Update the README**

In `README.md`, under `## CLI`, replace:
```markdown
    nb tui --engine bubble   # launch the Bubble Tea TUI (or NB_TUI=bubble)
```
with:
```markdown
    nb tui --engine bubble   # launch the Bubble Tea TUI (or NB_TUI=bubble)
    nb tui --engine bubble --theme nord   # themes: default, nord, dracula, gruvbox, mono (or NB_THEME)
```

- [ ] **Step 2: Build, vet, full suite**

Run:
```bash
go build ./...
go vet ./...
go test ./...
```
Expected: all PASS, no vet output.

- [ ] **Step 3: Non-interactive smoke**

Run:
```bash
go install ./cmd/nb
NB_DIR="$(mktemp -d)" nb add "theme smoke" -l demo
nb tui --engine bubble --theme nope ; echo "exit=$?"
```
Expected: `add` prints `Added [demo] #1: theme smoke`; `nb tui --engine bubble --theme nope` prints `error: invalid theme "nope" (want default, nord, dracula, gruvbox, mono)` and exits non-zero (the error surfaces from `bubbletui.Run` before the program starts — no TTY needed for the error path). Do NOT launch the interactive TUIs headlessly; live theme verification is left to the human.

- [ ] **Step 4: Commit**

```bash
git add README.md
git commit -m "docs: document --theme presets"
```

---

## Notes for the implementer

- The styling refactor (Task 1) must land atomically: removing the package-global style vars breaks `view.go`/`forms.go` until they use `m.styles`, and `New`'s new signature breaks `Run` + `newTestModel` until updated — all in the same task.
- `lipgloss.Style.GetForeground()`/`GetBackground()`/`GetBorderTopForeground()` return the `lipgloss.TerminalColor` that was set, so theme tests can assert styles derive from the chosen palette without a terminal. Color is otherwise stripped in headless tests.
- The cli never imports bubbletui; it passes the raw theme string and `bubbletui.Run` validates it. So `--theme nope` errors only under `--engine bubble` — the unit test for the invalid path is `resolveTheme` (Task 2), not a cli test.
- Preset ANSI-256 values are approximate and tunable by eye later; don't "improve" them — they're the spec's chosen values.
