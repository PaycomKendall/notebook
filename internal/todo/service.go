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
	list, err := NormalizeListName(list)
	if err != nil {
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

// GetList returns a list by canonical name (ErrListNotFound if absent).
func (s *Service) GetList(name string) (*List, error) {
	name, err := NormalizeListName(name)
	if err != nil {
		return nil, err
	}
	return s.repo.Load(name)
}

// ListNames returns the names of all stored lists.
func (s *Service) ListNames() ([]string, error) { return s.repo.Names() }

// mutate loads an existing list (by canonical name), applies fn, and saves it.
func (s *Service) mutate(list string, fn func(*List) error) error {
	name, err := NormalizeListName(list)
	if err != nil {
		return err
	}
	l, err := s.repo.Load(name)
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

func (s *Service) CreateList(name string) error {
	name, err := NormalizeListName(name)
	if err != nil {
		return err
	}
	_, err = s.repo.Create(name)
	return err
}

func (s *Service) DeleteList(name string) error {
	name, err := NormalizeListName(name)
	if err != nil {
		return err
	}
	return s.repo.Delete(name)
}

func (s *Service) RenameList(old, newName string) error {
	oldName, err := NormalizeListName(old)
	if err != nil {
		return err
	}
	newName, err = NormalizeListName(newName)
	if err != nil {
		return err
	}
	return s.repo.Rename(oldName, newName)
}
