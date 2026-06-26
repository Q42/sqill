# sqill

> **⚠️ The registry of installable skills is hardcoded into the binary.**
> Adding a new skill today requires editing `src/lib/registry/hardcoded.go` and shipping a new release of sqill itself. There is no external registry, no pluggable source, and no way for end users to add skills to the catalog without recompiling. Skills can still be installed ad hoc via `sqill install <name> --source <url>`, but the curated registry ships in-binary.

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

## How does it work

Sqill is a thin wrapper around standard tools — it does no networking or authentication of its own.

- **Git sources** are cloned by shelling out to your system `git`. SSH keys, HTTPS credentials, host-key checks, proxies, and 2FA are all handled by git itself, using whatever is already configured on your machine (`~/.ssh/config`, `~/.gitconfig`, credential helpers, `gh auth login`, etc.).
- **Local sources** are copied straight from the filesystem.
- **Archive sources** are downloaded over HTTPS with Go's standard library.

Because access control is delegated to git, nothing extra is required to install skills from private repos: if `git clone <url>` works in your shell, `sqill install <name>` will work, and any auth prompt (SSH passphrase, GitHub device login, credential helper, …) is the same one git would have shown you directly.

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
| `https://.../repo.git`      | Cloned via system `git`              |
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
