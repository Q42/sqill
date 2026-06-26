package registry

import (
	"fmt"
	"sort"
	"strings"
)

type SkillEntry struct {
	Name        string
	Source      string
	Description string
}

type Provider interface {
	Search(query string) []SkillEntry
	Resolve(name string) (SkillEntry, error)
	All() []SkillEntry
}

func match(entry SkillEntry, query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return true
	}
	if strings.Contains(strings.ToLower(entry.Name), q) {
		return true
	}
	if entry.Description != "" && strings.Contains(strings.ToLower(entry.Description), q) {
		return true
	}
	return false
}

func sortEntries(entries []SkillEntry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
}

func ErrNotFound(name string) error {
	return fmt.Errorf("skill %q not found in registry", name)
}
