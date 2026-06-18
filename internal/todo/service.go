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
