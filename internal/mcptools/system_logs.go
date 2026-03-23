package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

type SystemLogsToolResponse struct {
	LogType string `json:"logType"`
	Content string `json:"content"`
}

func NewSystemLogsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"get_system_logs",
		"Get System Logs — Get SonarQube Server system logs in plain text. Requires Administer permission.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"name": {"type": "string", "enum": ["access", "app", "ce", "deprecation", "es", "web"], "description": "Log file name. Defaults to \"app\".", "default": "app"}
			},
			"additionalProperties": false
	}`))
}

func SystemLogsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	logName := mcputils.GetOptionalString(args, "name")
	if logName == "" {
		logName = "app"
	}

	params := url.Values{}
	params.Set("name", logName)

	client := mcputils.NewSQClient()
	raw, err := client.DoGetRaw(ctx, "/api/system/logs", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Get system logs failed: %v", err)), nil
	}

	if raw == "" {
		raw = "No logs available."
	}

	response := SystemLogsToolResponse{
		LogType: logName,
		Content: raw,
	}
	respJSON, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(respJSON)), nil
}
