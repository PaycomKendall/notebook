# Todo CLI + TUI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go task tracker with a full Cobra CLI and an interactive tview TUI over one shared, dependency-free domain core.

**Architecture:** Hexagonal (ports & adapters). A core package `internal/todo` holds the domain (`Task`, `List`), a `ListRepository` port, and an application `Service`. Adapters: `jsonstore` (driven, JSON-file persistence), `cli` (driving, Cobra), `tui` (driving, tview). `cmd/todo/main.go` is the composition root that wires them.

**Tech Stack:** Go 1.22+, `github.com/spf13/cobra`, `github.com/rivo/tview`, `github.com/gdamore/tcell/v2`, stdlib `encoding/json`.

## Global Constraints

- **Module path:** `github.com/kendallowen/todo`.
- **Go version floor:** 1.22 (in `go.mod`).
- **Dependency rule:** `internal/todo` imports no other project package and no `cobra`/`tview`/`tcell`. `jsonstore` imports only `internal/todo` + stdlib. `cli`/`tui` import `internal/todo` (+ their framework). Only `cmd/todo/main.go` imports adapters.
- **Storage:** one pretty-printed JSON file per list at `<dir>/<name>.json`; dir resolved `TODO_DIR` → `$XDG_DATA_HOME/todo` → `~/.local/share/todo`. Writes are atomic (temp file + rename).
- **Task fields:** `ID, Title, Done, Tags, Notes, Created, Updated`. No priority/due date.
- **IDs:** per-list, monotonic via `List.NextID`, stable, never reused.
- **List names:** lowercased, must match `^[a-z0-9._-]+$`.
- **Default list (adapters only):** `-l` flag → `$TODO_LIST` → `inbox`.
- **TDD:** every code change is test-first. Commit after each task.
- **Sentinel errors:** `ErrListNotFound`, `ErrListExists`, `ErrTaskNotFound`, `ErrEmptyTitle`, `ErrInvalidName`.

**Prerequisites (one-time, not a task):** `brew install go`; ensure `~/go/bin` is on `PATH` (for `go install`). Verify with `go version`.

---

### Task 1: Module bootstrap + domain types + `List.Add`

**Files:**
- Create: `go.mod`
- Create: `internal/todo/errors.go`
- Create: `internal/todo/task.go`
- Create: `internal/todo/list.go`
- Test: `internal/todo/list_test.go`

**Interfaces:**
- Produces: `todo.Task` struct; `todo.List` struct `{Name string; NextID int; Tasks []Task}`; `func (l *List) Add(title string) (*Task, error)`; sentinel errors; `func normalizeTag(string) string`; `func ValidateListName(string) error`.

- [ ] **Step 1: Initialize the module**

Run:
```bash
go mod init github.com/kendallowen/todo
```
Expected: creates `go.mod` with `module github.com/kendallowen/todo` and `go 1.22` (or newer).

- [ ] **Step 2: Create the sentinel errors**

`internal/todo/errors.go`:
```go
package todo

import "errors"

var (
	ErrListNotFound = errors.New("list not found")
	ErrListExists   = errors.New("list already exists")
	ErrTaskNotFound = errors.New("task not found")
	ErrEmptyTitle   = errors.New("task title must not be empty")
	ErrInvalidName  = errors.New("invalid list name")
)
```

- [ ] **Step 3: Create task types and helpers**

`internal/todo/task.go`:
```go
package todo

import (
	"regexp"
	"strings"
	"time"
)

// Task is a single todo item. IDs are stable within a List.
type Task struct {
	ID      int
	Title   string
	Done    bool
	Tags    []string
	Notes   string
	Created time.Time
	Updated time.Time
}

var listNameRe = regexp.MustCompile(`^[a-z0-9._-]+$`)

// normalizeTag trims surrounding space and lowercases a tag.
func normalizeTag(tag string) string {
	return strings.ToLower(strings.TrimSpace(tag))
}

// ValidateListName ensures a name is safe to use as a filename stem.
func ValidateListName(name string) error {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "" || !listNameRe.MatchString(n) {
		return ErrInvalidName
	}
	return nil
}
```

- [ ] **Step 4: Write the failing test for `List.Add`**

`internal/todo/list_test.go`:
```go
package todo

import "testing"

func TestAddAssignsMonotonicIDs(t *testing.T) {
	l := &List{Name: "inbox"}

	a, err := l.Add("first")
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if a.ID != 1 {
		t.Errorf("first task ID = %d, want 1", a.ID)
	}

	b, err := l.Add("second")
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if b.ID != 2 {
		t.Errorf("second task ID = %d, want 2", b.ID)
	}
	if l.NextID != 3 {
		t.Errorf("NextID = %d, want 3", l.NextID)
	}
	if len(l.Tasks) != 2 {
		t.Errorf("len(Tasks) = %d, want 2", len(l.Tasks))
	}
	if a.Created.IsZero() || a.Updated.IsZero() {
		t.Error("Created/Updated timestamps should be set")
	}
}

func TestAddRejectsEmptyTitle(t *testing.T) {
	l := &List{Name: "inbox"}
	if _, err := l.Add("   "); err != ErrEmptyTitle {
		t.Errorf("Add(blank) err = %v, want ErrEmptyTitle", err)
	}
	if len(l.Tasks) != 0 {
		t.Errorf("blank title should not append; len = %d", len(l.Tasks))
	}
}
```

- [ ] **Step 5: Run the test to verify it fails**

Run: `go test ./internal/todo/`
Expected: FAIL — `List` / `Add` undefined (build error).

- [ ] **Step 6: Implement `List` and `Add`**

`internal/todo/list.go`:
```go
package todo

import (
	"strings"
	"time"
)

// List is an aggregate of tasks persisted as a single file.
type List struct {
	Name   string
	NextID int
	Tasks  []Task
}

// Add appends a new task with a stable, monotonic ID.
func (l *List) Add(title string) (*Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrEmptyTitle
	}
	if l.NextID == 0 {
		l.NextID = 1
	}
	now := time.Now()
	t := Task{ID: l.NextID, Title: title, Created: now, Updated: now}
	l.NextID++
	l.Tasks = append(l.Tasks, t)
	return &l.Tasks[len(l.Tasks)-1], nil
}

// index returns the slice position of id, or -1 if absent.
func (l *List) index(id int) int {
	for i := range l.Tasks {
		if l.Tasks[i].ID == id {
			return i
		}
	}
	return -1
}
```

- [ ] **Step 7: Run the test to verify it passes**

Run: `go test ./internal/todo/`
Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add go.mod internal/todo/
git commit -m "feat(domain): module bootstrap, Task/List types, List.Add"
```

---

### Task 2: `List` read / done / remove methods

**Files:**
- Modify: `internal/todo/list.go`
- Test: `internal/todo/list_methods_test.go`

**Interfaces:**
- Consumes: `List`, `List.index`, `ErrTaskNotFound`.
- Produces: `func (l *List) Get(id int) (*Task, error)`; `func (l *List) Toggle(id int) error`; `func (l *List) SetDone(id int, done bool) error`; `func (l *List) Remove(id int) error`.

- [ ] **Step 1: Write the failing tests**

`internal/todo/list_methods_test.go`:
```go
package todo

import "testing"

func seeded(t *testing.T) *List {
	t.Helper()
	l := &List{Name: "inbox"}
	if _, err := l.Add("alpha"); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Add("beta"); err != nil {
		t.Fatal(err)
	}
	return l
}

func TestToggleFlipsDone(t *testing.T) {
	l := seeded(t)
	if err := l.Toggle(1); err != nil {
		t.Fatal(err)
	}
	got, _ := l.Get(1)
	if !got.Done {
		t.Error("Toggle should set Done=true")
	}
	if err := l.Toggle(1); err != nil {
		t.Fatal(err)
	}
	got, _ = l.Get(1)
	if got.Done {
		t.Error("second Toggle should set Done=false")
	}
}

func TestSetDoneAndMissingID(t *testing.T) {
	l := seeded(t)
	if err := l.SetDone(2, true); err != nil {
		t.Fatal(err)
	}
	got, _ := l.Get(2)
	if !got.Done {
		t.Error("SetDone(true) failed")
	}
	if err := l.Toggle(99); err != ErrTaskNotFound {
		t.Errorf("Toggle(missing) = %v, want ErrTaskNotFound", err)
	}
	if _, err := l.Get(99); err != ErrTaskNotFound {
		t.Errorf("Get(missing) = %v, want ErrTaskNotFound", err)
	}
}

