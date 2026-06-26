# Architecture

## Project structure

```text
src/
  cmd/              — cobra root command + wiring
    init/           — `init` command (state file + tool symlinks)
    install/         — `install` command
    remove/         — `remove` command
    update/         — `update` command
    list/           — `list` command
    search/         — `search` command
    info/           — `info` command
  lib/
    utils/          — shared helpers (path display, validation, dedup)
    registry/       — RegistryProvider interface + hardcoded impl
    source/         — SourceProvider interface (git, file, archive)
    installer/      — install/remove/update orchestration
    metadata/       — sqill.json read/write store
    runtime/        — shared Runtime struct passed to all non-init subcommands
```

## Data model

### Per-skill manifest

`.agents/skills/<name>/sqill.json`:

```json
{ "name": "...", "version": "x.y.z", "description": "..." }
```

### Unified state

`.agents/skills/sqill.json`:

```json
{
  "installed": {
    "<name>": { "version": "...", "source": "...", "installed_at": "..." }
  },
  "registries": []
}
```

### Registry (in binary)

```go
var defaultRegistry = map[string]SkillEntry{
    "github-search": {
        Name:        "github-search",
        Source:      "https://github.com/org/github-search-skill.git",
        Description: "Search GitHub repositories.",
    },
    "jira": {
        Name:        "jira",
        Source:      "git@github.com:org/jira-skill.git",
        Description: "Manage Jira issues.",
    },
    "postgres": {
        Name:        "postgres",
        Source:      "file:///opt/skills/postgres",
        Description: "Manage Postgres databases.",
    },
}
```

## Source types

| Prefix                   | Handler                                      |
| ------------------------ | -------------------------------------------- |
| `git@`, `https://...git` | `GitSourceProvider` (go-git clone)           |
| `file://`                | `LocalSourceProvider` (recursive copy)       |
| `https://...tar.gz`      | `ArchiveSourceProvider` (download + extract) |

## Safety invariants

1. Skill name must not contain `..`, `/`, `\`, or start with `.`
2. Manifest (`sqill.json`) must exist and `name` field must match
3. Metadata writes are atomic (write to temp, rename)

## We don't

- Store anything outside `.agents/skills/`
- Require network for `list`, `info`, `remove`
- Cache downloaded sources (fetched fresh each time)
- Support version pinning yet (always latest from source)

## Build + test

```bash
go build -o sqill .              # compile binary
go test ./...                     # run all tests
go vet ./...                      # lint
```