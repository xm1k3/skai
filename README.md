<p align="center">
  <img src="assets/skai-logo.svg" alt="skai" width="360">
</p>

<p align="center">
  Community-driven aggregator and installer for Agent Skills.
</p>

<p align="center">
  <a href="https://golang.org"><img src="https://img.shields.io/badge/made%20with-Go-00ADD8.svg"></a>
  <a href="https://github.com/xm1k3/skai/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-green.svg"></a>
  <a href="https://github.com/xm1k3/skai/releases"><img src="https://img.shields.io/github/v/release/xm1k3/skai"></a>
</p>

<p align="center">
  <img src="assets/demo.gif" alt="skai demo" width="760">
</p>

---

`skai` (SKill kit for AI) collects [Agent Skills](https://docs.claude.com/en/docs/agents-and-tools/agent-skills) scattered across dozens of community repositories into a single searchable catalog, runs static risk analysis on every skill, and installs them into Claude Code, Codex CLI or claude.ai with one command.

## Features

- Aggregates skills from any number of git repositories defined in a simple `sources.yaml`
- Builds a flat JSON index (`~/.skai/index.json`) with per-skill metadata and risk flags
- **Static analysis only** — scripts are never executed during `sync`, `validate` or `list`
- Risk-aware install with an explicit summary (network calls, destructive ops) before copying
- Installs to Claude Code, Codex CLI (copy or symlink) or exports a zip for claude.ai
- Fuzzy search, category / risk / tool filters, content-hash deduplication and catalog stats
- `export-awesome-list` regenerates an awesome-list markdown from the whole catalog

## Installation

```bash
go install github.com/xm1k3/skai@latest
```

Or grab a static binary from the [releases](https://github.com/xm1k3/skai/releases) page.

Requires the `git` binary in `PATH`.

## Quick start

```bash
skai init          # write ~/.skai/sources.yaml with the default sources
skai sync          # clone/pull every enabled source and build the index
skai search pdf    # find a skill
skai info pdf      # inspect its metadata and risk flags
skai install pdf --target claude-code --project
```

## Examples

### Sync the sources

```console
$ skai sync
composio-awesome-claude-skills           cloned, 864 skills
sickn33-antigravity-awesome-skills       cloned, 5996 skills
alirezarezvani-claude-skills             cloned, 758 skills
behisecc-awesome-claude-skills           cloned, 0 skills
travisvn-awesome-claude-skills           cloned, 0 skills

Indexed 7618 skills from 5 sources into /home/user/.skai/index.json
```

### Search the catalog

```console
$ skai search pdf
NAME                 CATEGORY   RISK    SOURCE                          DESCRIPTION
pdf                  documents  medium  composio-awesome-claude-skills  Comprehensive PDF manipulation toolkit for extra...
api2pdf-automation   documents  medium  composio-awesome-claude-skills  Automate Api2pdf tasks via Rube MCP (Composio). ...
craftmypdf-automation documents medium  composio-awesome-claude-skills  Automate Craftmypdf tasks via Rube MCP (Composio...
pdf-co-automation    documents  medium  composio-awesome-claude-skills  Automate PDF co tasks via Rube MCP (Composio). A...

4 skills
```

### Filter by category and risk

```console
$ skai list --category security --risk high
NAME                    CATEGORY  RISK  SOURCE                              DESCRIPTION
007                     security  high  sickn33-antigravity-awesome-skills  Security audit, hardening, threat modeling (STRI...
api-fuzzing-bug-bounty  security  high  sickn33-antigravity-awesome-skills  Provide comprehensive techniques for testing RES...
```

### Inspect a skill

```console
$ skai info pdf
Name:                        pdf
Description:                 Comprehensive PDF manipulation toolkit for extracting text and tables, creating new PDFs, merging/splitting documents, and handling forms.
Category:                    documents
Source:                      composio-awesome-claude-skills
Repository:                  https://github.com/ComposioHQ/awesome-claude-skills
Path:                        document-skills/pdf
Last commit:                 92568c1edaff (2026-05-22T08:47:49+05:30)
Risk level:                  medium
Network calls:               yes
Destructive ops:             no
Confirms before destructive: no
Has scripts:                 yes
Claude Code only:            no
Extra frontmatter fields:    license
Lines:                       295
Token estimate:              1767
Content hash:                38d8559d4899
```

### Install with a risk summary

Every install prints the risk summary and asks for confirmation before copying
(skip the prompt with `--yes`, the summary is still printed):

```console
$ skai install pdf --target claude-code --project
Skill:                       pdf (composio-awesome-claude-skills)
Destination:                 .claude/skills/pdf
Risk level:                  medium
Network calls:               yes
Destructive ops:             no
Confirms before destructive: no
Has scripts:                 yes
Proceed with install? [y/N]: y
Installed pdf (copied) to .claude/skills/pdf
```

### Validate a skill

```console
$ skai validate pdf
composio-awesome-claude-skills/pdf
  WARN  description is 252 characters, agents work best under 200

Validated 1 skills: 0 errors, 1 warnings
```

### Catalog statistics

```console
$ skai stats
Total skills: 3161
Last sync:    2026-07-06 21:40:26 UTC

BY CATEGORY    COUNT
productivity   895
uncategorized  610
web            384
security       272
devops         234
...

BY RISK  COUNT
medium   1672
low      1227
high     262

FLAGS             COUNT
has scripts       1058
network calls     1798
destructive ops   427
claude code only  84
```

## Commands

| Command                    | Description                                                    |
|----------------------------|----------------------------------------------------------------|
| `skai init`                | Create `~/.skai/sources.yaml` with the default sources         |
| `skai sync`                | Clone or pull every enabled source and rebuild the index       |
| `skai list`                | List indexed skills with filters                               |
| `skai search <query>`      | Fuzzy search over name and description                         |
| `skai validate [skill]`    | Validate frontmatter, referenced paths and length              |
| `skai info <skill>`        | Show full metadata, risk flags, source and last commit         |
| `skai install <skill>`     | Install into a target with a risk summary and confirmation     |
| `skai uninstall <skill>`   | Remove an installed skill                                      |
| `skai dedupe`              | Remove duplicate skills from the index by content hash         |
| `skai stats`               | Show counts by category, source and risk                       |
| `skai export-awesome-list` | Generate an `awesome-list.md` from the catalog                 |

### list filters

```bash
skai list --category security --risk high
skai list --tool codex           # hides Claude Code only skills
skai list --has-scripts --network
skai list --destructive
```

## Install targets

| Target        | Personal                    | Project                                        |
|---------------|-----------------------------|------------------------------------------------|
| `claude-code` | `~/.claude/skills/<name>/`  | `.claude/skills/<name>/`                       |
| `codex`       | `~/.codex/skills/<name>/`   | `.codex/skills/<name>/`                         |
| `web`         | n/a                         | `./skai-exports/<name>.zip` for manual upload  |

```bash
skai install pdf --target claude-code --personal
skai install pdf --target codex --project --link   # symlink, future syncs propagate
skai install pdf --target web                       # zip for claude.ai upload
```

- `--personal` (default) installs into the user directory, `--project` into the current repo
- `--link` symlinks instead of copying so future `skai sync` updates propagate automatically
- A name conflict with a skill already installed from a different source warns and requires `--force`
- `--yes` skips the confirmation prompt (the risk summary is still printed)

## Risk analysis

For every skill `skai` records, from static analysis only:

- `has_scripts` — ships a `scripts/` directory or executable code blocks in the body
- `network_calls` — references `curl`, `wget`, `fetch(` or an `http(s)://` URL
- `destructive_ops` — references `rm`, `mv`, `DROP TABLE`, `delete`
- `confirms_before_destructive` — pairs a destructive op with a confirmation pattern
- `claude_code_only` — uses Claude Code specific frontmatter (`allowed-tools`, `context`, `hooks`)

Risk level is derived as `high` (destructive without confirmation), `medium` (destructive or network), or `low`.

## sources.yaml

`skai init` writes the following defaults; add, remove or disable entries freely:

```yaml
sources:
  - name: composio-awesome-claude-skills
    repo: https://github.com/ComposioHQ/awesome-claude-skills
    enabled: true
  - name: sickn33-antigravity-awesome-skills
    repo: https://github.com/sickn33/antigravity-awesome-skills
    enabled: true
  - name: alirezarezvani-claude-skills
    repo: https://github.com/alirezarezvani/claude-skills
    enabled: true
```

## How indexing works

For each enabled source `skai` clones (or pulls) the repository under `~/.skai/sources/<name>`, walks the tree, and indexes every directory containing a valid `SKILL.md` (frontmatter with at least `name` and `description`). Skills are stored in `~/.skai/index.json` together with their metadata, risk flags, last commit hash and date, and a content hash used for deduplication.

## License

Distributed under the Apache License 2.0. See [LICENSE](LICENSE).
