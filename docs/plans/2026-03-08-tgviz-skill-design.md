# tgviz Claude Code Skill Design

## Problem

Users need to copy CLAUDE.md instructions into every project to use tgviz with Claude Code. A global skill is more portable and triggers automatically.

## Design

**Location:** `~/.claude/skills/tgviz.md`

**Trigger:** When user asks to plan, review, apply, or manage infrastructure with terragrunt/terraform.

**Workflow:**

1. **Ensure tgviz installed** — `which tgviz`, if missing: clone from GitHub, `make build`, install to `~/.local/bin/`
2. **Plan** — `tgviz plan <dir> --json --no-browser --analyze`
3. **Analyze** — dangerous deletions, security issues, drift, blast radius
4. **Present** — summary, risk breakdown, recommendation
5. **Apply** (if approved) — `tgviz apply <dir> --json`, verify results
6. **Error recovery** — lock (offer unlock), auth (suggest creds refresh), network errors

**Skill type:** Rigid process — follow phases exactly.

**Auto-install:** Build from source to `~/.local/bin/` (no sudo). Warn if not in PATH.
