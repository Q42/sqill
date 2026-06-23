# Sqill — Agent Skill Registry CLI

Sqill is a single-binary CLI tool for installing, updating, removing, and discovering agent skills — reusable packages of prompts, templates, and tools that extend AI agents.

Sqill is currently only supported for Q42 internal usage. If succesful, let's share it with the world!

## Install

Temporary.

```bash
go build -o /usr/local/bin/sqill .
```

## Quick start

```bash
sqill setup
sqill search github
sqill install github-search
sqill list
sqill info github-search
sqill update github-search
sqill remove github-search --force
```

## Setup

All commands except `setup` require `.agents/skills/sqill.json` to exist. Run `sqill setup` once per project to:

1. Create `.agents/skills/` and `.agents/skills/sqill.json`.
2. Optionally symlink `.claude/skills`, `.cursor/skills`, and `.kilo/skills` into `.agents/skills`. Setup prompts for each; pass `--link-claude`, `--link-cursor`, `--link-kilo` to pre-select, or `--yes` to skip all prompts.

If a target like `.claude/skills` already exists as a directory, its contents are moved into `.agents/skills` and the directory is replaced with a symlink. If the existing directory contains a skill whose name already lives in `.agents/skills`, setup refuses and tells you to de-duplicate first.

## Commands

| Command                        | Description                                        |
| ------------------------------ | -------------------------------------------------- |
| `sqill setup`                  | Initialize `.agents/skills/` and optional symlinks |
| `sqill install <name>`         | Install a skill from the registry                  |
| `sqill install <name> --force` | Overwrite an existing skill                        |
| `sqill remove <name>`          | Delete an installed skill and its metadata         |
| `sqill update <name>`          | Fetch latest version and replace atomically        |
| `sqill list`                   | Show all installed skills (name, version, date)    |
| `sqill search <query>`         | Find matching skills in the registry               |
| `sqill info <name>`            | Display manifest, source, and install metadata     |

## Example directory Layout

```
.agents/
  skills/
    sqill.json               ← installed metadata (versions, sources, timestamps)
    github-search/            ← installed skill
      sqill.json              ← skill manifest (name, version, description)
      SKILL.md
    jira/
      sqill.json
    postgres/
      sqill.json
.claude/skills  → ../.agents/skills    (symlink created by `sqill setup`)
.cursor/skills  → ../.agents/skills    (symlink created by `sqill setup`)
.kilo/skills    → ../.agents/skills    (symlink created by `sqill setup`)
```

All state lives under `.agents/skills/`. No databases, no daemons.

## Registry

The registry is the catalog mapping skill names to sources. Currently hardcoded in the binary.

## Architecture

```text
cmd/          — cobra CLI commands
internal/
  registry/   — RegistryProvider interface + hardcoded impl
  source/     — SourceProvider interface (git, file, archive)
  installer/  — install/remove/update orchestration
  metadata/   — sqill.json read/write store
```
