package registry

import "strings"

var defaultRegistry = map[string]SkillEntry{
	"sRegressor": {
		Name:        "sRegressor",
		Source:      "https://github.com/Q42/sRegressor.git",
		Description: "Run regression tests directly in your project.",
	},
	"q-release": {
		Name:        "q-release",
		Source:      "https://github.com/Q42/q-release.git",
		Description: "Create a GitHub release from the git diff since the last release with standardized notes.",
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
