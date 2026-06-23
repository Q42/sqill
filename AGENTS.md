# AGENTS.md — Agent Instructions for Sqill

For any new behaviour requested by the user, update the readme to cover that. Consider it a source of truth.
Also update the AGENTS.md periodically with information.

## Project overview

Sqill is a Go CLI tool: single static binary, no database, no daemons. It installs skills into `.agents/skills/`, tracks metadata in `.agents/skills/sqill.json`, and resolves skill sources from a hardcoded registry.

## Setup workflow

All commands except `setup` require `.agents/skills/sqill.json` to exist. Running `sqill setup` will:

1. Create `.agents/skills/` and `.agents/skills/sqill.json` if missing (idempotent).
2. Optionally symlink `.claude/skills`, `.cursor/skills`, and `.kilo/skills` (siblings of `.agents/`) into `.agents/skills`. The user is prompted per target; `--link-claude`, `--link-cursor`, `--link-kilo` pre-select, `--yes` skips prompts.
3. If a target dir already exists, its contents are moved into `.agents/skills/` and the dir is replaced with a symlink. If both directories contain a skill with the same name, setup fails and asks the user to de-duplicate.

## Build + test

```bash
go build -o sqill .              # compile binary
go test ./...                     # run all tests
go vet ./...                      # lint
```

## Coding conventions

- **No comments** unless strictly necessary.
- Standard Go project layout: `cmd/` for CLI entry points, `internal/` for non-exported packages.
- Interfaces: `RegistryProvider`, `SourceProvider`, `MetadataStore`.
- Errors are returned, never panicked, except in `main.go` for fatal startup failures.
- Zero external configuration needed to run — registry is hardcoded.

## Key files

| File                              | Purpose                                                           |
| --------------------------------- | ----------------------------------------------------------------- |
| `cmd/cmd.go`                      | Root cobra command + subcommand registration + init guard         |
| `cmd/setup.go`                    | `setup` command: init state file + create tool symlinks           |
| `internal/registry/hardcoded.go`  | Hardcoded `map[string]string` of skill name → source URL          |
| `internal/metadata/store.go`      | Read/write `.agents/skills/sqill.json`                            |
| `internal/installer/installer.go` | Orchestrate resolve → fetch → validate → install → write metadata |
| `main.go`                         | Entry point, calls `cmd.Execute()`                                |

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
var defaultRegistry = map[string]string{
    "github-search": "https://github.com/org/github-search-skill.git",
    "jira":         "git@github.com:org/jira-skill.git",
    "postgres":     "file:///opt/skills/postgres",
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