func TestRemove(t *testing.T) {
	l := seeded(t)
	if err := l.Remove(1); err != nil {
		t.Fatal(err)
	}
	if len(l.Tasks) != 1 {
		t.Fatalf("len after remove = %d, want 1", len(l.Tasks))
	}
	if _, err := l.Get(1); err != ErrTaskNotFound {
		t.Error("removed task should be gone")
	}
	if err := l.Remove(1); err != ErrTaskNotFound {
		t.Errorf("Remove(missing) = %v, want ErrTaskNotFound", err)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/todo/`
Expected: FAIL — `Get`/`Toggle`/`SetDone`/`Remove` undefined.

- [ ] **Step 3: Implement the methods**

Append to `internal/todo/list.go`:
```go
// Get returns a pointer to the task with the given id.
func (l *List) Get(id int) (*Task, error) {
	i := l.index(id)
	if i < 0 {
		return nil, ErrTaskNotFound
	}
	return &l.Tasks[i], nil
}

// Toggle flips the done state of a task.
func (l *List) Toggle(id int) error {
	i := l.index(id)
	if i < 0 {
		return ErrTaskNotFound
	}
	l.Tasks[i].Done = !l.Tasks[i].Done
	l.Tasks[i].Updated = time.Now()
	return nil
}

// SetDone sets the done state explicitly.
func (l *List) SetDone(id int, done bool) error {
	i := l.index(id)
	if i < 0 {
		return ErrTaskNotFound
	}
	l.Tasks[i].Done = done
	l.Tasks[i].Updated = time.Now()
	return nil
}

// Remove deletes a task by id.
func (l *List) Remove(id int) error {
	i := l.index(id)
	if i < 0 {
		return ErrTaskNotFound
	}
	l.Tasks = append(l.Tasks[:i], l.Tasks[i+1:]...)
	return nil
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/todo/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/todo/
git commit -m "feat(domain): List Get/Toggle/SetDone/Remove"
```

---

### Task 3: `List` edit + tag methods

**Files:**
- Modify: `internal/todo/list.go`
- Test: `internal/todo/list_tags_test.go`

**Interfaces:**
- Produces: `func (l *List) SetTitle(id int, title string) error`; `func (l *List) SetNotes(id int, notes string) error`; `func (l *List) AddTag(id int, tag string) error`; `func (l *List) RemoveTag(id int, tag string) error`.

- [ ] **Step 1: Write the failing tests**

`internal/todo/list_tags_test.go`:
```go
package todo

import (
	"reflect"
	"testing"
)

func TestSetTitleAndNotes(t *testing.T) {
	l := seeded(t)
	if err := l.SetTitle(1, "  renamed  "); err != nil {
		t.Fatal(err)
	}
	got, _ := l.Get(1)
	if got.Title != "renamed" {
		t.Errorf("Title = %q, want %q", got.Title, "renamed")
	}
	if err := l.SetTitle(1, "   "); err != ErrEmptyTitle {
		t.Errorf("SetTitle(blank) = %v, want ErrEmptyTitle", err)
	}
	if err := l.SetNotes(1, "buy milk"); err != nil {
		t.Fatal(err)
	}
	got, _ = l.Get(1)
	if got.Notes != "buy milk" {
		t.Errorf("Notes = %q", got.Notes)
	}
}

func TestTagAddNormalizesAndDedupes(t *testing.T) {
	l := seeded(t)
	for _, tag := range []string{"Urgent", " urgent ", "home"} {
		if err := l.AddTag(1, tag); err != nil {
			t.Fatal(err)
		}
	}
	got, _ := l.Get(1)
	if !reflect.DeepEqual(got.Tags, []string{"urgent", "home"}) {
		t.Errorf("Tags = %v, want [urgent home]", got.Tags)
	}
	if err := l.AddTag(1, "  "); err != nil {
		t.Fatalf("blank tag should be a no-op, got %v", err)
	}
	got, _ = l.Get(1)
	if len(got.Tags) != 2 {
		t.Errorf("blank tag changed count: %v", got.Tags)
	}
}

func TestRemoveTag(t *testing.T) {
	l := seeded(t)
	_ = l.AddTag(1, "home")
	_ = l.AddTag(1, "urgent")
	if err := l.RemoveTag(1, "HOME"); err != nil {
		t.Fatal(err)
	}
	got, _ := l.Get(1)
	if !reflect.DeepEqual(got.Tags, []string{"urgent"}) {
		t.Errorf("Tags = %v, want [urgent]", got.Tags)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/todo/`
Expected: FAIL — methods undefined.

- [ ] **Step 3: Implement the methods**

Append to `internal/todo/list.go`:
```go
// SetTitle updates a task title (non-empty after trim).
func (l *List) SetTitle(id int, title string) error {
	i := l.index(id)
	if i < 0 {
		return ErrTaskNotFound
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return ErrEmptyTitle
	}
	l.Tasks[i].Title = title
	l.Tasks[i].Updated = time.Now()
	return nil
}

// SetNotes replaces a task's notes.
func (l *List) SetNotes(id int, notes string) error {
	i := l.index(id)
	if i < 0 {
		return ErrTaskNotFound
	}
	l.Tasks[i].Notes = notes
	l.Tasks[i].Updated = time.Now()
	return nil
}

// AddTag adds a normalized tag, ignoring blanks and duplicates.
func (l *List) AddTag(id int, tag string) error {
	i := l.index(id)
	if i < 0 {
		return ErrTaskNotFound
	}
	tag = normalizeTag(tag)
	if tag == "" {
		return nil
	}
	for _, existing := range l.Tasks[i].Tags {
		if existing == tag {
			return nil
		}
	}
	l.Tasks[i].Tags = append(l.Tasks[i].Tags, tag)
	l.Tasks[i].Updated = time.Now()
	return nil
}

// RemoveTag removes a normalized tag if present.
func (l *List) RemoveTag(id int, tag string) error {
	i := l.index(id)
	if i < 0 {
		return ErrTaskNotFound
	}
	tag = normalizeTag(tag)
	out := l.Tasks[i].Tags[:0]
	for _, existing := range l.Tasks[i].Tags {
		if existing != tag {
			out = append(out, existing)
		}
	}
	l.Tasks[i].Tags = out
	l.Tasks[i].Updated = time.Now()
	return nil
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/todo/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/todo/
git commit -m "feat(domain): List SetTitle/SetNotes/AddTag/RemoveTag"
```

---

### Task 4: Repository port + `Service.AddTask` (with auto-create)

**Files:**
- Create: `internal/todo/repository.go`
- Create: `internal/todo/service.go`
- Test: `internal/todo/service_test.go`

**Interfaces:**
- Produces: `type ListRepository interface { Names() ([]string, error); Load(name string) (*List, error); Save(list *List) error; Create(name string) (*List, error); Delete(name string) error; Rename(oldName, newName string) error }`; `func NewService(repo ListRepository) *Service`; `func (s *Service) AddTask(list, title string, tags []string, notes string) (Task, error)`.

- [ ] **Step 1: Define the port**

`internal/todo/repository.go`:
```go
package todo

// ListRepository is the persistence port for lists (implemented by adapters).
type ListRepository interface {
	Names() ([]string, error)
	Load(name string) (*List, error)
	Save(list *List) error
	Create(name string) (*List, error)
	Delete(name string) error
	Rename(oldName, newName string) error
}
```

- [ ] **Step 2: Write the failing test (with an in-memory fake repo)**

`internal/todo/service_test.go`:
```go
package todo

import (
	"errors"
	"testing"
)

// fakeRepo is an in-memory ListRepository for service tests.
type fakeRepo struct {
	lists map[string]*List
}

func newFakeRepo() *fakeRepo { return &fakeRepo{lists: map[string]*List{}} }

func clone(l *List) *List {
	cp := *l
	cp.Tasks = append([]Task(nil), l.Tasks...)
	return &cp
}

func (f *fakeRepo) Names() ([]string, error) {
	out := make([]string, 0, len(f.lists))
	for name := range f.lists {
		out = append(out, name)
	}
	return out, nil
}
func (f *fakeRepo) Load(name string) (*List, error) {
	l, ok := f.lists[name]
	if !ok {
		return nil, ErrListNotFound
	}
	return clone(l), nil
}
func (f *fakeRepo) Save(l *List) error { f.lists[l.Name] = clone(l); return nil }
func (f *fakeRepo) Create(name string) (*List, error) {
	if _, ok := f.lists[name]; ok {
		return nil, ErrListExists
	}
	l := &List{Name: name, NextID: 1}
	f.lists[name] = clone(l)
	return l, nil
}
func (f *fakeRepo) Delete(name string) error {
	if _, ok := f.lists[name]; !ok {
		return ErrListNotFound
	}
	delete(f.lists, name)
	return nil
}
func (f *fakeRepo) Rename(oldName, newName string) error {
	l, ok := f.lists[oldName]
	if !ok {
		return ErrListNotFound
	}
	if _, ok := f.lists[newName]; ok {
		return ErrListExists
	}
	l.Name = newName
	f.lists[newName] = l
	delete(f.lists, oldName)
	return nil
}

func TestAddTaskAutoCreatesList(t *testing.T) {
	svc := NewService(newFakeRepo())
	task, err := svc.AddTask("groceries", "milk", []string{"Store"}, "2%")
	if err != nil {
		t.Fatalf("AddTask error: %v", err)
	}
	if task.ID != 1 || task.Title != "milk" {
		t.Errorf("task = %+v", task)
	}
	if len(task.Tags) != 1 || task.Tags[0] != "store" {
		t.Errorf("tags = %v, want [store]", task.Tags)
	}
	if task.Notes != "2%" {
		t.Errorf("notes = %q", task.Notes)
	}
	l, err := svc.GetList("groceries")
	if err != nil {
		t.Fatalf("list should have been created: %v", err)
	}
	if len(l.Tasks) != 1 {
		t.Errorf("persisted task count = %d, want 1", len(l.Tasks))
	}
}

func TestAddTaskRejectsBadListName(t *testing.T) {
	svc := NewService(newFakeRepo())
	if _, err := svc.AddTask("Bad Name", "x", nil, ""); !errors.Is(err, ErrInvalidName) {
		t.Errorf("err = %v, want ErrInvalidName", err)
	}
}
```

(Note: `GetList` is implemented in Task 6 — if running tasks strictly in order, temporarily inline a `svc.repo.Load` check; the final test above is correct once Task 6 lands. To keep this task self-contained, replace the `svc.GetList("groceries")` block with `svc.repo.Load("groceries")` until Task 6, then restore. Simplest path: implement the tiny `GetList` shim in Step 4 below.)

- [ ] **Step 3: Run the test to verify it fails**

Run: `go test ./internal/todo/`
Expected: FAIL — `NewService`/`AddTask`/`GetList` undefined.

- [ ] **Step 4: Implement the service core + AddTask + GetList shim**

`internal/todo/service.go`:
```go
package todo

import "errors"

// Service exposes the application use cases both front-ends call.
type Service struct {
	repo ListRepository
}

func NewService(repo ListRepository) *Service { return &Service{repo: repo} }

// loadOrCreate returns the named list, creating it if missing.
func (s *Service) loadOrCreate(name string) (*List, error) {
	l, err := s.repo.Load(name)
	if errors.Is(err, ErrListNotFound) {
		return s.repo.Create(name)
	}
	return l, err
}

// AddTask adds a task to a list, creating the list if it does not exist.
func (s *Service) AddTask(list, title string, tags []string, notes string) (Task, error) {
	if err := ValidateListName(list); err != nil {
		return Task{}, err
	}
	l, err := s.loadOrCreate(list)
	if err != nil {
		return Task{}, err
	}
	t, err := l.Add(title)
	if err != nil {
		return Task{}, err
	}
	for _, tag := range tags {
		_ = l.AddTag(t.ID, tag)
	}
	if notes != "" {
		_ = l.SetNotes(t.ID, notes)
	}
	if err := s.repo.Save(l); err != nil {
		return Task{}, err
	}
	return *t, nil
}

// GetList returns a list by name (ErrListNotFound if absent).
func (s *Service) GetList(name string) (*List, error) { return s.repo.Load(name) }
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./internal/todo/`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/todo/
git commit -m "feat(core): ListRepository port + Service.AddTask with auto-create"
```

---

### Task 5: Service mutations on existing lists

**Files:**
- Modify: `internal/todo/service.go`
- Test: `internal/todo/service_mutate_test.go`

**Interfaces:**
- Produces: `func (s *Service) ToggleTask(list string, id int) error`; `SetTaskDone(list string, id int, done bool) error`; `RemoveTask(list string, id int) error`; `EditTask(list string, id int, title, notes *string) error`; `AddTaskTag(list string, id int, tag string) error`; `RemoveTaskTag(list string, id int, tag string) error`. Missing list → `ErrListNotFound`.

- [ ] **Step 1: Write the failing tests**

`internal/todo/service_mutate_test.go`:
```go
package todo

import "testing"

func svcWithTask(t *testing.T) *Service {
	t.Helper()
	svc := NewService(newFakeRepo())
	if _, err := svc.AddTask("work", "ship it", nil, ""); err != nil {
		t.Fatal(err)
	}
	return svc
}

func TestToggleTaskPersists(t *testing.T) {
	svc := svcWithTask(t)
	if err := svc.ToggleTask("work", 1); err != nil {
		t.Fatal(err)
	}
	l, _ := svc.GetList("work")
	if !l.Tasks[0].Done {
		t.Error("toggle did not persist")
	}
}

func TestEditTaskPartialUpdate(t *testing.T) {
	svc := svcWithTask(t)
	newNotes := "with details"
	if err := svc.EditTask("work", 1, nil, &newNotes); err != nil {
		t.Fatal(err)
	}
	l, _ := svc.GetList("work")
	if l.Tasks[0].Title != "ship it" {
		t.Error("title should be unchanged when nil")
	}
	if l.Tasks[0].Notes != "with details" {
		t.Error("notes not updated")
	}
}

func TestMutationOnMissingList(t *testing.T) {
	svc := NewService(newFakeRepo())
	if err := svc.ToggleTask("nope", 1); err != ErrListNotFound {
		t.Errorf("err = %v, want ErrListNotFound", err)
	}
}

func TestTaskTagThroughService(t *testing.T) {
	svc := svcWithTask(t)
	if err := svc.AddTaskTag("work", 1, "Urgent"); err != nil {
		t.Fatal(err)
	}
	l, _ := svc.GetList("work")
	if len(l.Tasks[0].Tags) != 1 || l.Tasks[0].Tags[0] != "urgent" {
		t.Errorf("tags = %v", l.Tasks[0].Tags)
	}
	if err := svc.RemoveTaskTag("work", 1, "urgent"); err != nil {
		t.Fatal(err)
	}
	l, _ = svc.GetList("work")
	if len(l.Tasks[0].Tags) != 0 {
		t.Errorf("tags after remove = %v", l.Tasks[0].Tags)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/todo/`
Expected: FAIL — methods undefined.

- [ ] **Step 3: Implement the mutation helpers**

Append to `internal/todo/service.go`:
```go
// mutate loads an existing list, applies fn, and saves it.
func (s *Service) mutate(list string, fn func(*List) error) error {
	l, err := s.repo.Load(list)
	if err != nil {
		return err
	}
	if err := fn(l); err != nil {
		return err
	}
	return s.repo.Save(l)
}

func (s *Service) ToggleTask(list string, id int) error {
	return s.mutate(list, func(l *List) error { return l.Toggle(id) })
}

func (s *Service) SetTaskDone(list string, id int, done bool) error {
	return s.mutate(list, func(l *List) error { return l.SetDone(id, done) })
}

func (s *Service) RemoveTask(list string, id int) error {
	return s.mutate(list, func(l *List) error { return l.Remove(id) })
}

func (s *Service) EditTask(list string, id int, title, notes *string) error {
	return s.mutate(list, func(l *List) error {
		if title != nil {
			if err := l.SetTitle(id, *title); err != nil {
				return err
			}
		}
		if notes != nil {
			if err := l.SetNotes(id, *notes); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) AddTaskTag(list string, id int, tag string) error {
	return s.mutate(list, func(l *List) error { return l.AddTag(id, tag) })
}

func (s *Service) RemoveTaskTag(list string, id int, tag string) error {
	return s.mutate(list, func(l *List) error { return l.RemoveTag(id, tag) })
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/todo/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/todo/
git commit -m "feat(core): Service task mutations on existing lists"
```

---

### Task 6: Service list management

**Files:**
- Modify: `internal/todo/service.go`
- Test: `internal/todo/service_lists_test.go`

**Interfaces:**
- Produces: `func (s *Service) ListNames() ([]string, error)`; `CreateList(name string) error`; `DeleteList(name string) error`; `RenameList(old, new string) error`. (`GetList` already exists from Task 4.)

- [ ] **Step 1: Write the failing tests**

`internal/todo/service_lists_test.go`:
```go
package todo

import (
	"errors"
	"testing"
)

func TestCreateListValidatesAndRejectsDup(t *testing.T) {
	svc := NewService(newFakeRepo())
	if err := svc.CreateList("Bad Name"); !errors.Is(err, ErrInvalidName) {
		t.Errorf("err = %v, want ErrInvalidName", err)
	}
	if err := svc.CreateList("work"); err != nil {
		t.Fatal(err)
	}
	if err := svc.CreateList("work"); err != ErrListExists {
		t.Errorf("dup err = %v, want ErrListExists", err)
	}
}

func TestRenameAndDeleteList(t *testing.T) {
	svc := NewService(newFakeRepo())
	_ = svc.CreateList("old")
	if err := svc.RenameList("old", "new"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GetList("new"); err != nil {
		t.Errorf("renamed list missing: %v", err)
	}
	if err := svc.DeleteList("new"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GetList("new"); err != ErrListNotFound {
		t.Error("deleted list should be gone")
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/todo/`
Expected: FAIL — methods undefined.

- [ ] **Step 3: Implement list management**

Append to `internal/todo/service.go`:
```go
func (s *Service) ListNames() ([]string, error) { return s.repo.Names() }

func (s *Service) CreateList(name string) error {
	if err := ValidateListName(name); err != nil {
		return err
	}
	_, err := s.repo.Create(name)
	return err
}

func (s *Service) DeleteList(name string) error { return s.repo.Delete(name) }

func (s *Service) RenameList(old, newName string) error {
	if err := ValidateListName(newName); err != nil {
		return err
	}
	return s.repo.Rename(old, newName)
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/todo/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/todo/
git commit -m "feat(core): Service list management"
```

---

### Task 7: jsonstore — dir resolution, DTO mapping, Save/Load

**Files:**
- Create: `internal/adapter/jsonstore/store.go`
- Test: `internal/adapter/jsonstore/store_test.go`

**Interfaces:**
- Consumes: `todo.List`, `todo.Task`, `todo.ErrListNotFound`.
- Produces: `func New(dir string) (*Store, error)`; `func DefaultDir() (string, error)`; `func (s *Store) Save(l *todo.List) error`; `func (s *Store) Load(name string) (*todo.List, error)`. (`Store` satisfies `todo.ListRepository` after Task 8.)

- [ ] **Step 1: Write the failing tests**

`internal/adapter/jsonstore/store_test.go`:
```go
package jsonstore

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kendallowen/todo/internal/todo"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	l := &todo.List{Name: "work", NextID: 2,
		Tasks: []todo.Task{{ID: 1, Title: "ship", Tags: []string{"urgent"}, Notes: "go"}}}
	if err := s.Save(l); err != nil {
		t.Fatal(err)
	}
	got, err := s.Load("work")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "work" || got.NextID != 2 || len(got.Tasks) != 1 {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
	if got.Tasks[0].Title != "ship" || got.Tasks[0].Tags[0] != "urgent" {
		t.Errorf("task mismatch: %+v", got.Tasks[0])
	}
}

func TestLoadMissingIsErrListNotFound(t *testing.T) {
	s, _ := New(t.TempDir())
	if _, err := s.Load("nope"); err != todo.ErrListNotFound {
		t.Errorf("err = %v, want ErrListNotFound", err)
	}
}

func TestSaveIsAtomicNoTempLeft(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(dir)
	_ = s.Save(&todo.List{Name: "x", NextID: 1})
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("temp file left behind: %s", e.Name())
		}
	}
}

func TestDefaultDirHonorsTODODIR(t *testing.T) {
	t.Setenv("TODO_DIR", "/tmp/custom-todo")
	d, err := DefaultDir()
	if err != nil {
		t.Fatal(err)
	}
	if d != "/tmp/custom-todo" {
		t.Errorf("DefaultDir = %q, want /tmp/custom-todo", d)
	}
}

func TestLoadMalformedJSONErrors(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(dir)
	if err := os.WriteFile(filepath.Join(dir, "broken.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := s.Load("broken")
	if err == nil {
		t.Fatal("expected error loading malformed JSON")
	}
	if err == todo.ErrListNotFound {
		t.Error("malformed JSON should not be reported as not-found")
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/adapter/jsonstore/`
Expected: FAIL — package/symbols undefined.

- [ ] **Step 3: Implement Store with Save/Load**

`internal/adapter/jsonstore/store.go`:
```go
package jsonstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kendallowen/todo/internal/todo"
)

// Store persists lists as one JSON file per list.
type Store struct{ dir string }

// New creates a Store rooted at dir (DefaultDir() if dir == "").
func New(dir string) (*Store, error) {
	if dir == "" {
		d, err := DefaultDir()
		if err != nil {
			return nil, err
		}
		dir = d
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Store{dir: dir}, nil
}

// DefaultDir resolves the data directory: TODO_DIR, then XDG, then ~/.local/share.
func DefaultDir() (string, error) {
	if d := os.Getenv("TODO_DIR"); d != "" {
		return d, nil
	}
	if d := os.Getenv("XDG_DATA_HOME"); d != "" {
		return filepath.Join(d, "todo"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "todo"), nil
}

func (s *Store) path(name string) string { return filepath.Join(s.dir, name+".json") }

type taskDTO struct {
	ID      int       `json:"id"`
	Title   string    `json:"title"`
	Done    bool      `json:"done"`
	Tags    []string  `json:"tags,omitempty"`
	Notes   string    `json:"notes,omitempty"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

type listDTO struct {
	Name   string    `json:"name"`
	NextID int       `json:"next_id"`
	Tasks  []taskDTO `json:"tasks"`
}

func toDTO(l *todo.List) listDTO {
	d := listDTO{Name: l.Name, NextID: l.NextID}
	for _, t := range l.Tasks {
		d.Tasks = append(d.Tasks, taskDTO{
			ID: t.ID, Title: t.Title, Done: t.Done, Tags: t.Tags,
			Notes: t.Notes, Created: t.Created, Updated: t.Updated,
		})
	}
	return d
}

func fromDTO(d listDTO) *todo.List {
	l := &todo.List{Name: d.Name, NextID: d.NextID}
	for _, t := range d.Tasks {
		l.Tasks = append(l.Tasks, todo.Task{
			ID: t.ID, Title: t.Title, Done: t.Done, Tags: t.Tags,
			Notes: t.Notes, Created: t.Created, Updated: t.Updated,
		})
	}
	return l
}

// Save writes a list atomically (temp file + rename).
func (s *Store) Save(l *todo.List) error {
	data, err := json.MarshalIndent(toDTO(l), "", "  ")
	if err != nil {
		return err
	}
	path := s.path(l.Name)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Load reads a list by name.
func (s *Store) Load(name string) (*todo.List, error) {
	data, err := os.ReadFile(s.path(name))
	if errors.Is(err, os.ErrNotExist) {
		return nil, todo.ErrListNotFound
	}
	if err != nil {
		return nil, err
	}
	var d listDTO
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("parse list %q: %w", name, err)
	}
	return fromDTO(d), nil
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/adapter/jsonstore/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/adapter/jsonstore/
git commit -m "feat(jsonstore): dir resolution, DTO mapping, atomic Save/Load"
```

---

### Task 8: jsonstore — Names/Create/Delete/Rename + error mapping

**Files:**
- Modify: `internal/adapter/jsonstore/store.go`
- Test: `internal/adapter/jsonstore/store_manage_test.go`

**Interfaces:**
- Produces: `func (s *Store) Names() ([]string, error)`; `Create(name string) (*todo.List, error)`; `Delete(name string) error`; `Rename(oldName, newName string) error`. After this task `*Store` satisfies `todo.ListRepository`.

- [ ] **Step 1: Write the failing tests**

`internal/adapter/jsonstore/store_manage_test.go`:
```go
package jsonstore

import (
	"testing"

	"github.com/kendallowen/todo/internal/todo"
)

func TestNamesSortedAndCreate(t *testing.T) {
	s, _ := New(t.TempDir())
	if _, err := s.Create("work"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create("home"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create("work"); err != todo.ErrListExists {
		t.Errorf("dup Create = %v, want ErrListExists", err)
	}
	names, err := s.Names()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 || names[0] != "home" || names[1] != "work" {
		t.Errorf("Names = %v, want [home work]", names)
	}
}

func TestDeleteAndRename(t *testing.T) {
	s, _ := New(t.TempDir())
	_, _ = s.Create("old")
	if err := s.Rename("old", "new"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Load("new"); err != nil {
		t.Errorf("renamed list missing: %v", err)
	}
	if _, err := s.Load("old"); err != todo.ErrListNotFound {
		t.Error("old name should be gone")
	}
	if err := s.Delete("new"); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete("new"); err != todo.ErrListNotFound {
		t.Errorf("Delete(missing) = %v, want ErrListNotFound", err)
	}
}

// Compile-time check that Store implements the port.
var _ todo.ListRepository = (*Store)(nil)
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/adapter/jsonstore/`
Expected: FAIL — methods undefined; interface assertion fails to compile.

- [ ] **Step 3: Implement the management methods**

Append to `internal/adapter/jsonstore/store.go` (add `"sort"` and `"strings"` to imports):
```go
// Names returns the stems of all *.json list files, sorted.
func (s *Store) Names() ([]string, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".json") {
			names = append(names, strings.TrimSuffix(e.Name(), ".json"))
		}
	}
	sort.Strings(names)
	return names, nil
}

// Create makes a new empty list, erroring if it already exists.
func (s *Store) Create(name string) (*todo.List, error) {
	if err := todo.ValidateListName(name); err != nil {
		return nil, err
	}
	if _, err := os.Stat(s.path(name)); err == nil {
		return nil, todo.ErrListExists
	}
	l := &todo.List{Name: name, NextID: 1}
	if err := s.Save(l); err != nil {
		return nil, err
	}
	return l, nil
}

// Delete removes a list file.
func (s *Store) Delete(name string) error {
	err := os.Remove(s.path(name))
	if errors.Is(err, os.ErrNotExist) {
		return todo.ErrListNotFound
	}
	return err
}

// Rename moves a list to a new name (and updates its stored Name).
func (s *Store) Rename(oldName, newName string) error {
	if err := todo.ValidateListName(newName); err != nil {
		return err
	}
	if _, err := os.Stat(s.path(newName)); err == nil {
		return todo.ErrListExists
	}
	l, err := s.Load(oldName)
	if err != nil {
		return err
	}
	l.Name = newName
	if err := s.Save(l); err != nil {
		return err
	}
	return s.Delete(oldName)
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/adapter/jsonstore/`
Expected: PASS (including the `var _ todo.ListRepository` assertion compiling).

- [ ] **Step 5: Commit**

```bash
git add internal/adapter/jsonstore/
git commit -m "feat(jsonstore): Names/Create/Delete/Rename, satisfies ListRepository"
```

---

### Task 9: CLI scaffolding + `add` command

**Files:**
- Create: `internal/adapter/cli/cli.go`
- Create: `internal/adapter/cli/add.go`
- Test: `internal/adapter/cli/cli_test.go`

**Interfaces:**
- Consumes: `*todo.Service`.
- Produces: `func NewRootCmd(svc *todo.Service, launchTUI func() error) *cobra.Command`; `func resolveList(flag string) string`; helper `newAddCmd(svc *todo.Service) *cobra.Command`.

- [ ] **Step 1: Add the Cobra dependency**

Run:
```bash
go get github.com/spf13/cobra@latest
```
Expected: updates `go.mod`/`go.sum`.

- [ ] **Step 2: Write the failing test**

`internal/adapter/cli/cli_test.go`:
```go
package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kendallowen/todo/internal/adapter/jsonstore"
	"github.com/kendallowen/todo/internal/todo"
)

// newTestCmd builds a root command backed by a temp-dir store.
func newTestCmd(t *testing.T) (*todo.Service, func(args ...string) (string, error)) {
	t.Helper()
	store, err := jsonstore.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	svc := todo.NewService(store)
	run := func(args ...string) (string, error) {
		cmd := NewRootCmd(svc, func() error { return nil })
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs(args)
		err := cmd.Execute()
		return buf.String(), err
	}
	return svc, run
}

func TestAddCommand(t *testing.T) {
	svc, run := newTestCmd(t)
	out, err := run("add", "buy milk", "-l", "groceries", "-t", "store")
	if err != nil {
		t.Fatalf("add error: %v", err)
	}
	if !strings.Contains(out, "#1") || !strings.Contains(out, "buy milk") {
		t.Errorf("unexpected output: %q", out)
	}
	l, err := svc.GetList("groceries")
	if err != nil {
		t.Fatalf("list not created: %v", err)
	}
	if len(l.Tasks) != 1 || l.Tasks[0].Tags[0] != "store" {
		t.Errorf("task not persisted correctly: %+v", l.Tasks)
	}
}
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `go test ./internal/adapter/cli/`
Expected: FAIL — `NewRootCmd` undefined.

- [ ] **Step 4: Implement the root command + resolveList**

`internal/adapter/cli/cli.go`:
```go
package cli

import (
	"os"

	"github.com/kendallowen/todo/internal/todo"
	"github.com/spf13/cobra"
)

// resolveList applies the default-list rule: flag, then $TODO_LIST, then "inbox".
func resolveList(flag string) string {
	if flag != "" {
		return flag
	}
	if env := os.Getenv("TODO_LIST"); env != "" {
		return env
	}
	return "inbox"
}

// NewRootCmd builds the command tree. launchTUI runs the interactive UI.
func NewRootCmd(svc *todo.Service, launchTUI func() error) *cobra.Command {
	root := &cobra.Command{
		Use:           "todo",
		Short:         "A CLI + TUI task tracker",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return launchTUI()
		},
	}
	tui := &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive TUI",
		RunE:  func(cmd *cobra.Command, args []string) error { return launchTUI() },
	}
	root.AddCommand(tui)
	root.AddCommand(newAddCmd(svc))
	return root
}
```

`internal/adapter/cli/add.go`:
```go
package cli

import (
	"fmt"
	"strings"

	"github.com/kendallowen/todo/internal/todo"
	"github.com/spf13/cobra"
)

func newAddCmd(svc *todo.Service) *cobra.Command {
	var list, note string
	var tags []string
	cmd := &cobra.Command{
		Use:   "add <title>...",
		Short: "Add a task",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := resolveList(list)
			title := strings.Join(args, " ")
			task, err := svc.AddTask(name, title, tags, note)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Added [%s] #%d: %s\n", name, task.ID, task.Title)
			return nil
		},
	}
	cmd.Flags().StringVarP(&list, "list", "l", "", "list name (default: inbox or $TODO_LIST)")
	cmd.Flags().StringSliceVarP(&tags, "tag", "t", nil, "tag (repeatable or comma-separated)")
	cmd.Flags().StringVarP(&note, "note", "n", "", "note text")
	return cmd
}
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./internal/adapter/cli/`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum internal/adapter/cli/
git commit -m "feat(cli): root command, resolveList, add"
```

---

### Task 10: CLI `ls`, `done`, `undone`, `rm`

**Files:**
- Create: `internal/adapter/cli/ls.go`
- Create: `internal/adapter/cli/done.go`
- Create: `internal/adapter/cli/rm.go`
- Modify: `internal/adapter/cli/cli.go` (register the new commands)
- Test: `internal/adapter/cli/cli_tasks_test.go`

**Interfaces:**
- Produces: `newLsCmd`, `newDoneCmd`, `newUndoneCmd`, `newRmCmd` (all `func(*todo.Service) *cobra.Command`); shared `parseIDs([]string) ([]int, error)`.

- [ ] **Step 1: Write the failing tests**

`internal/adapter/cli/cli_tasks_test.go`:
```go
package cli

import (
	"strings"
	"testing"
)

func TestLsShowsTasks(t *testing.T) {
	_, run := newTestCmd(t)
	_, _ = run("add", "alpha", "-l", "work")
	_, _ = run("add", "beta", "-l", "work")
	out, err := run("ls", "-l", "work")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
		t.Errorf("ls missing tasks: %q", out)
	}
	if !strings.Contains(out, "[ ]") {
		t.Errorf("ls missing open checkbox: %q", out)
	}
}

func TestDoneAndUndone(t *testing.T) {
	svc, run := newTestCmd(t)
	_, _ = run("add", "alpha", "-l", "work")
	if _, err := run("done", "1", "-l", "work"); err != nil {
		t.Fatal(err)
	}
	l, _ := svc.GetList("work")
	if !l.Tasks[0].Done {
		t.Error("done did not mark task")
	}
	if _, err := run("undone", "1", "-l", "work"); err != nil {
		t.Fatal(err)
	}
	l, _ = svc.GetList("work")
	if l.Tasks[0].Done {
		t.Error("undone did not clear task")
	}
}

func TestRmDeletesTask(t *testing.T) {
	svc, run := newTestCmd(t)
	_, _ = run("add", "alpha", "-l", "work")
	if _, err := run("rm", "1", "-l", "work"); err != nil {
		t.Fatal(err)
	}
	l, _ := svc.GetList("work")
	if len(l.Tasks) != 0 {
		t.Errorf("task not removed: %+v", l.Tasks)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/adapter/cli/`
Expected: FAIL — `ls`/`done`/`undone`/`rm` unknown commands.

- [ ] **Step 3: Implement `ls`**

`internal/adapter/cli/ls.go`:
```go
package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kendallowen/todo/internal/todo"
	"github.com/spf13/cobra"
)

// parseIDs converts string args to ints.
func parseIDs(args []string) ([]int, error) {
	ids := make([]int, 0, len(args))
	for _, a := range args {
		n, err := strconv.Atoi(a)
		if err != nil {
			return nil, fmt.Errorf("invalid id %q", a)
		}
		ids = append(ids, n)
	}
	return ids, nil
}

func formatTask(t todo.Task) string {
	box := "[ ]"
	if t.Done {
		box = "[x]"
	}
	line := fmt.Sprintf("%s #%d %s", box, t.ID, t.Title)
	if len(t.Tags) > 0 {
		line += "  #" + strings.Join(t.Tags, " #")
	}
	return line
}

func newLsCmd(svc *todo.Service) *cobra.Command {
	var list, tag string
	var all, doneOnly, openOnly bool
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			var names []string
			if all {
				n, err := svc.ListNames()
				if err != nil {
					return err
				}
				names = n
			} else {
				names = []string{resolveList(list)}
			}
			out := cmd.OutOrStdout()
			for _, name := range names {
				l, err := svc.GetList(name)
				if err == todo.ErrListNotFound {
					continue
				}
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "%s:\n", name)
				for _, task := range l.Tasks {
					if tag != "" && !hasTag(task, tag) {
						continue
					}
					if doneOnly && !task.Done {
						continue
					}
					if openOnly && task.Done {
						continue
					}
					fmt.Fprintf(out, "  %s\n", formatTask(task))
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&list, "list", "l", "", "list name (default: inbox or $TODO_LIST)")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "show all lists")
	cmd.Flags().StringVarP(&tag, "tag", "t", "", "filter by tag")
	cmd.Flags().BoolVar(&doneOnly, "done", false, "show only done tasks")
	cmd.Flags().BoolVar(&openOnly, "open", false, "show only open tasks")
	return cmd
}

func hasTag(t todo.Task, tag string) bool {
	tag = strings.ToLower(strings.TrimSpace(tag))
	for _, x := range t.Tags {
		if x == tag {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Implement `done` / `undone`**

`internal/adapter/cli/done.go`:
```go
package cli

import (
	"fmt"

	"github.com/kendallowen/todo/internal/todo"
	"github.com/spf13/cobra"
)

func newDoneCmd(svc *todo.Service) *cobra.Command  { return doneLikeCmd(svc, "done", true) }
func newUndoneCmd(svc *todo.Service) *cobra.Command { return doneLikeCmd(svc, "undone", false) }

func doneLikeCmd(svc *todo.Service, use string, done bool) *cobra.Command {
	var list string
	cmd := &cobra.Command{
		Use:   use + " <id>...",
		Short: "Mark task(s) " + use,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(args)
			if err != nil {
				return err
			}
			name := resolveList(list)
			for _, id := range ids {
				if err := svc.SetTaskDone(name, id, done); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s #%d\n", use, id)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&list, "list", "l", "", "list name (default: inbox or $TODO_LIST)")
	return cmd
}
```

- [ ] **Step 5: Implement `rm`**

`internal/adapter/cli/rm.go`:
```go
package cli

import (
	"fmt"

	"github.com/kendallowen/todo/internal/todo"
	"github.com/spf13/cobra"
)

func newRmCmd(svc *todo.Service) *cobra.Command {
	var list string
	cmd := &cobra.Command{
		Use:   "rm <id>...",
		Short: "Delete task(s)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(args)
			if err != nil {
				return err
			}
			name := resolveList(list)
			for _, id := range ids {
				if err := svc.RemoveTask(name, id); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "removed #%d\n", id)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&list, "list", "l", "", "list name (default: inbox or $TODO_LIST)")
	return cmd
}
```

- [ ] **Step 6: Register the new commands**

In `internal/adapter/cli/cli.go`, inside `NewRootCmd`, after `root.AddCommand(newAddCmd(svc))` add:
```go
	root.AddCommand(newLsCmd(svc))
	root.AddCommand(newDoneCmd(svc))
	root.AddCommand(newUndoneCmd(svc))
	root.AddCommand(newRmCmd(svc))
```

- [ ] **Step 7: Run the tests to verify they pass**

Run: `go test ./internal/adapter/cli/`
Expected: PASS. Also run `go vet ./...` to confirm no leftover `var _` reminder lines cause issues.

- [ ] **Step 8: Commit**

```bash
git add internal/adapter/cli/
git commit -m "feat(cli): ls, done, undone, rm"
```

---

### Task 11: CLI `edit` and `tag`

**Files:**
- Create: `internal/adapter/cli/edit.go`
- Create: `internal/adapter/cli/tag.go`
- Modify: `internal/adapter/cli/cli.go` (register)
- Test: `internal/adapter/cli/cli_edit_test.go`

**Interfaces:**
- Produces: `newEditCmd`, `newTagCmd` (`func(*todo.Service) *cobra.Command`).

- [ ] **Step 1: Write the failing tests**

`internal/adapter/cli/cli_edit_test.go`:
```go
package cli

import "testing"

func TestEditUpdatesTitleAndNote(t *testing.T) {
	svc, run := newTestCmd(t)
	_, _ = run("add", "old title", "-l", "work")
	if _, err := run("edit", "1", "--title", "new title", "-n", "a note", "-l", "work"); err != nil {
		t.Fatal(err)
	}
	l, _ := svc.GetList("work")
	if l.Tasks[0].Title != "new title" || l.Tasks[0].Notes != "a note" {
		t.Errorf("edit failed: %+v", l.Tasks[0])
	}
}

func TestTagAddAndRemove(t *testing.T) {
	svc, run := newTestCmd(t)
	_, _ = run("add", "task", "-l", "work")
	if _, err := run("tag", "1", "--add", "urgent", "--add", "home", "-l", "work"); err != nil {
		t.Fatal(err)
	}
	l, _ := svc.GetList("work")
	if len(l.Tasks[0].Tags) != 2 {
		t.Fatalf("tags = %v", l.Tasks[0].Tags)
	}
	if _, err := run("tag", "1", "--rm", "home", "-l", "work"); err != nil {
		t.Fatal(err)
	}
	l, _ = svc.GetList("work")
	if len(l.Tasks[0].Tags) != 1 || l.Tasks[0].Tags[0] != "urgent" {
		t.Errorf("tags after rm = %v", l.Tasks[0].Tags)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/adapter/cli/`
Expected: FAIL — `edit`/`tag` unknown.

- [ ] **Step 3: Implement `edit`**

`internal/adapter/cli/edit.go`:
```go
package cli

import (
	"fmt"

	"github.com/kendallowen/todo/internal/todo"
	"github.com/spf13/cobra"
)

func newEditCmd(svc *todo.Service) *cobra.Command {
	var list, title, note string
	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit a task's title and/or note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(args)
			if err != nil {
				return err
			}
			var tp, np *string
			if cmd.Flags().Changed("title") {
				tp = &title
			}
			if cmd.Flags().Changed("note") {
				np = &note
			}
			if err := svc.EditTask(resolveList(list), ids[0], tp, np); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "edited #%d\n", ids[0])
			return nil
		},
	}
	cmd.Flags().StringVarP(&list, "list", "l", "", "list name (default: inbox or $TODO_LIST)")
	cmd.Flags().StringVar(&title, "title", "", "new title")
	cmd.Flags().StringVarP(&note, "note", "n", "", "new note")
	return cmd
}
```

- [ ] **Step 4: Implement `tag`**

`internal/adapter/cli/tag.go`:
```go
package cli

import (
	"fmt"

	"github.com/kendallowen/todo/internal/todo"
	"github.com/spf13/cobra"
)

func newTagCmd(svc *todo.Service) *cobra.Command {
	var list string
	var add, rm []string
	cmd := &cobra.Command{
		Use:   "tag <id>",
		Short: "Add or remove tags on a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(args)
			if err != nil {
				return err
			}
			name := resolveList(list)
			for _, tag := range add {
				if err := svc.AddTaskTag(name, ids[0], tag); err != nil {
					return err
				}
			}
			for _, tag := range rm {
				if err := svc.RemoveTaskTag(name, ids[0], tag); err != nil {
					return err
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "updated tags on #%d\n", ids[0])
			return nil
		},
	}
	cmd.Flags().StringVarP(&list, "list", "l", "", "list name (default: inbox or $TODO_LIST)")
	cmd.Flags().StringSliceVar(&add, "add", nil, "tag to add (repeatable)")
	cmd.Flags().StringSliceVar(&rm, "rm", nil, "tag to remove (repeatable)")
	return cmd
}
```

- [ ] **Step 5: Register the commands**

In `internal/adapter/cli/cli.go`, after the Task 10 registrations add:
```go
	root.AddCommand(newEditCmd(svc))
	root.AddCommand(newTagCmd(svc))
```

- [ ] **Step 6: Run the tests to verify they pass**

Run: `go test ./internal/adapter/cli/`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/adapter/cli/
git commit -m "feat(cli): edit and tag commands"
```

---

### Task 12: CLI `lists` command group

**Files:**
- Create: `internal/adapter/cli/lists.go`
- Modify: `internal/adapter/cli/cli.go` (register)
- Test: `internal/adapter/cli/cli_lists_test.go`

**Interfaces:**
- Produces: `newListsCmd(svc *todo.Service) *cobra.Command` with subcommands `new`, `rm`, `rename`.

- [ ] **Step 1: Write the failing tests**

`internal/adapter/cli/cli_lists_test.go`:
```go
package cli

import (
	"strings"
	"testing"
)

func TestListsShowsCounts(t *testing.T) {
	_, run := newTestCmd(t)
	_, _ = run("add", "a", "-l", "work")
	_, _ = run("done", "1", "-l", "work")
	_, _ = run("add", "b", "-l", "work")
	out, err := run("lists")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "work") {
		t.Errorf("lists missing 'work': %q", out)
	}
}

func TestListsNewRenameRm(t *testing.T) {
	svc, run := newTestCmd(t)
	if _, err := run("lists", "new", "ideas"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GetList("ideas"); err != nil {
		t.Fatalf("list not created: %v", err)
	}
	if _, err := run("lists", "rename", "ideas", "later"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GetList("later"); err != nil {
		t.Errorf("rename failed: %v", err)
	}
	if _, err := run("lists", "rm", "later", "--force"); err != nil {
		t.Fatal(err)
	}
	names, _ := svc.ListNames()
	if len(names) != 0 {
		t.Errorf("list not removed: %v", names)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/adapter/cli/`
Expected: FAIL — `lists` unknown.

- [ ] **Step 3: Implement the `lists` group**

`internal/adapter/cli/lists.go`:
```go
package cli

import (
	"fmt"

	"github.com/kendallowen/todo/internal/todo"
	"github.com/spf13/cobra"
)

func newListsCmd(svc *todo.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lists",
		Short: "Manage lists",
		RunE: func(cmd *cobra.Command, args []string) error {
			names, err := svc.ListNames()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			for _, name := range names {
				l, err := svc.GetList(name)
				if err != nil {
					return err
				}
				open, done := 0, 0
				for _, task := range l.Tasks {
					if task.Done {
						done++
					} else {
						open++
					}
				}
				fmt.Fprintf(out, "%s (%d open, %d done)\n", name, open, done)
			}
			return nil
		},
	}

	newCmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Create a list",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := svc.CreateList(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created list %q\n", args[0])
			return nil
		},
	}

	var force bool
	rmCmd := &cobra.Command{
		Use:   "rm <name>",
		Short: "Delete a list",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				return fmt.Errorf("refusing to delete %q without --force", args[0])
			}
			if err := svc.DeleteList(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "deleted list %q\n", args[0])
			return nil
		},
	}
	rmCmd.Flags().BoolVar(&force, "force", false, "confirm deletion")

	renameCmd := &cobra.Command{
		Use:   "rename <old> <new>",
		Short: "Rename a list",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := svc.RenameList(args[0], args[1]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "renamed %q to %q\n", args[0], args[1])
			return nil
		},
	}

	cmd.AddCommand(newCmd, rmCmd, renameCmd)
	return cmd
}
```

- [ ] **Step 4: Register the command**

In `internal/adapter/cli/cli.go`, after the Task 11 registrations add:
```go
	root.AddCommand(newListsCmd(svc))
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./internal/adapter/cli/`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/adapter/cli/
git commit -m "feat(cli): lists command group (new/rm/rename + counts)"
```

---

### Task 13: Composition root (runnable CLI)

**Files:**
- Create: `cmd/todo/main.go`

**Interfaces:**
- Consumes: `jsonstore.New`, `todo.NewService`, `cli.NewRootCmd`.
- Produces: a runnable `todo` binary. TUI launch is a temporary placeholder, replaced in Task 16.

- [ ] **Step 1: Write `main.go`**

`cmd/todo/main.go`:
```go
package main

import (
	"fmt"
	"os"

	"github.com/kendallowen/todo/internal/adapter/cli"
	"github.com/kendallowen/todo/internal/adapter/jsonstore"
	"github.com/kendallowen/todo/internal/todo"
)

func main() {
	dir := os.Getenv("TODO_DIR")
	store, err := jsonstore.New(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	svc := todo.NewService(store)

	// Placeholder until the TUI adapter lands in Task 16.
	launchTUI := func() error {
		fmt.Fprintln(os.Stderr, "TUI not yet implemented — use subcommands (todo --help)")
		return nil
	}

	root := cli.NewRootCmd(svc, launchTUI)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Build and smoke-test the CLI end-to-end**

Run:
```bash
go build ./...
TODO_DIR="$(mktemp -d)" go run ./cmd/todo add "first task" -l demo
TODO_DIR=/tmp/does-not-persist go run ./cmd/todo --help
```
Expected: build succeeds; the `add` prints `Added [demo] #1: first task`; `--help` lists `add`, `ls`, `done`, `undone`, `rm`, `edit`, `tag`, `lists`, `tui`.

- [ ] **Step 3: Run the full test suite + vet**

Run: `go test ./... && go vet ./...`
Expected: PASS, no vet complaints.

- [ ] **Step 4: Commit**

```bash
git add cmd/todo/main.go
git commit -m "feat(cmd): composition root wiring CLI (TUI placeholder)"
```

---

### Task 14: TUI three-pane skeleton

**Files:**
- Create: `internal/adapter/tui/app.go`
- Test: `internal/adapter/tui/app_test.go`

**Interfaces:**
- Consumes: `*todo.Service`.
- Produces: `func New(svc *todo.Service) *App`; `func (a *App) Run() error`; internal `func (a *App) buildUI()`, `func (a *App) refreshLists()`, `func (a *App) refreshTasks()`, `func (a *App) refreshDetail()`; field `a.app *tview.Application`. The skeleton populates Lists/Tasks/Detail panes from the service.

- [ ] **Step 1: Add the tview/tcell dependencies**

Run:
```bash
go get github.com/rivo/tview@latest
go get github.com/gdamore/tcell/v2@latest
```
Expected: updates `go.mod`/`go.sum`.

- [ ] **Step 2: Write the failing smoke test (SimulationScreen)**

`internal/adapter/tui/app_test.go`:
```go
package tui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/kendallowen/todo/internal/adapter/jsonstore"
	"github.com/kendallowen/todo/internal/todo"
)

func newTestApp(t *testing.T) (*App, *todo.Service) {
	t.Helper()
	store, err := jsonstore.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	svc := todo.NewService(store)
	return New(svc), svc
}

func TestSkeletonRendersListsAndTasks(t *testing.T) {
	app, svc := newTestApp(t)
	if _, err := svc.AddTask("work", "ship it", nil, ""); err != nil {
		t.Fatal(err)
	}

	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	screen.SetSize(100, 24)

	app.buildUI()
	app.refreshLists()
	app.refreshTasks()
	app.refreshDetail()

	app.app.SetScreen(screen)
	go func() { _ = app.app.Run() }()
	defer app.app.Stop()
	app.app.Draw()

	if !screenContains(screen, "work") {
		t.Error("Lists pane should show 'work'")
	}
	if !screenContains(screen, "ship it") {
		t.Error("Tasks pane should show 'ship it'")
	}
}

// screenContains scans the simulation screen's cells for a substring on any row.
func screenContains(s tcell.SimulationScreen, want string) bool {
	cells, w, h := s.GetContents()
	for y := 0; y < h; y++ {
		row := make([]rune, 0, w)
		for x := 0; x < w; x++ {
			c := cells[y*w+x]
			if len(c.Runes) > 0 {
				row = append(row, c.Runes[0])
			} else {
				row = append(row, ' ')
			}
		}
		if containsRunes(string(row), want) {
			return true
		}
	}
	return false
}

func containsRunes(haystack, needle string) bool {
	return len(needle) == 0 || (len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0)
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `go test ./internal/adapter/tui/`
Expected: FAIL — `New`/`App`/`buildUI` undefined.

- [ ] **Step 4: Implement the skeleton**

`internal/adapter/tui/app.go`:
```go
package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kendallowen/todo/internal/todo"
	"github.com/rivo/tview"
)

// App is the tview front-end over the shared Service.
type App struct {
	svc *todo.Service
	app *tview.Application

	lists  *tview.List
	tasks  *tview.List
	detail *tview.TextView

	listNames []string
	current   *todo.List // currently displayed list
}

// New creates a TUI App.
func New(svc *todo.Service) *App {
	return &App{svc: svc, app: tview.NewApplication()}
}

// buildUI constructs the three-pane layout.
func (a *App) buildUI() {
	a.lists = tview.NewList().ShowSecondaryText(false)
	a.lists.SetBorder(true).SetTitle(" Lists ")

	a.tasks = tview.NewList().ShowSecondaryText(false)
	a.tasks.SetBorder(true).SetTitle(" Tasks ")

	a.detail = tview.NewTextView().SetDynamicColors(true)
	a.detail.SetBorder(true).SetTitle(" Detail ")

	a.lists.SetChangedFunc(func(int, string, string, rune) {
		a.refreshTasks()
		a.refreshDetail()
	})
	a.tasks.SetChangedFunc(func(int, string, string, rune) {
		a.refreshDetail()
	})

	flex := tview.NewFlex().
		AddItem(a.lists, 24, 0, true).
		AddItem(a.tasks, 0, 2, false).
		AddItem(a.detail, 0, 2, false)

	a.app.SetRoot(flex, true)
}

// refreshLists reloads the Lists pane from the service.
func (a *App) refreshLists() {
	names, err := a.svc.ListNames()
	if err != nil {
		return
	}
	a.listNames = names
	a.lists.Clear()
	for _, name := range names {
		a.lists.AddItem(name, "", 0, nil)
	}
}

// selectedListName returns the highlighted list name (or "").
func (a *App) selectedListName() string {
	i := a.lists.GetCurrentItem()
	if i < 0 || i >= len(a.listNames) {
		return ""
	}
	return a.listNames[i]
}

// refreshTasks reloads the Tasks pane for the selected list.
func (a *App) refreshTasks() {
	a.tasks.Clear()
	name := a.selectedListName()
	if name == "" {
		a.current = nil
		return
	}
	l, err := a.svc.GetList(name)
	if err != nil {
		a.current = nil
		return
	}
	a.current = l
	for _, task := range l.Tasks {
		box := "[ ]"
		if task.Done {
			box = "[x]"
		}
		label := fmt.Sprintf("%s #%d %s", box, task.ID, task.Title)
		if len(task.Tags) > 0 {
			label += "  #" + strings.Join(task.Tags, " #")
		}
		a.tasks.AddItem(label, "", 0, nil)
	}
}

// selectedTask returns the highlighted task pointer (or nil).
func (a *App) selectedTask() *todo.Task {
	if a.current == nil {
		return nil
	}
	i := a.tasks.GetCurrentItem()
	if i < 0 || i >= len(a.current.Tasks) {
		return nil
	}
	return &a.current.Tasks[i]
}

// refreshDetail updates the Detail pane for the selected task.
func (a *App) refreshDetail() {
	t := a.selectedTask()
	if t == nil {
		a.detail.SetText("")
		return
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", t.Title)
	if len(t.Tags) > 0 {
		fmt.Fprintf(&b, "#%s\n\n", strings.Join(t.Tags, " #"))
	}
	fmt.Fprintf(&b, "Notes:\n%s\n", t.Notes)
	a.detail.SetText(b.String())
}

// Run builds the UI and starts the event loop.
func (a *App) Run() error {
	a.buildUI()
	a.refreshLists()
	a.refreshTasks()
	a.refreshDetail()
	a.bindKeys()
	return a.app.Run()
}

// bindKeys is implemented in Task 15; defined here so Run compiles.
func (a *App) bindKeys() {}

var _ = tcell.KeyTab
```

(Remove the trailing `var _ = tcell.KeyTab` line once Task 15 uses `tcell`.)

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./internal/adapter/tui/`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum internal/adapter/tui/
git commit -m "feat(tui): three-pane skeleton populated from Service"
```

---

### Task 15: TUI interactions (keys + modals)

**Files:**
- Modify: `internal/adapter/tui/app.go`
- Create: `internal/adapter/tui/forms.go`
- Test: `internal/adapter/tui/app_interact_test.go`

**Interfaces:**
- Produces: real `func (a *App) bindKeys()`; pane focus cycling; key handlers calling the service; `addTaskForm`, `editTaskForm`, `confirm` modal helpers. After mutations, panes refresh from the service.

- [ ] **Step 1: Write the failing interaction test**

`internal/adapter/tui/app_interact_test.go`:
```go
package tui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/kendallowen/todo/internal/todo"
)

func TestToggleKeyMarksTaskDone(t *testing.T) {
	app, svc := newTestApp(t)
	if _, err := svc.AddTask("work", "ship it", nil, ""); err != nil {
		t.Fatal(err)
	}

	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	screen.SetSize(100, 24)

	app.buildUI()
	app.refreshLists()
	app.refreshTasks()
	app.bindKeys()
	app.app.SetScreen(screen)
	go func() { _ = app.app.Run() }()
	defer app.app.Stop()

	// Focus the tasks pane and toggle the first task.
	app.app.QueueUpdateDraw(func() {
		app.focus(1) // tasks pane
		app.toggleSelected()
	})
	app.app.QueueUpdate(func() {}) // flush

	l, _ := svc.GetList("work")
	if !l.Tasks[0].Done {
		t.Errorf("expected task to be done after toggle; got %+v", l.Tasks[0])
	}
}

func TestFocusCyclesWithinRange(t *testing.T) {
	app, _ := newTestApp(t)
	app.buildUI()
	app.focus(0)
	app.cycleFocus(1)
	if app.focusIdx != 1 {
		t.Errorf("focusIdx = %d, want 1", app.focusIdx)
	}
	app.cycleFocus(1)
	app.cycleFocus(1) // wraps 2 -> 0
	if app.focusIdx != 0 {
		t.Errorf("focusIdx after wrap = %d, want 0", app.focusIdx)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/adapter/tui/`
Expected: FAIL — `focus`/`toggleSelected`/`cycleFocus`/`focusIdx` undefined.

- [ ] **Step 3: Implement focus + key handling**

Replace the placeholder `bindKeys` and the `var _ = tcell.KeyTab` line in `internal/adapter/tui/app.go`, and add a `focusIdx int` field + `panes []tview.Primitive` to the `App` struct. Add:
```go
// focus sets the focused pane by index (0=lists,1=tasks,2=detail).
func (a *App) focus(i int) {
	if a.panes == nil {
		a.panes = []tview.Primitive{a.lists, a.tasks, a.detail}
	}
	if i < 0 || i >= len(a.panes) {
		return
	}
	a.focusIdx = i
	a.app.SetFocus(a.panes[i])
}

// cycleFocus moves focus by delta with wraparound.
func (a *App) cycleFocus(delta int) {
	n := 3
	a.focus(((a.focusIdx+delta)%n + n) % n)
}

// toggleSelected flips the done state of the selected task and refreshes.
func (a *App) toggleSelected() {
	t := a.selectedTask()
	if t == nil || a.current == nil {
		return
	}
	if err := a.svc.ToggleTask(a.current.Name, t.ID); err != nil {
		return
	}
	idx := a.tasks.GetCurrentItem()
	a.refreshTasks()
	if idx < a.tasks.GetItemCount() {
		a.tasks.SetCurrentItem(idx)
	}
	a.refreshDetail()
}

// deleteSelected removes the selected task after confirmation.
func (a *App) deleteSelected() {
	t := a.selectedTask()
	if t == nil || a.current == nil {
		return
	}
	name, id := a.current.Name, t.ID
	a.confirm(fmt.Sprintf("Delete #%d %q?", id, t.Title), func() {
		_ = a.svc.RemoveTask(name, id)
		a.refreshTasks()
		a.refreshDetail()
	})
}

// bindKeys installs global and pane key handlers.
func (a *App) bindKeys() {
	a.panes = []tview.Primitive{a.lists, a.tasks, a.detail}
	a.app.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyTab:
			a.cycleFocus(1)
			return nil
		case tcell.KeyBacktab:
			a.cycleFocus(-1)
			return nil
		case tcell.KeyCtrlC:
			a.app.Stop()
			return nil
		}
		switch ev.Rune() {
		case 'q':
			a.app.Stop()
			return nil
		case 'a':
			if a.focusIdx == 0 {
				a.newListForm()
			} else {
				a.addTaskForm()
			}
			return nil
		case 'd':
			if a.focusIdx == 1 {
				a.toggleSelected()
			}
			return nil
		case 'x':
			if a.focusIdx == 0 {
				a.deleteSelectedList()
			} else if a.focusIdx == 1 {
				a.deleteSelected()
			}
			return nil
		case 'r':
			if a.focusIdx == 0 {
				a.renameListForm()
			}
			return nil
		case 'e':
			if a.focusIdx == 1 {
				a.editTaskForm()
			}
			return nil
		case 'n':
			if a.focusIdx == 1 {
				a.editTaskForm()
			}
			return nil
		}
		return ev
	})
	a.focus(0)
}
```

- [ ] **Step 4: Implement the forms/modals**

`internal/adapter/tui/forms.go`:
```go
package tui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

// confirm shows a yes/no modal, running onYes if confirmed.
func (a *App) confirm(prompt string, onYes func()) {
	prev := a.app.GetFocus()
	modal := tview.NewModal().
		SetText(prompt).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(_ int, label string) {
			a.rebuildRoot()
			if label == "Yes" {
				onYes()
			}
			a.app.SetFocus(prev)
		})
	a.app.SetRoot(modal, true)
}

// rebuildRoot restores the main three-pane flex as the root.
func (a *App) rebuildRoot() {
	flex := tview.NewFlex().
		AddItem(a.lists, 24, 0, true).
		AddItem(a.tasks, 0, 2, false).
		AddItem(a.detail, 0, 2, false)
	a.app.SetRoot(flex, true)
	a.focus(a.focusIdx)
}

// addTaskForm collects a new task and adds it to the current list.
func (a *App) addTaskForm() {
	list := "inbox"
	if a.current != nil {
		list = a.current.Name
	} else if n := a.selectedListName(); n != "" {
		list = n
	}
	var title, tags, notes string
	form := tview.NewForm().
		AddInputField("Title", "", 40, nil, func(s string) { title = s }).
		AddInputField("Tags (space-sep)", "", 40, nil, func(s string) { tags = s }).
		AddInputField("Notes", "", 40, nil, func(s string) { notes = s })
	form.AddButton("Add", func() {
		if strings.TrimSpace(title) != "" {
			_, _ = a.svc.AddTask(list, title, strings.Fields(tags), notes)
		}
		a.rebuildRoot()
		a.refreshLists()
		a.refreshTasks()
		a.refreshDetail()
	})
	form.AddButton("Cancel", func() { a.rebuildRoot() })
	form.SetBorder(true).SetTitle(" Add task ")
	a.app.SetRoot(form, true)
}

// editTaskForm edits the selected task's title and notes.
func (a *App) editTaskForm() {
	t := a.selectedTask()
	if t == nil || a.current == nil {
		return
	}
	name, id := a.current.Name, t.ID
	title, notes := t.Title, t.Notes
	form := tview.NewForm().
		AddInputField("Title", title, 40, nil, func(s string) { title = s }).
		AddInputField("Notes", notes, 40, nil, func(s string) { notes = s })
	form.AddButton("Save", func() {
		tp, np := title, notes
		_ = a.svc.EditTask(name, id, &tp, &np)
		a.rebuildRoot()
		a.refreshTasks()
		a.refreshDetail()
	})
	form.AddButton("Cancel", func() { a.rebuildRoot() })
	form.SetBorder(true).SetTitle(" Edit task ")
	a.app.SetRoot(form, true)
}

// newListForm creates a new list.
func (a *App) newListForm() {
	var name string
	form := tview.NewForm().
		AddInputField("List name", "", 30, nil, func(s string) { name = s })
	form.AddButton("Create", func() {
		if strings.TrimSpace(name) != "" {
			_ = a.svc.CreateList(name)
		}
		a.rebuildRoot()
		a.refreshLists()
	})
	form.AddButton("Cancel", func() { a.rebuildRoot() })
	form.SetBorder(true).SetTitle(" New list ")
	a.app.SetRoot(form, true)
}

// deleteSelectedList removes the highlighted list after confirmation.
func (a *App) deleteSelectedList() {
	name := a.selectedListName()
	if name == "" {
		return
	}
	a.confirm(fmt.Sprintf("Delete list %q and all its tasks?", name), func() {
		_ = a.svc.DeleteList(name)
		a.refreshLists()
		a.refreshTasks()
		a.refreshDetail()
	})
}

// renameListForm renames the highlighted list.
func (a *App) renameListForm() {
	old := a.selectedListName()
	if old == "" {
		return
	}
	name := old
	form := tview.NewForm().
		AddInputField("New name", old, 30, nil, func(s string) { name = s })
	form.AddButton("Rename", func() {
		if strings.TrimSpace(name) != "" && name != old {
			_ = a.svc.RenameList(old, name)
		}
		a.rebuildRoot()
		a.refreshLists()
		a.refreshTasks()
		a.refreshDetail()
	})
	form.AddButton("Cancel", func() { a.rebuildRoot() })
	form.SetBorder(true).SetTitle(" Rename list ")
	a.app.SetRoot(form, true)
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./internal/adapter/tui/`
Expected: PASS.

- [ ] **Step 6: Run vet + full suite**

Run: `go vet ./... && go test ./...`
Expected: PASS, no vet complaints (confirm no leftover `var _` / unused helpers).

- [ ] **Step 7: Commit**

```bash
git add internal/adapter/tui/
git commit -m "feat(tui): key bindings, focus cycling, add/edit/delete modals"
```

---

### Task 16: Wire TUI into main + README + verification

**Files:**
- Modify: `cmd/todo/main.go` (replace the TUI placeholder)
- Create: `README.md`

**Interfaces:**
- Consumes: `tui.New(svc).Run`.

- [ ] **Step 1: Replace the TUI placeholder in `main.go`**

In `cmd/todo/main.go`, add the import `"github.com/kendallowen/todo/internal/adapter/tui"` and replace the `launchTUI` placeholder with:
```go
	launchTUI := func() error {
		return tui.New(svc).Run()
	}
```

- [ ] **Step 2: Build and verify the binary**

Run:
```bash
go build ./...
go vet ./...
go test ./...
```
Expected: all PASS, no vet output.

- [ ] **Step 3: Write the README**

`README.md`:
```markdown
# todo

A task tracker with a full CLI and an interactive three-pane TUI, built in Go
using a hexagonal (ports & adapters) architecture.

## Install

    brew install go            # if not already installed
    go install ./cmd/todo      # puts `todo` on ~/go/bin (ensure it's on PATH)

## Storage

One JSON file per list under (in order): `$TODO_DIR`, `$XDG_DATA_HOME/todo`,
or `~/.local/share/todo`. Point `TODO_DIR` at a git repo to version your todos.

## CLI

    todo                       # launch the TUI
    todo tui                   # launch the TUI explicitly
    todo add "buy milk" -l groceries -t store -n "2%"
    todo ls [-l list | -a] [-t tag] [--done|--open]
    todo done 3 [-l list]
    todo undone 3 [-l list]
    todo rm 3 [-l list]
    todo edit 3 --title "new" -n "note" [-l list]
    todo tag 3 --add urgent --rm home [-l list]
    todo lists
    todo lists new ideas
    todo lists rename ideas later
    todo lists rm later --force

Default list is `inbox`; override with `-l` or `$TODO_LIST`.

## TUI keys

Panes: Lists | Tasks | Detail. `Tab`/`Shift-Tab` switch panes.
Tasks: `a` add, `d` toggle done, `e`/`n` edit, `x` delete.
Lists: `a` new, `r` rename, `x` delete. `q`/`Ctrl-C` quit.

## Development

    go test ./...
    go vet ./...
```

- [ ] **Step 4: Manual end-to-end check (real terminal)**

Run:
```bash
go install ./cmd/todo
TODO_DIR="$(mktemp -d)" todo add "try the TUI" -l demo
TODO_DIR="$(mktemp -d)" todo            # launches TUI; press a, d, q to verify
```
Expected: `add` persists and prints; bare `todo` opens the three-pane TUI; adding/toggling/quitting works.

- [ ] **Step 5: Commit**

```bash
git add cmd/todo/main.go README.md
git commit -m "feat(cmd): wire TUI into root command; add README"
```

---

## Notes for the implementer

- Task 14 deliberately keeps `var _ = tcell.KeyTab` so the `tcell` import stays
  satisfied before Task 15's key handlers use it; Task 15 removes that line. It is
  valid Go and compiles — leave it until Task 15. Run `go vet ./...` at the end of
  every task to catch unused imports or dead code.
- Keep all logic in `internal/todo`; if you find yourself writing task/list logic
  inside a `cli` or `tui` file, move it into the core and call it from the adapter.
- The TUI smoke tests run the tview event loop on a `SimulationScreen`; they assert
  on rendered cells / service state, not on exact pixel layout, so they stay robust.
