# sqill

A CLI for installing and managing [agent skills](https://kilo.ai/docs) — reusable bundles of prompts, templates, and tools that extend AI agents.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/Q42/sqill/main/install.sh | sh
```

To pin a version: append `--version v0.1.0`. Falls back to `~/.local/bin` if it can't write to `/usr/local/bin`.

## Usage

```bash
sqill init                       # one-time: create .agents/skills/
sqill install github-search      # install one
sqill list                       # show installed skills
sqill info github-search         # manifest, source, install metadata
sqill update github-search       # pull latest
sqill remove github-search       # uninstall (prompts; --force to skip)
sqill install my-skill --source git@github.com:you/my-skill.git  # from any git/url
```

`init` also offers to symlink `.claude/skills`, `.cursor/skills`, and `.kilo/skills` into `.agents/skills/` so your skills are visible to every agent.

| Flag                       | Effect                                           |
| -------------------------- | ------------------------------------------------ |
| `--skills-dir <path>`      | Override the skills directory (default `.agents/skills`) |
| `--yes` / `-y`             | Skip interactive prompts                         |
| `--force`                  | Overwrite an existing install                    |
| `--source <url>`           | Install from a specific git/file/archive URL     |

## Build a skill

A skill is just a directory containing a `sqill.json` manifest. The minimum:

```json
{
  "name": "my-skill",
  "version": "0.1.0",
  "description": "What it does, in one line."
}
```

Add any files you want next to it — `SKILL.md` is conventional for the prompt/instructions your agent should read. The directory name must match the `name`.

Host it anywhere `sqill` knows how to fetch from:

| Source                      | Notes                                |
| --------------------------- | ------------------------------------ |
| `https://.../repo.git`      | Cloned with go-git                   |
| `git@github.com:.../repo.git` | Same, over SSH                     |
| `file:///path/to/skill`     | Copied locally                       |
| `https://.../skill.tar.gz`  | Downloaded and extracted             |

To add a skill to the built-in registry, edit `src/lib/registry/hardcoded.go` and open a PR.

## Develop

Requires Go 1.25+.

```bash
go build -o sqill .
go test ./...
```

To cut a release: `git tag v0.1.0 && git push --tags`. `.github/workflows/release.yml` builds and publishes binaries for macOS and Linux (amd64, arm64).
