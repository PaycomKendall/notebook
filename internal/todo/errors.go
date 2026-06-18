package todo

import "errors"

var (
	ErrListNotFound = errors.New("list not found")
	ErrListExists   = errors.New("list already exists")
	ErrTaskNotFound = errors.New("task not found")
	ErrEmptyTitle   = errors.New("task title must not be empty")
	ErrInvalidName  = errors.New("invalid list name")
)
