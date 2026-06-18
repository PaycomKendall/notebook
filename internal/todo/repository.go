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
