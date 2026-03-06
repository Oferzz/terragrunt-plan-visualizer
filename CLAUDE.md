# tgviz - Claude Code Integration

## What is tgviz

`tgviz` is a CLI tool for visualizing and managing Terragrunt/Terraform plan output. When working with infrastructure-as-code in this project or any project that uses terragrunt/terraform, use `tgviz` for plan visualization, risk analysis, and apply workflows.

## Building tgviz

```bash
# Full build (React UI + Go binary)
make build

# The binary is at ./tgviz
```

## Using tgviz as a Sub-Agent

When a user asks you to plan, review, or apply infrastructure changes, follow this 5-phase workflow:

### Phase 1: Plan

Run tgviz with `--json` flag to get structured output:

```bash
tgviz plan <directory> --json --no-browser
```

Parse the JSON output from stdout. The structure is:

```json
{
  "status": "success | error | locked | no_changes",
  "plan": {
    "resource_changes": [...],
    "summary": { "total_changes", "adds", "changes", "destroys", "replaces", "high_risk", "medium_risk", "low_risk" }
  },
  "error": { "type", "message", "details", "hint" },
  "lock": { "id", "who", "operation", "created" }
}
```

### Phase 2: Analyze

Review the JSON output for:
- **Dangerous deletions**: Any resource being destroyed, especially databases, VPCs, IAM roles
- **Security issues**: Overly permissive IAM policies, open security groups (0.0.0.0/0), public S3 buckets, KMS key deletion
- **Blast radius**: How many resources are affected, especially high-risk ones
- **Unusual patterns**: Resources being replaced that shouldn't need replacement, unexpected dependencies being modified

### Phase 3: Review

Present findings to the user:
- Summary of changes (X adds, Y changes, Z destroys)
- Risk breakdown (high/medium/low counts)
- Specific concerns for high-risk changes
- Recommendation: proceed, proceed with caution, or abort

Ask the user: "Would you like me to apply these changes?"

### Phase 4: Apply

If approved:
```bash
tgviz apply <directory> --json
```

The output follows the same JSON structure with an `apply` field containing results.

### Phase 5: Verify

Read the apply result JSON and confirm:
- All expected resources were created/modified/destroyed
- No errors or partial failures occurred
- Report any failures with full error details

## Error Recovery

### State Lock
If status is `"locked"`, show the lock details to the user and ask:
"State is locked by {who} since {created}. Want me to force-unlock?"

If yes:
```bash
tgviz unlock <lock-id> <directory>
```
Then retry the original operation.

### Terraform/Terragrunt Errors
If status is `"error"`, check the error type:
- `auth`: Suggest credential refresh (e.g., `aws sso login`)
- `network`: Suggest checking VPN/network connectivity
- `timeout`: Suggest increasing timeout with `--plan-timeout` or `--apply-timeout`
- `terraform`: Show the error details, offer to analyze and fix the underlying issue

### Fix-Retry Loop
When a terraform error occurs:
1. Show the error to the user
2. Ask: "Want me to try to fix this?"
3. If yes: analyze the error, make code/config fixes, then re-run the plan
4. Repeat until success or user decides to stop

## Flags Reference

| Flag | Default | Description |
|------|---------|-------------|
| `--json` | false | JSON-only output, no web UI (use this for sub-agent mode) |
| `--no-browser` | false | Don't auto-open browser |
| `--port` | 9090 | Web UI port |
| `--plan-timeout` | 30m | Plan timeout |
| `--apply-timeout` | 60m | Apply timeout |
