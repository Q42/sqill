# AGENTS.md — Agent Instructions for Sqill

- `README.md` is for **end users only**: install, usage, build, source layout. Never add contributor workflow, release process, agent internals, or internal conventions there.
- `AGENTS.md` (this file) is the source of truth for development workflow and agent behavior. Update it whenever a new convention is introduced.
- `ARCHITECTURE.md` documents the project structure and source layout. Update it when the code changes in a way that affects the structural overview.
- For any **user-facing behavior** change, also update `README.md`.

## Project overview

Sqill is a Go CLI tool: single static binary, no database, no daemons. It installs skills into `.agents/skills/`, tracks metadata in a JSON state file, and resolves skill sources from a hardcoded registry.

## Architecture

- **Source types**: git (clone via system `git`), local filesystem copy, archive download — auto-detected from the source URL.
- **Registry**: hardcoded map of skill name → metadata (source URL, description). Resides in the binary; no external registry.
- **Tracking**: skills are gitignored by default; `track`/`untrack` toggle git inclusion via a per-skill flag. `.gitignore` is regenerated on mutating operations.
- Build/test: standard Go tooling (`go build`, `go test`, `go vet`).

## Commands

| Command   | Signature        | Purpose                                          |
| --------- | ---------------- | ------------------------------------------------ |
| `init`    | `init`           | Create `.agents/skills/` and optionally symlink agent tool dirs |
| `install` | `install [name]` | Install a skill from the registry, or all listed in state |
| `remove`  | `remove <name>`  | Delete an installed skill and its metadata        |
| `update`  | `update <name>`  | Fetch latest version and replace atomically       |
| `list`    | `list`           | Show all installed skills                         |
| `info`    | `info <name>`    | Display manifest, source, and install metadata    |
| `track`   | `track <name>`   | Include a skill's directory in git                |
| `untrack` | `untrack <name>` | Exclude a skill's directory from git              |
| `upgrade` | `upgrade`        | Self-update `sqill` binary to latest release      |

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
  }
}
```

## Agent rules

- **No comments** in code unless strictly necessary.
- Errors are returned, never panicked (except fatal startup errors in `main.go`).
- **Do not touch the git state.** The agent is not allowed to run `git add`, `git commit`, `git push`, or `git tag`. The user manages the git state.
  - **Exception:** the `q-release` skill may use `git add`, `git commit`, `git push`, and `git tag` — only for release notes and the release tag.

## Safety invariants

1. Skill names must not contain `..`, `/`, `\`, or start with `.`
2. A skill's manifest (`sqill.json`) must exist and its `name` field must match the directory name
3. State writes are atomic (write to temp file, rename into place)
4. Everything lives under `.agents/skills/` — nothing stored outside

## Release workflow

Releases are tag-triggered via CI. The release body comes from a curated notes file in the repo (produced by the `q-release` skill), not from GitHub's auto-generator.

- Notes are committed on the default branch before the tag is created.
- The tag push triggers the CI workflow which builds binaries and publishes the release.
- Never use `generate_release_notes: true` on the workflow action.
- Never run `gh release create` after pushing the tag — the workflow owns release creation.