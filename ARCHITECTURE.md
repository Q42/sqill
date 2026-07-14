# Architecture

## Project structure

```text
src/
  cmd/              — cobra root command + wiring
    init/           — `init` command (state file + tool symlinks)
    install/        — `install` command
    remove/         — `remove` command
    update/         — `update` command
    list/           — `list` command
    info/           — `info` command
    track/          — `track` command
    untrack/        — `untrack` command
    upgrade/        — `upgrade` command (self-update binary)
  lib/
    utils/          — shared helpers (path display, validation, dedup)
    registry/       — RegistryProvider interface + hardcoded impl
    source/         — SourceProvider interface (git, file, archive)
    installer/      — install/remove/update orchestration
    metadata/       — sqill.json read/write store
    runtime/        — shared Runtime struct passed to all non-init subcommands
    buildinfo/      — version string injected at build time
    upgrader/       — self-update: fetch latest tag, download, replace binary
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
    "<name>": {
      "version": "...",
      "source": "...",
      "installed_at": "...",
      "description": "...",
      "tracked": true
    }
  },
  "registries": []
}
```

`tracked` on an installed entry marks its directory for inclusion in git. When `false` or absent, the skill is listed in `.agents/skills/.gitignore` (default behavior). `sqill track <name>` sets it to `true`; `sqill untrack <name>` sets it to `false`. `.gitignore` is regenerated on every `init`, `install`, `remove`, `track`, and `untrack`. On load, an old top-level `"tracked": ["<name>"]` array is migrated into the per-entry `tracked` flag.

### Registry (in binary)

```go
var defaultRegistry = map[string]SkillEntry{
    "s-regressor": {
        Name:        "s-regressor",
        Source:      "https://github.com/Q42/sqill-s-regressor.git",
        Description: "Run regression tests directly in your project.",
    },
    "q-release": {
        Name:        "q-release",
        Source:      "https://github.com/Q42/sqill-q-release.git",
        Description: "Create a GitHub release from the git diff since the last release with standardized notes.",
    },
    "read-sanity": {
        Name:        "read-sanity",
        Source:      "https://github.com/Q42/sqill-read-sanity.git",
        Description: "Read data from a sanity environment to use it for investigation and debugging.",
    },
}
```

## Source types

| Prefix                   | Handler                                      |
| ------------------------ | -------------------------------------------- |
| `git@`, `https://...git` | `GitSourceProvider` (shells out to system `git clone`) |
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