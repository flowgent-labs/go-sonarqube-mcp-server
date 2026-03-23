package mcptools

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
)

func NewToggleAutomaticAnalysisMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"toggle_automatic_analysis",
		"Toggle SonarQube for IDE Automatic Analysis — Enable or disable SonarQube for IDE automatic analysis. When enabled, SonarQube for IDE will automatically analyze files as they are modified in the working directory. When disabled, automatic analysis is turned off.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"enabled": {"type": "boolean", "description": "Enable or disable the automatic analysis."}
			},
			"required": ["enabled"],
			"additionalProperties": false
	}`))
}

func ToggleAutomaticAnalysisHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError(
		"toggle_automatic_analysis requires SonarQube for IDE (IDE bridge). " +
			"Please ensure SonarQube for IDE is running and connect via the official SonarQube MCP Server. " +
			"The standalone Go MCP server does not support IDE bridge integration."), nil
}
