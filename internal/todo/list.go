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

// Append adds an existing task, assigning it a fresh ID from this list's NextID
// and bumping Updated, while preserving its other fields. Tags are cloned so the
// source task is not aliased.
func (l *List) Append(t Task) *Task {
	if l.NextID == 0 {
		l.NextID = 1
	}
	t.ID = l.NextID
	l.NextID++
	t.Updated = time.Now()
	if t.Tags != nil {
		t.Tags = append([]string(nil), t.Tags...)
	}
	l.Tasks = append(l.Tasks, t)
	return &l.Tasks[len(l.Tasks)-1]
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
