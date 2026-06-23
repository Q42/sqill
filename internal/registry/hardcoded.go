package registry

import "strings"

var defaultRegistry = map[string]SkillEntry{
	"github-search": {
		Name:        "github-search",
		Source:      "https://github.com/org/github-search-skill.git",
		Description: "Search GitHub repositories",
	},
	"jira": {
		Name:        "jira",
		Source:      "git@github.com:org/jira-skill.git",
		Description: "Interact with Jira issues",
	},
	"postgres": {
		Name:        "postgres",
		Source:      "file:///opt/skills/postgres",
		Description: "Postgres database helpers",
	},
	"example-tgz": {
		Name:        "example-tgz",
		Source:      "https://example.com/skill.tar.gz",
		Description: "Example tarball-distributed skill",
	},
}

type Hardcoded struct {
	entries map[string]SkillEntry
}

func NewHardcoded() *Hardcoded {
	entries := make(map[string]SkillEntry, len(defaultRegistry))
	for k, v := range defaultRegistry {
		entries[k] = v
	}
	return &Hardcoded{entries: entries}
}

func (h *Hardcoded) Search(query string) []SkillEntry {
	var out []SkillEntry
	for _, e := range h.entries {
		if match(e, query) {
			out = append(out, e)
		}
	}
	sortEntries(out)
	return out
}

func (h *Hardcoded) Resolve(name string) (SkillEntry, error) {
	e, ok := h.entries[strings.TrimSpace(name)]
	if !ok {
		return SkillEntry{}, ErrNotFound(name)
	}
	return e, nil
}

func (h *Hardcoded) All() []SkillEntry {
	out := make([]SkillEntry, 0, len(h.entries))
	for _, e := range h.entries {
		out = append(out, e)
	}
	sortEntries(out)
	return out
}

func DefaultRegistry() map[string]SkillEntry {
	out := make(map[string]SkillEntry, len(defaultRegistry))
	for k, v := range defaultRegistry {
		out[k] = v
	}
	return out
}
