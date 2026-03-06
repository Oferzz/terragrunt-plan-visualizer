# tgviz - Terragrunt Plan Visualizer

A CLI tool that provides rich web-based visualization of Terragrunt/Terraform plan output, with built-in risk analysis and Claude Code integration for AI-powered security review.

## Features

- **Web UI**: Dark-themed dashboard showing plan changes with attribute-level diffs
- **Risk Analysis**: Built-in heuristics categorize changes as high/medium/low risk
- **Claude Code Integration**: Structured JSON output for AI-powered security analysis
- **Apply Flow**: Approve and apply plans with real-time WebSocket streaming
- **Error Recovery**: Structured error handling with state lock detection and unlock
- **Single Binary**: Go binary with embedded React UI, no external dependencies

## Installation

```bash
# Build from source
make build

# Or build Go binary directly (requires pre-built web assets)
go build -o tgviz ./cmd/tgviz
```

## Usage

### Plan and Visualize

```bash
# Run terragrunt plan and open web UI
tgviz plan ./module

# JSON-only output (no web UI) for Claude Code
tgviz plan ./module --json

# Pass extra args to terragrunt
tgviz plan ./module -- -var="env=prod"
```

### Show Existing Plan

```bash
# Visualize a terraform show -json output file
tgviz show plan.json
```

### Apply

```bash
# Apply a previously created plan
tgviz apply ./module
```

### Unlock State

```bash
# Force-unlock terraform state
tgviz unlock <lock-id> ./module
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | 9090 | Web UI port |
| `--no-browser` | false | Don't auto-open browser |
| `--plan-timeout` | 30m | Plan execution timeout |
| `--apply-timeout` | 60m | Apply execution timeout |
| `--json` | false | JSON-only output, no web UI |

## Claude Code Integration

When used as a sub-agent from Claude Code, tgviz outputs structured JSON to stdout:

```bash
# Claude runs plan, reads JSON output
tgviz plan ./module --json

# Claude analyzes and decides to apply
tgviz apply ./module --json

# Claude handles lock errors
tgviz unlock <lock-id> ./module
```

### JSON Output Format

```json
{
  "status": "success",
  "plan": {
    "resource_changes": [...],
    "summary": {
      "total_changes": 5,
      "adds": 2,
      "changes": 2,
      "destroys": 1,
      "high_risk": 1,
      "medium_risk": 2,
      "low_risk": 2
    }
  }
}
```

Status values: `success`, `error`, `locked`, `no_changes`

## Risk Heuristics

| Level | Triggers |
|-------|----------|
| High | Any destroy; IAM changes; security group changes; RDS/database mods; KMS key changes; VPC/subnet deletion |
| Medium | EC2 instance mods; ECS/Lambda changes; network changes; auto-scaling changes |
| Low | Tag-only changes; description updates; new non-sensitive resource additions |

## Development

```bash
# Install dependencies
cd web && npm install && cd ..
go mod download

# Run React dev server (with API proxy to :9090)
make dev

# Run tests
make test

# Full build (React + Go)
make build

# Clean
make clean
```

## Architecture

```
tgviz CLI (Go binary)
├── runs terragrunt plan/apply (subprocess)
├── parses JSON plan output
├── categorizes risk (built-in heuristics)
├── serves React UI on localhost:9090
├── WebSocket for real-time apply streaming
└── outputs structured JSON to stdout
```

## Project Structure

```
cmd/tgviz/          CLI entry point (cobra)
internal/
├── plan/           Plan runner, parser, models
├── risk/           Risk heuristic engine
├── server/         HTTP server + WebSocket + embedded React
└── output/         Structured JSON output
web/                React app (Vite + TypeScript)
```
