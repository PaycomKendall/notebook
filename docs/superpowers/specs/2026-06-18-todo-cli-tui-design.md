# Todo — CLI + TUI task tracker (design)

**Date:** 2026-06-18
**Status:** Approved, ready for planning

## 1. Summary

A personal task tracker written in **Go**, exposing two front-ends over one shared
core:

- A **full CLI** (Cobra) of scriptable subcommands.
- An **interactive TUI** (tview/tcell) with a three-pane layout.

Tasks are organized into named **lists**. Each task has a title, a done/not-done
state, free-form **tags**, and a longer **notes** field. Data is stored as one
human-readable JSON file per list.

The architecture is **hexagonal (ports & adapters)**: a dependency-free domain
core defines a repository *port*; the JSON store, the CLI, and the TUI are
*adapters*. Both front-ends drive the same application `Service`, so their
behavior cannot diverge.

## 2. Goals / non-goals

**Goals**
- Capture and manage todos from the shell (scriptable) *and* interactively (TUI).
- Multiple named lists; switch between them in the TUI, target them by flag in the CLI.
- Human-readable, hand-editable, git-trackable storage.
- A clean core that is fully unit-testable without a terminal or disk.

**Non-goals (explicitly out of scope)**
- Priorities and due dates (deliberately omitted — only tags and notes).
- Sync, multi-user, networking, or any server.
- SQL/database storage, migrations, CQRS, idempotency keys, soft-deletes.
- File locking / concurrent-writer safety (single-user tool; last-write-wins).
- `--json` machine-readable CLI output (may be added later if needed).

## 3. Stack & tooling

- **Language:** Go (1.22 or newer). Installed natively via `brew install go`.
- **CLI framework:** `github.com/spf13/cobra`.
- **TUI:** `github.com/rivo/tview` on `github.com/gdamore/tcell/v2`.
- **Storage:** standard library `encoding/json`, `os`, `path/filepath`, `time`.
- **Module path:** `github.com/kendallowen/todo` (adjust if needed).
- **Repo:** standalone git repository at `~/projects/todo`.
- **Run (dev):** `go run ./cmd/todo [...]`.
- **Run (daily):** `go install ./cmd/todo` → native `todo` on `$PATH` (`~/go/bin`).
  (User should ensure `~/go/bin` is on `PATH`.)
- **Not used:** Docker — redundant for a self-contained native binary and hostile
  to an interactive TUI that needs a TTY and a local data dir.

## 4. Architecture — hexagonal (ports & adapters)

```
todo/
  cmd/todo/main.go              composition root — wires repo -> service -> front-ends
  internal/
    todo/                       CORE (the hexagon) — no I/O, no tview, no cobra
      task.go                   Task entity + tag handling + validation
      list.go                   List + mutation methods
      repository.go             ListRepository PORT (interface), defined by the core
      service.go                Service — application use cases both front-ends call
    adapter/
      jsonstore/                driven adapter: ListRepository over JSON files
        store.go
      cli/                      driving adapter: Cobra commands -> Service
        root.go add.go ls.go done.go rm.go edit.go tag.go lists.go
      tui/                      driving adapter: tview app -> Service
        app.go
```

