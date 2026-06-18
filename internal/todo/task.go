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

// NormalizeListName trims and lowercases a name, validates it, and returns the
// canonical (storage) form. It is the single source of truth for list-name
// validity; the Service normalizes every list-name argument through it so the
// adapters only ever see canonical names.
func NormalizeListName(name string) (string, error) {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "" || !listNameRe.MatchString(n) {
		return "", ErrInvalidName
	}
	return n, nil
}

// ValidateListName reports whether a name is a valid list name.
func ValidateListName(name string) error {
	_, err := NormalizeListName(name)
	return err
}
