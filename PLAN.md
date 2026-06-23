# Sqill CLI — Implementation Plan

## Data Model

### Per-skill manifest

Path: `.agents/skills/<name>/sqill.json`

```json
{
  "name": "github-search",
  "version": "1.0.0",
  "description": "Search GitHub repositories"
}
```

### Unified state file

Path: `.agents/skills/sqill.json`

```json
{
  "installed": {
    "github-search": {
      "version": "1.0.0",
      "source": "https://github.com/org/github-search-skill.git",
      "installed_at": "2026-06-23T12:00:00Z"
    }
  },
  "registries": []
}
```

`registries` is reserved for future use; not wired to commands yet.

### Hardcoded registry (in binary)

```go
var defaultRegistry = map[string]string{
    "github-search": "https://github.com/org/github-search-skill.git",
    "jira":         "git@github.com:org/jira-skill.git",
    "postgres":     "file:///opt/skills/postgres",
    "example-tgz":  "https://example.com/skill.tar.gz",
}
```

## Project Structure

```
sqill/
  go.mod
  go.sum
  main.go                  // entry point, calls cmd.Execute()
  cmd/
    cmd.go                 // root command registration
    install.go
    remove.go
    update.go
    list.go
    search.go
    info.go
  internal/
    registry/
      registry.go         // RegistryProvider interface
      hardcoded.go        // HardcodedRegistryProvider impl
    source/
      provider.go         // SourceProvider interface
      git.go             // GitSourceProvider (go-git clone)
      local.go           // LocalSourceProvider (file:// copy)
      archive.go         // ArchiveSourceProvider (tar.gz download+extract)
    installer/
      installer.go        // Installer: orchestrate registry→source→install
      uninstaller.go     // Remove logic
      updater.go         // Update logic
    metadata/
      store.go           // MetadataStore: read/write sqill.json
  .agents/
    skills/
      sqill.json        // Example state file
      github-search/
        sqill.json       // Example manifest
        SKILL.md
```

## Architecture

### Interfaces

```go
type RegistryProvider interface {
    Search(query string) []SkillEntry
    Resolve(name string) (SkillEntry, error)
}

type SourceProvider interface {
    Fetch(source string, dest string) error
}

type Installer struct {
    registry RegistryProvider
    sources  map[string]SourceProvider  // keyed by source type: "git", "file", "http"
    meta     MetadataStore
}

type MetadataStore interface {
    Load() (State, error)
    Save(State) error
    IsInstalled(name string) bool
    Get(name string) (InstalledEntry, error)
    Add(name string, entry InstalledEntry) error
    Remove(name string) error
}
```

### Flow: install

```
1. Resolve skill name → registry returns source URL
2. Detect source type from URL prefix (git@/https://...git → git, file:// → local, http(s)://...tar.gz → archive)
3. Fetch to temp directory via appropriate SourceProvider
4. Validate: sqill.json exists, name matches
5. Check: target .agents/skills/<name> doesn't exist (unless --force)
6. Guard: path traversal check on name (no "../", no absolute paths)
7. Copy/clone temp → .agents/skills/<name>
8. Write metadata to .agents/skills/sqill.json
9. Clean up temp
```

### Flow: update

```
1. Resolve skill from metadata (source URL)
2. Fetch to temp directory
3. Validate manifest
4. Replace .agents/skills/<name> atomically (os.Rename)
5. Update metadata version + installed_at
```

### Flow: remove

```
1. Check installed in metadata
2. os.RemoveAll(.agents/skills/<name>)
3. Remove from metadata
```

## CLI UX

```
sqill install <name>          # Install a skill
sqill install <name> --force   # Overwrite existing
sqill remove <name>            # Remove a skill
sqill update <name>            # Update to latest
sqill list                    # List installed skills (table)
sqill search <query>          # Search hardcoded registry
sqill info <name>             # Show manifest + metadata
```

### Output formats

`sqill list`:

```
github-search  1.2.0  2026-06-23
jira           0.9.1  2026-06-18
postgres       2.0.0  2026-06-10
```

`sqill info github-search`:

```
Name:        github-search
Version:     1.2.0
Description:  Search GitHub repositories
Source:      https://github.com/org/github-search-skill.git
Installed:   2026-06-23T12:00:00Z
```

## Safety

- **Path traversal**: reject skill names containing `..`, `/`, `\`, or starting with `.`
- **Overwrite guard**: `install` fails if target exists; `--force` skips
- **Manifest validation**: `sqill.json` must exist and `name` field must match requested name
- **Atomic updates**: use `os.Rename` from temp dir to avoid partial replacements

## Implementation Order

1. `go mod init sqill` — bootstrap Go module
2. `internal/metadata/store.go` — MetadataStore (read/write sqill.json)
3. `internal/registry/hardcoded.go` — hardcoded registry + search
4. `internal/source/git.go`, `local.go`, `archive.go` — source providers
5. `internal/installer/installer.go` — install orchestration
6. `cmd/install.go`, `cmd/remove.go`, `cmd/update.go`, `cmd/list.go`, `cmd/search.go`, `cmd/info.go` — CLI
7. `main.go` + `cmd/cmd.go` — wiring
8. Unit tests per package
9. Example `.agents/skills/` with `github-search` skill and sample manifest

## Dependencies

```
github.com/spf13/cobra        # CLI framework
gopkg.in/yaml.v3              # YAML parsing (registry manifests)
github.com/go-git/go-git/v5   # Git clone support
```

## Open Questions (answered)

- **Language**: Go ✅
- **Config location**: everything under `.agents/skills/` ✅
- **Manifest file**: `sqill.json` (JSON) inside each skill ✅
- **Metadata file**: `.agents/skills/sqill.json` unified state ✅
- **Registry**: hardcoded in binary for now ✅
