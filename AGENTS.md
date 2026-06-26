# AGENTS.md — Agent Instructions for Sqill

- `README.md` is for **end users only**: install, usage, build, source layout. Never add contributor workflow, release process, agent internals, or internal conventions there.
- `AGENTS.md` (this file) is the source of truth for development workflow, release process, and agent behavior. Update it whenever a new convention is introduced.
- For any **user-facing behavior** change, also update `README.md`.

## Project overview

Sqill is a Go CLI tool: single static binary, no database, no daemons. It installs skills into `.agents/skills/`, tracks metadata in `.agents/skills/sqill.json`, and resolves skill sources from a hardcoded registry.

## Setup workflow

All commands except `init` require `.agents/skills/sqill.json` to exist. Running `sqill init` will:

1. Create `.agents/skills/` and `.agents/skills/sqill.json` if missing (idempotent). If the state file already exists, `init` prints "already initialized" and skips creation.
2. Optionally symlink `.claude/skills`, `.cursor/skills`, and `.kilo/skills` (siblings of `.agents/`) into `.agents/skills`. The user is prompted per target unless `--yes` is passed or flags pre-select (`--link-claude`, `--link-cursor`, `--link-kilo`); when already initialized, prompting is skipped.
3. If a target dir already exists, its contents are moved into `.agents/skills/` and the dir is replaced with a symlink. If both directories contain a skill with the same name, `init` fails and asks the user to de-duplicate.

## Build + test

```bash
go build -o sqill .              # compile binary
go test ./...                     # run all tests
go vet ./...                      # lint
```

## Coding conventions

- **No comments** unless strictly necessary.
- Standard Go project layout: `src/cmd/` for CLI entry points, `src/lib/` for non-exported packages.
- Interfaces: `RegistryProvider`, `SourceProvider`, `MetadataStore`.
- Errors are returned, never panicked, except in `main.go` for fatal startup failures.
- Zero external configuration needed to run — registry is hardcoded.

## Key files

| File                                          | Purpose                                                               |
| --------------------------------------------- | --------------------------------------------------------------------- |
| `main.go`                                     | Entry point, calls `cmd.Execute()`                                    |
| `src/cmd/cmd.go`                              | Root cobra command + subcommand wiring + init guard                   |
| `src/cmd/init/init.go`                        | `init` command: create state file + tool symlinks                     |
| `src/cmd/install/install.go`                  | `install <name>` command                                              |
| `src/cmd/remove/remove.go`                    | `remove <name>` command                                               |
| `src/cmd/update/update.go`                    | `update <name>` command                                               |
| `src/cmd/list/list.go`                         | `list` command                                                        |
| `src/cmd/search/search.go`                    | `search <query>` command                                              |
| `src/cmd/info/info.go`                        | `info <name>` command                                                 |
| `src/lib/runtime/runtime.go`                   | Shared `Runtime` struct (skillsDir, store, installer, registry)       |
| `src/lib/registry/hardcoded.go`               | Hardcoded `map[string]string` of skill name → source URL              |
| `src/lib/metadata/store.go`                   | Read/write `.agents/skills/sqill.json`                                |
| `src/lib/installer/installer.go`              | Orchestrate resolve → fetch → validate → install → write metadata     |
| `src/lib/utils/utils.go`                       | Shared helpers (path display, validation, safe join, dedup)        |
| `.github/workflows/release.yml`                | Tag-triggered release: build for macOS/Linux × amd64/arm64, publish release with body from `.github/release-notes/v<tag>.md` |
| `.github/release-notes/vX.Y.Z.md`              | Curated release notes for tag `vX.Y.Z`, rendered by the `q-release` skill |
| `.agents/skills/q-release/`                    | Installed (gitignored) `q-release` skill — walks Conventional Commits since the last published tag and produces the next release's notes file |

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

## Release workflow

Tag-triggered CI publishes binaries and a GitHub release. The release body comes from a file in this repo, not from GitHub's auto-generator.

### Flow

1. Use the `q-release` skill to render notes into `.github/release-notes/vX.Y.Z.md` (categorized by Conventional Commits prefix).
2. Commit the notes file on the default branch and push.
3. Create and push the tag: `git tag -a vX.Y.Z -m vX.Y.Z && git push origin vX.Y.Z`.
4. `.github/workflows/release.yml` builds artifacts and creates the release, reading the body from `.github/release-notes/<tag>.md` via `softprops/action-gh-release@v2`'s `body_path`.

### Hard rules

- **Never** set `generate_release_notes: true` on the workflow — it overwrites the curated body.
- **Never** call `gh release create` after pushing the tag — the workflow owns release creation. To re-sync notes after the fact, use `gh release edit <tag> --notes-file .github/release-notes/<tag>.md` (do **not** use `--notes -`, which silently no-ops in `gh release edit`).
- The notes file must be committed on the default branch **before** the tag is created, so the workflow checkout at the tag SHA sees it.
- Do not write release-related scratch files to `/tmp` or `os.TempDir()` — they sit outside the project and trip permission checks. Use files inside the project tree (e.g. `.github/release-notes/`) or pipe via heredoc into `gh`/`git`.

## Safety invariants

1. Skill name must not contain `..`, `/`, `\`, or start with `.`
2. Manifest (`sqill.json`) must exist and `name` field must match
3. Metadata writes are atomic (write to temp, rename)

## We don't

- Store anything outside `.agents/skills/`
- Require network for `list`, `info`, `remove`
- Cache downloaded sources (fetched fresh each time)
- Support version pinning yet (always latest from source)