**Dependency rule:** `domain (todo) <- service <- driving adapters (cli, tui)`,
and `jsonstore -> todo` (it implements the core's port). `cmd/todo/main.go` is the
only place that imports all of them and wires them together.

## 5. Domain model (`internal/todo`)

```go
type Task struct {
    ID      int       // stable, monotonic per list, never reused (gaps OK)
    Title   string
    Done    bool
    Tags    []string
    Notes   string
    Created time.Time
    Updated time.Time
}

type List struct {
    Name   string      // also the filename stem
    NextID int         // monotonic counter -> stable IDs
    Tasks  []Task
}
```

**List mutation methods** (the shared behavior; pure, no I/O):
- `Add(title string) (*Task, error)` — validates non-empty title, assigns `NextID`, increments it.
- `Get(id int) (*Task, error)`
- `Toggle(id int) error` — flips `Done`, bumps `Updated`.
- `SetDone(id int, done bool) error`
- `Remove(id int) error`
- `SetTitle(id int, title string) error`
- `SetNotes(id int, notes string) error`
- `AddTag(id int, tag string) error` — normalizes (trim, lowercase), dedupes.
- `RemoveTag(id int, tag string) error`

**Validation rules**
- Task title: required, non-empty after trim.
- Tag: non-empty after trim; stored normalized (trimmed, lowercased); duplicates ignored.
- List name: non-empty; restricted to a filesystem-safe charset
  (`[a-z0-9._-]`, lowercased) so it can be a filename stem.

## 6. Port (`internal/todo/repository.go`)

```go
type ListRepository interface {
    Names() ([]string, error)          // all list names
    Load(name string) (*List, error)   // ErrListNotFound if missing
    Save(list *List) error             // create-or-overwrite
    Create(name string) (*List, error) // ErrListExists if already present
    Delete(name string) error
    Rename(oldName, newName string) error
}
```

Sentinel errors: `ErrListNotFound`, `ErrListExists`, `ErrTaskNotFound`.

## 7. Application service (`internal/todo/service.go`)

`Service` wraps a `ListRepository` and exposes the use cases both front-ends call.
Each mutating use case follows load → mutate (via List methods) → save. Examples:

- `AddTask(list, title string, tags []string, notes string) (Task, error)`
- `ToggleTask(list string, id int) error`
- `SetTaskDone(list string, id int, done bool) error`
- `RemoveTask(list string, id int) error`
- `EditTask(list string, id int, title, notes *string) error` (nil pointer = leave unchanged)
- `AddTaskTag(list string, id int, tag string) error` / `RemoveTaskTag(...)`
- `GetList(name string) (*List, error)`
- `ListNames() ([]string, error)`
- `CreateList(name string) error` / `DeleteList(name string) error` / `RenameList(old, new string) error`

**List name resolution & auto-creation.** Front-ends always pass an explicit
list name to the service — defaulting an absent `-l` flag to `inbox` is an
*adapter* concern (§9), not the service's. The service's rule is about *missing*
lists:
- `AddTask` (and `CreateList`) **auto-create** the target list if it doesn't exist —
  so `todo add -l groceries "milk"` just works and the first-ever `todo add`
  creates `inbox`.
- All other mutations (`ToggleTask`, `RemoveTask`, `EditTask`, tag ops) and reads
  (`GetList`) return `ErrListNotFound` for a missing list rather than creating it,
  so a mistyped list name surfaces as an error instead of silently spawning an
  empty list.

## 8. Storage adapter (`internal/todo/.../jsonstore`)

- **Location resolution (in order):** `TODO_DIR` env var → `$XDG_DATA_HOME/todo`
  → `~/.local/share/todo`. Directory created on first use.
- **Layout:** one file per list, `<dir>/<name>.json`, pretty-printed JSON.
- **Serialization:** a persistence DTO mirrors the domain `List`/`Task`; the adapter
  maps domain ↔ DTO (keeps JSON tags out of the domain types).
- **Atomic writes:** write to `<name>.json.tmp`, `fsync`, then `rename` over the
  target, so an interrupted save never corrupts an existing list.
- **`Names()`:** scans the dir for `*.json` files, returns stems (sorted).
- **Errors:** missing file → `ErrListNotFound`; existing on `Create` → `ErrListExists`;
  malformed JSON → wrapped parse error (never silently dropped).

## 9. CLI adapter (`cli`, Cobra)

Default list `inbox`; `-l/--list` overrides; `TODO_LIST` env changes the default.
Global `--dir` mirrors `TODO_DIR`.

```
todo                          launch TUI (no subcommand)
todo tui                      launch TUI explicitly
todo add <title…> [-l list] [-t tag]… [-n note]   create task; prints new ID
todo ls [-l list | -a/--all] [-t tag] [--done|--open]   list tasks: ID, checkbox, tags
todo done <id…> [-l list]     mark done (accepts multiple ids)
todo undone <id…> [-l list]   unmark
todo rm <id…> [-l list]       delete
todo edit <id> [--title …] [-n note] [-l list]    edit fields
todo tag <id> [--add t]… [--rm t]… [-l list]      manage tags
todo lists                    all lists with open/done counts
todo lists new <name>
todo lists rm <name> [--force]
todo lists rename <old> <new>
```

- Output is plain text, human-readable.
- The command tree is built around an injected `*todo.Service` (not a global), so
  tests construct a fresh tree over a temp-dir repository.
- `todo` with no subcommand and `todo tui` both launch the TUI adapter.

## 10. TUI adapter (`tui`, tview)

Three-pane `Flex`: **Lists | Tasks | Detail**.

- **Lists pane** (`tview.List`): list names with open/done counts.
- **Tasks pane** (`tview.List` or `tview.Table`): tasks of the selected list, each
  rendered as `[ ]`/`[x]` + title + inline tags.
- **Detail pane** (`tview.TextView`): selected task's title, tags, and notes.

**Navigation**
- `Tab` / `Shift-Tab` cycle panes; arrows / `j` `k` move within a pane.
- Selecting a list reloads the Tasks pane; moving in Tasks updates Detail.

**Keys**
- Tasks pane: `a` add, `d` toggle done, `e` edit (modal form: title/notes/tags),
  `x` delete (confirm modal), `n` edit notes.
- Lists pane: `a` new list, `r` rename, `x` delete list (confirm).
- Global: `?` help overlay, `q` / `Ctrl-C` quit.

**Modals:** `tview.Form` for add/edit; `tview.Modal` for confirmations.

**Persistence:** every mutation calls a `Service` method that persists immediately;
there is no explicit "save" action.

## 11. Composition root (`cmd/todo/main.go`)

1. Resolve the data dir (flag/env/default).
2. Construct the `jsonstore` repository.
3. Construct the `todo.Service` over it.
4. Build the Cobra command tree with the service injected; the root command (no
   subcommand) and `tui` launch the TUI adapter with the same service.
5. `Execute()`.

## 12. Testing (TDD)

- **domain** — pure unit tests: `Add` assigns/increments `NextID`; `Toggle` flips
  `Done` and bumps `Updated`; `AddTag` normalizes + dedupes; validation rejects
  empty titles and bad list names. Table-driven.
- **service** — tested against an **in-memory fake `ListRepository`** (no disk):
  verifies each use case orchestrates load→mutate→save correctly.
- **jsonstore** — temp-dir integration tests: save/load round-trip; atomic write
  leaves no `.tmp` file; `Names()` scan; create/delete/rename; name sanitization;
  malformed-JSON handling; location resolution honoring `TODO_DIR`.
- **cli** — invoke Cobra commands against a `Service` over a temp dir; capture
  stdout; assert output and side effects.
- **tui** — kept thin (logic lives in `Service`); a light smoke test drives key
  events through `tcell`'s `SimulationScreen` for add + toggle. TUI coverage is
  acknowledged to be lighter than the other layers.

## 13. Open items / future (not in this build)

- `--json` CLI output for scripting.
- Cross-list tag search (`todo find #tag`).
- Advisory file locking if concurrent CLI+TUI editing becomes a real problem.
- Shell completions (Cobra can generate them) and a Homebrew formula.
