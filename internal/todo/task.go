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
