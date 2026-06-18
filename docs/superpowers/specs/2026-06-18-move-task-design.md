# Move a task between lists (design)

**Date:** 2026-06-18
**Status:** Approved, ready for planning

## 1. Summary

Add the ability to move a task from one list to another, in the CLI
(`nb mv <id> <dest> [-l <src>]`) and in both TUIs (press `m` in the Tasks pane).
The destination list is auto-created if it doesn't exist (like `add`). Because
task IDs are per-list, a moved task receives a fresh ID in the destination but
keeps its title, done state, tags, notes, and original `Created` time.

## 2. Goals / non-goals

**Goals**
- `MoveTask` use case in the domain `Service`, used by all three front-ends.
- CLI `mv` command (alias `move`).
- A move action in the tview and Bubble Tea TUIs.

**Non-goals**
- Moving multiple tasks in one call (single id per CLI invocation; possible later).
- Reordering within a list, or choosing the insertion position (appended to dest).
- Preserving the source ID in the destination (IDs are per-list).

## 3. Domain (`internal/todo`)

**`func (l *List) Append(t Task) *Task`** — appends an existing task to the list:
assigns `t.ID = l.NextID` (lazy-init `NextID` to 1 if zero), increments `NextID`,
sets `t.Updated = time.Now()`, appends, and returns a pointer to the stored task.
Other fields (Title, Done, Tags, Notes, Created) are preserved as given.

**`func (s *Service) MoveTask(srcList string, id int, destList string) (Task, error)`**:
1. `src, err := NormalizeListName(srcList)`; `dest, err := NormalizeListName(destList)` (return on error).
2. Load `src` (propagate `ErrListNotFound`); `t, err := srcList.Get(id)` (propagate `ErrTaskNotFound`). Copy the task value.
3. If `src == dest`: return the copied task, `nil` (no-op).
4. `dl, err := s.loadOrCreate(dest)` — **auto-creates the destination** (same helper `AddTask` uses).
5. `added := dl.Append(taskCopy)`.
6. `srcList.Remove(id)`.
7. `s.repo.Save(dl)` **then** `s.repo.Save(srcList)` — dest first, so a failure between saves leaves a recoverable duplicate, never a lost task.
8. Return `*added` (the moved task with its new destination ID).

Errors: invalid name → `ErrInvalidName`; missing source list → `ErrListNotFound`;
missing task → `ErrTaskNotFound`.

## 4. CLI (`internal/adapter/cli`)

- New `mv.go`: `nb mv <id> <dest>` with alias `move`, `cobra.ExactArgs(2)`.
  - `-l/--list` = source list (default via `resolveList`: `inbox` / `$NB_LIST`).
  - `<dest>` is the second positional arg.
  - Parse the id (reuse `parseIDs` on the single id arg, or `strconv`); call
    `svc.MoveTask(resolveList(list), id, dest)`; on success print
    `moved %q to %s as #%d` (title, dest, new id).
- Registered in `NewRootCmd` alongside the other task commands.

## 5. TUIs

Both adapters add a move action in the **Tasks pane**:

- **tview** (`internal/adapter/tui`): key `m` opens a single-field form
  ("Move to list") via the existing `showModalForm` helper; submit calls
  `svc.MoveTask(currentList, taskID, destName)` then refreshes panes; Esc cancels.
  Footer/keys updated to include `m`.
- **Bubble Tea** (`internal/adapter/bubbletui`): a new `modeMoveTask`; key `m`
  (Tasks pane) → `openMoveTask` (single `textinput` "Move to list"); `submitForm`
  gains a `modeMoveTask` case calling `MoveTask`; the Tasks footer gains `m move`.
  `formTitle` returns "Move task". Validation: empty destination keeps the form
  open (no-op); a Service error closes it and sets `status`.

Both reuse each TUI's existing modal/form infrastructure; no new mechanisms.

## 6. Testing

- **domain:** `List.Append` (assigns next id, increments NextID, preserves
  fields, bumps Updated); `Service.MoveTask` (task gone from src, present in dest
  with a new id and preserved Title/Done/Tags/Notes/Created; dest auto-created;
  `src == dest` no-op; missing source list → `ErrListNotFound`; missing task →
  `ErrTaskNotFound`).
- **cli:** `mv` moves the task, prints the new id, and auto-creates the dest.
- **tview:** pressing `m` opens the move form; submitting calls `MoveTask` (driven
  through the form's input handler / `showModalForm`, as the existing form tests do).
- **bubble:** `m` enters `modeMoveTask`; `submitForm` in that mode moves via the
  Service (driven through `Update`/`updateForm`, as the existing form tests do).

## 7. Files

- Modify: `internal/todo/list.go` (`Append`), `internal/todo/service.go` (`MoveTask`),
  `internal/todo/list_tags_test.go` or a new test file, `internal/todo/service_*_test.go`.
- Create: `internal/adapter/cli/mv.go`; modify `cli.go` (register); test `cli_mv_test.go`.
- Modify: `internal/adapter/tui/forms.go` + `app.go` (move form + `m` key + footer);
  test additions.
- Modify: `internal/adapter/bubbletui/model.go` (mode const), `forms.go`
  (`openMoveTask`, `submitForm` case, `formTitle`), `update.go` (`m` key),
  `view.go` (footer hint); test additions.
