# Sqill — Agent Skill Registry CLI

Sqill is a single-binary CLI tool for installing, updating, removing, and discovering agent skills — reusable packages of prompts, templates, and tools that extend AI agents.

Sqill is currently only supported for Q42 internal usage. If succesful, let's share it with the world!

## Install

```bash
go build -o /usr/local/bin/sqill .
```

## Quick start

```bash
sqill init
sqill search github
sqill install github-search
sqill list
sqill info github-search
sqill update github-search
sqill remove github-search --force
```

## Setup

All commands except `init` require `.agents/skills/sqill.json` to exist. Run `sqill init` once per project to:

1. Create `.agents/skills/` and `.agents/skills/sqill.json`. If the state file already exists, `init` reports it as already initialized and skips recreation.
2. Optionally symlink `.claude/skills`, `.cursor/skills`, and `.kilo/skills` into `.agents/skills`. `init` prompts for each; pass `--link-claude`, `--link-cursor`, `--link-kilo` to pre-select, or `--yes` to skip all prompts.

If a target like `.claude/skills` already exists as a directory, its contents are moved into `.agents/skills` and the directory is replaced with a symlink. If the existing directory contains a skill whose name already lives in `.agents/skills`, `init` refuses and tells you to de-duplicate first.

## Commands

| Command                        | Description                                        |
| ------------------------------ | -------------------------------------------------- |
| `sqill init`                   | Initialize `.agents/skills/` and optional symlinks |
| `sqill install <name>`         | Install a skill from the registry                  |
| `sqill install <name> --force` | Overwrite an existing skill                        |
| `sqill remove <name>`          | Delete an installed skill and its metadata         |
| `sqill update <name>`          | Fetch latest version and replace atomically        |
| `sqill list`                   | Show all installed skills (name, version, date)    |
| `sqill search <query>`         | Find matching skills in the registry               |
| `sqill info <name>`            | Display manifest, source, and install metadata     |

## Example directory layout

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
.claude/skills  → ../.agents/skills    (symlink created by `sqill init`)
.cursor/skills  → ../.agents/skills    (symlink created by `sqill init`)
.kilo/skills    → ../.agents/skills    (symlink created by `sqill init`)
```

All state lives under `.agents/skills/`. No databases, no daemons.

## Using Sqill as a git repository

You can manage your skills as a git repository for version control, collaboration, and CI/CD.

```bash
git init .agents/skills
git add -A
git commit -m "Add skills"
```

### Installing skills from your own git repo

```bash
sqill install my-skill --source git@github.com:your-org/my-skill.git
```

### Updating from remote

```bash
sqill update my-skill
```

### Sharing skills with your team

Add `.agents/skills/` to a team git repo so everyone gets the same skills:

```bash
git clone git@github.com:your-org/shared-skills.git .agents/skills
sqill list
```