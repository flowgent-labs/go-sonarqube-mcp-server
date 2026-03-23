package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	mcputils "sonarqube-mcp/internal/helpers"
	"sonarqube-mcp/internal/mcptools"
)

// requestLoggerMiddleware logs MCP tool requests and responses
func requestLoggerMiddleware(next server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, err error) {
		start := time.Now()
		v := mcputils.GetVerbosity()
		sessionID := mcputils.GetSessionID(ctx)
		ts := start.Format(time.RFC3339)

		if v >= 2 {
			argsJSON, _ := json.Marshal(request.GetArguments())
			sid := sessionID
			if sid == "" {
				sid = "-"
			}
			fmt.Fprintf(os.Stderr, "%s [mcp] sid=%s request tool=%s args=%s\n", ts, sid, request.Params.Name, string(argsJSON))
		}

		defer func() {
			duration := time.Since(start).Round(time.Millisecond)
			sid := sessionID
			if sid == "" {
				sid = "-"
			}
			if v >= 1 {
				status := 200
				if err != nil {
					status = 500
				} else if result != nil && result.IsError {
					status = 500
				}
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s [mcp] sid=%s %d %s (%s) error=%v\n", time.Now().Format(time.RFC3339), sid, status, request.Params.Name, duration, err)
				} else {
					fmt.Fprintf(os.Stderr, "%s [mcp] sid=%s %d %s (%s)\n", time.Now().Format(time.RFC3339), sid, status, request.Params.Name, duration)
				}
			}
			if v >= 2 && result != nil {
				resultText := ""
				for _, c := range result.Content {
					if tc, ok := c.(mcp.TextContent); ok {
						resultText = tc.Text
						break
					}
				}
				fmt.Fprintf(os.Stderr, "%s [mcp] response body=%s\n", time.Now().Format(time.RFC3339), truncate(resultText, 200))
			}
		}()

		return next(ctx, request)
	}
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}

// systemToolNames lists tool names that are only available on SonarQube Server.
var systemToolNames = map[string]bool{
	"ping_system":       true,
	"get_system_status": true,
	"get_system_health": true,
	"get_system_info":   true,
	"get_system_logs":   true,
}

// cloudOnlyToolNames lists tool names that are only available on SonarQube Cloud.
var cloudOnlyToolNames = map[string]bool{
	"list_enterprises":          true,
	"run_advanced_code_analysis": true,
}

// NewMCPServer creates and returns an MCP server with all tools registered
func NewMCPServer() *server.MCPServer {
	s := server.NewMCPServer(
		"sonarqube-mcp",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithLogging(),
		server.WithToolHandlerMiddleware(requestLoggerMiddleware),
	)

	isCloud := mcputils.IsCloud()

	for name, entry := range mcptools.Registry {
		// System tools only on Server
		if systemToolNames[name] && isCloud {
			continue
		}
		// Cloud-only tools
		if cloudOnlyToolNames[name] && !isCloud {
			continue
		}
		s.AddTool(entry.Tool, entry.Handler)
	}

	return s
}
