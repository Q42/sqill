package registry

import (
	"fmt"
	"sort"
)

type SkillEntry struct {
	Name        string
	Source      string
	Description string
}

type Provider interface {
	Resolve(name string) (SkillEntry, error)
	All() []SkillEntry
}

func sortEntries(entries []SkillEntry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
}

func ErrNotFound(name string) error {
	return fmt.Errorf("skill %q not found in registry", name)
}
