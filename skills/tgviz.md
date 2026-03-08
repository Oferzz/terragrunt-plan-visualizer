---
name: tgviz-infrastructure
description: Use when user asks to plan, review, apply, or manage terragrunt/terraform infrastructure changes, run infra plans, check what changed, or deploy infrastructure
---

# tgviz Infrastructure Management

## Overview

Use `tgviz` CLI to plan, analyze, and apply terragrunt/terraform changes. **Always dispatch a subagent** to run tgviz and analyze output — this keeps the large JSON out of the main conversation context.

## Phase 0: Ensure tgviz Installed

Run in subagent or main session:

```bash
which tgviz || {
  echo "Installing tgviz..." >&2
  git clone https://github.com/Oferzz/terragrunt-plan-visualizer.git /tmp/tgviz-build
  cd /tmp/tgviz-build
  make build
  mkdir -p ~/.local/bin
  mv tgviz ~/.local/bin/
  rm -rf /tmp/tgviz-build
  export PATH="$HOME/.local/bin:$PATH"
  echo "tgviz installed to ~/.local/bin/tgviz" >&2
}
```

If `make build` fails (missing Node.js), fall back to Go-only build:
```bash
cd /tmp/tgviz-build && mkdir -p internal/server/static && echo '<html><body>tgviz</body></html>' > internal/server/static/index.html && go build -o tgviz ./cmd/tgviz
```

Warn user if `~/.local/bin` is not in PATH.

## Phase 1: Plan via Subagent

**Dispatch a subagent** with this prompt pattern:

```
Run tgviz plan on <directory> and analyze the results.

1. Run: tgviz plan <directory> --json --no-browser --analyze --detail high
2. Parse the JSON output from stdout
3. Check for:
   - Status: success/error/locked/no_changes
   - Dangerous deletions (databases, VPCs, IAM roles being destroyed)
   - Security issues (open security groups, public S3 buckets, permissive IAM)
   - Blast radius (high_risk count, total destroys)
   - Drift (feature.unrelated_count > 0 means changes unrelated to code diff)
4. Return a concise summary:
   - One line: "X adds, Y changes, Z destroys (H high-risk, M medium, L low)"
   - List each high-risk change with address and reason
   - List any unrelated/drift changes
   - List any findings from analysis.findings
   - Your recommendation: PROCEED / PROCEED WITH CAUTION / ABORT
   - If status is "locked", return lock details (id, who, created)
   - If status is "error", return error type, message, and hint
```

The subagent runs tgviz, processes the full JSON, and returns only a concise summary to the main session.

### JSON Output Structure

```json
{
  "status": "success | error | locked | no_changes",
  "plan": {
    "resource_changes": [{ "address": "...", "action": "create|update|delete|replace", "risk_level": "high|medium|low", "risk_reasons": [...] }],
    "summary": { "total_changes": 5, "adds": 2, "changes": 2, "destroys": 1, "high_risk": 1, "medium_risk": 2, "low_risk": 2 }
  },
  "analysis": { "findings": [...], "recommendations": [...] },
  "feature": { "expected_count": 3, "unrelated_count": 1 }
}
```

## Phase 2: Present to User

Using the subagent's summary, present to the user:
- Change counts and risk breakdown
- Specific concerns for high-risk changes
- Drift warnings if any
- Recommendation: **proceed**, **proceed with caution**, or **abort**

Ask: "Would you like me to apply these changes?"

## Phase 3: Apply via Subagent

Only if user approves. **Dispatch a subagent:**

```
Run tgviz apply on <directory> and report results.

1. Run: tgviz apply <directory> --json
2. Parse the JSON output
3. Return: success/failure, resources created/modified/destroyed, any errors
```

## Phase 4: Error Recovery

### State Lock
If status is `"locked"`, show lock details and ask user. If approved:
```bash
tgviz unlock <lock-id> <directory>
```
Then retry the plan.

### Auth Error
Suggest: `aws sso login` or credential refresh.

### Network Error
Suggest: check VPN/network connectivity.

### Terraform Error
Show error details, offer to analyze and fix the underlying code/config, then re-plan.

## Flags Reference

| Flag | Default | Description |
|------|---------|-------------|
| `--json` | false | JSON output for automation |
| `--no-browser` | false | Don't open web UI |
| `--analyze` | false | Include risk analysis report |
| `--detail` | low | Analysis detail: low, medium, high |
| `--port` | 9090 | Web UI port |
| `--plan-timeout` | 30m | Plan timeout |
| `--apply-timeout` | 60m | Apply timeout |
| `--feature-aware` | true | Git diff correlation |
| `--base-branch` | main | Base branch for diff |
