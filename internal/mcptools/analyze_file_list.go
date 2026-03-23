package mcptools

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
)

func NewAnalyzeFileListMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"analyze_file_list",
		"SonarQube for IDE File Analysis — Analyze files in the current working directory using SonarQube for IDE. This tool connects to a running SonarQube for IDE instance to perform code quality analysis on a list of files.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"file_absolute_paths": {"type": "array", "items": {"type": "string"}, "description": "List of absolute file paths to analyze."}
			},
			"required": ["file_absolute_paths"],
			"additionalProperties": false
	}`))
}

func AnalyzeFileListHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError(
		"analyze_file_list requires SonarQube for IDE (IDE bridge). " +
			"Please ensure SonarQube for IDE is running and connect via the official SonarQube MCP Server. " +
			"The standalone Go MCP server does not support IDE bridge integration."), nil
}
