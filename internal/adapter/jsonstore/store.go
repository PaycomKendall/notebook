package jsonstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kendallowen/notebook/internal/todo"
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

// DefaultDir resolves the data directory: NB_DIR, then XDG, then ~/.local/share.
func DefaultDir() (string, error) {
	if d := os.Getenv("NB_DIR"); d != "" {
		return d, nil
	}
	if d := os.Getenv("XDG_DATA_HOME"); d != "" {
		return filepath.Join(d, "notebook"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "notebook"), nil
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
