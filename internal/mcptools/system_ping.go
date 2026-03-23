package mcptools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

type SystemPingToolResponse struct {
	Response string `json:"response"`
}

func NewSystemPingMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ping_system",
		`Ping System — Ping the SonarQube Server to check if it is alive. Returns "pong" on success.`,
		json.RawMessage(`{
			"type": "object",
			"properties": {},
			"additionalProperties": false
		}`),
	)
}

func SystemPingHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := mcputils.NewSQClient()
	raw, err := client.DoGetRaw(ctx, "/api/system/ping", nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("System ping failed: %v", err)), nil
	}

	respJSON, _ := json.MarshalIndent(SystemPingToolResponse{Response: raw}, "", "  ")
	return mcp.NewToolResultText(string(respJSON)), nil
}
