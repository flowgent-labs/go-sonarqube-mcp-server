# sonarqube-mcp

A Go MCP server for SonarQube, providing AI agents with access to code quality, security, and project management capabilities.

## Quick Start

### Build from Source

```sh
make
```

### Usage Example

- https://sonarqube.example.com/component_measures?id=rengine&metric=duplicated_lines&selected=rengine%3Aeventbus%2Feventbus-base%2Fsrc%2Fmain%2Fjava%2Fcom%2Fwl4g%2Frengine%2Feventbus%2Frecorder%2FRedisEventRecorder.java

```bash
export SONARQUBE_URL=https://sonarqube.example.com
export SONARQUBE_TOKEN=squ_xxxx
./bin/sonarqube-mcp -t cli list
./bin/sonarqube-mcp -t cli agg_search_issue_summary --branch main --ps 20 --key rengine:eventbus/eventbus-base/src/main/java/com/wl4g/rengine/eventbus/recorder/RedisEventRecorder.java
```

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `SONARQUBE_URL` | Conditional | SonarQube Server URL (e.g. `https://sonarqube.example.com`). Not needed if `SONARQUBE_ORG` is set. |
| `SONARQUBE_TOKEN` | Yes (stdio) | SonarQube user token. In HTTP mode, can be provided via `Authorization: Bearer` header. |
| `SONARQUBE_TOKEN_FILE` | No | Path to a file containing the token (for Docker/Kubernetes secrets). |
| `SONARQUBE_ORG` | No | SonarQube Cloud organization key. Setting this implies Cloud mode. |
| `SONARQUBE_PROJECT_KEY` | No | Default project key used as fallback when tools omit `projectKey`. |
| `SONARQUBE_TOOLSETS` | No | Comma-separated tool categories to enable. Defaults to all. |
| `SONARQUBE_READ_ONLY` | No | Set to `true` to expose only read-only tools. |

## Token Configuration

The server retrieves the SonarQube token using the following priority:

1. `SONARQUBE_TOKEN` environment variable
2. `SONARQUBE_TOKEN_FILE` (read token from file, for Kubernetes secrets etc.)
3. macOS Keychain / Windows Credential Manager

## Agent Integration

### Claude Code

`~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "sonarqube": {
      "command": "./sonarqube-mcp",
      "args": ["--transport", "stdio"],
      "env": {
        "SONARQUBE_URL": "https://sonarqube.example.com",
        "SONARQUBE_TOKEN": "your-token"
      }
    }
  }
}
```

### Codex CLI

`~/.codex/config.yaml`:

```yaml
mcp:
  servers:
    sonarqube:
      command: ./sonarqube-mcp
      args: ["--transport", "stdio"]
      env:
        SONARQUBE_URL: https://sonarqube.example.com
        SONARQUBE_TOKEN: your-token
```

### Remote Mode (HTTP)

Start the server:

```sh
export SONARQUBE_URL=https://sonarqube.example.com
export SONARQUBE_TOKEN=your-token
./sonarqube-mcp --transport http --port 8080
```

Claude Code (remote):

```json
{
  "mcpServers": {
    "sonarqube": {
      "url": "http://localhost:8080/mcp",
      "env": {
        "SONARQUBE_URL": "https://sonarqube.example.com",
        "SONARQUBE_TOKEN": "your-token"
      }
    }
  }
}
```

## Tools Reference

### Projects
- `search_my_sonarqube_projects` — Find projects by name or key
- `list_branches` — List long-lived branches for a project
- `list_pull_requests` — List pull requests for a project

### Issues
- `search_sonar_issues_in_projects` — Search issues with filters for severity, status, impact
- `change_sonar_issue_status` — Change issue status (accept/falsepositive/reopen)

### Security Hotspots
- `search_security_hotspots` — Search security hotspots
- `show_security_hotspot` — Get detailed hotspot information
- `change_security_hotspot_status` — Change hotspot status and resolution

### Quality Gates
- `list_quality_gates` — List all quality gates
- `get_project_quality_gate_status` — Get project quality gate status

### Measures & Coverage
- `get_component_measures` — Get project metrics (ncloc, complexity, coverage, etc.)
- `search_files_by_coverage` — Find files with low coverage
- `get_file_coverage_details` — Line-by-line coverage details

### Duplications
- `get_duplications` — Get duplications for a file
- `search_duplicated_files` — Find duplicated files in a project

### Rules & Languages
- `show_rule` — Show rule details
- `list_languages` — List supported languages
- `search_metrics` — Search available metrics

### System
- `ping_system` — Ping the instance
- `get_system_status` — Get system status and version
- `get_system_health` — Get health status with causes
- `get_system_info` — Get detailed system configuration
- `get_system_logs` — Get server logs

### Portfolios
- `list_portfolios` — List portfolios/views

### Sources
- `get_raw_source` — Get file source code
- `get_scm_info` — Get SCM blame information

### Webhooks
- `list_webhooks` — List webhooks
- `create_webhook` — Create a new webhook

### Dependency Risks
- `search_dependency_risks` — Search SCA/dependency risk issues

## SonarQube Cloud

To connect to SonarQube Cloud, set `SONARQUBE_ORG` instead of `SONARQUBE_URL`:

```sh
export SONARQUBE_ORG=my-organization
export SONARQUBE_TOKEN=your-cloud-token
./sonarqube-mcp --transport stdio
```

## Read-Only Mode

Restrict to read-only tools:

```sh
export SONARQUBE_READ_ONLY=true
```

This disables `change_sonar_issue_status`, `change_security_hotspot_status`, and `create_webhook`.

## CLI Mode

List tools:
```sh
./sonarqube-mcp -t cli list
```

Call a tool:
```sh
./sonarqube-mcp -t cli call search_my_sonarqube_projects q=myproject
./sonarqube-mcp -t cli call get_component_measures projectKey=my_project metricKeys=ncloc,coverage,violations
```
