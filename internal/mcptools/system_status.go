package mcptools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

type systemStatusRaw struct {
	Status  string `json:"status"`
	ID      string `json:"id"`
	Version string `json:"version"`
}

type SystemStatusToolResponse struct {
	Status      string `json:"status"`
	Description string `json:"description"`
	ID          string `json:"id"`
	Version     string `json:"version"`
}

func statusDescription(s string) string {
	switch s {
	case "STARTING":
		return "SonarQube Server Web Server is up and serving some Web Services but initialization is still ongoing"
	case "UP":
		return "SonarQube Server instance is up and running"
	case "DOWN":
		return "SonarQube Server instance is up but not running because migration has failed or some other reason"
	case "RESTARTING":
		return "SonarQube Server instance is still up but a restart has been requested"
	case "DB_MIGRATION_NEEDED":
		return "Database migration is required"
	case "DB_MIGRATION_RUNNING":
		return "DB migration is running"
	default:
		return "Unknown status"
	}
}

func NewSystemStatusMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"get_system_status",
		"Get System Status — Get SonarQube Server status (STARTING, UP, DOWN, RESTARTING, DB_MIGRATION_NEEDED, DB_MIGRATION_RUNNING), version, and ID.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {},
			"additionalProperties": false
	}`))
}

func SystemStatusHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := mcputils.NewSQClient()
	var raw systemStatusRaw
	if err := client.DoGet(ctx, "/api/system/status", nil, &raw); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Get system status failed: %v", err)), nil
	}

	response := SystemStatusToolResponse{
		Status:      raw.Status,
		Description: statusDescription(raw.Status),
		ID:          raw.ID,
		Version:     raw.Version,
	}
	respJSON, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(respJSON)), nil
}
