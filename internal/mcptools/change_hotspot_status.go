package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

// Tool response types (matching Java ChangeSecurityHotspotStatusToolResponse)
type ChangeSecurityHotspotStatusToolResponse struct {
	Success       bool    `json:"success"`
	Message       string  `json:"message"`
	HotspotKey    string  `json:"hotspotKey"`
	NewStatus     string  `json:"newStatus"`
	NewResolution *string `json:"newResolution,omitempty"`
}

func NewChangeSecurityHotspotStatusMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"change_security_hotspot_status",
		"Change Security Hotspot Status — Change security hotspot status (TO_REVIEW or REVIEWED with resolution FIXED/SAFE/ACKNOWLEDGED).",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"hotspotKey": {"type": "string", "description": "Security hotspot key."},
				"status": {"type": "string", "enum": ["TO_REVIEW", "REVIEWED"], "description": "New status for the hotspot."},
				"resolution": {"type": "string", "enum": ["FIXED", "SAFE", "ACKNOWLEDGED"], "description": "Resolution (required if status is REVIEWED)."},
				"comment": {"type": "string", "description": "Optional comment."}
			},
			"required": ["hotspotKey", "status"],
			"additionalProperties": false
}`))
}

func ChangeSecurityHotspotStatusHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if mcputils.IsReadOnly() {
		return mcp.NewToolResultError("Cannot change hotspot status: SONARQUBE_READ_ONLY is enabled"), nil
	}

	args := request.GetArguments()
	hotspotKey := mcputils.GetOptionalString(args, "hotspotKey")
	status := mcputils.GetOptionalString(args, "status")
	resolution := mcputils.GetOptionalString(args, "resolution")
	comment := mcputils.GetOptionalString(args, "comment")

	if hotspotKey == "" || status == "" {
		return mcp.NewToolResultError("Both hotspotKey and status are required"), nil
	}
	if status == "REVIEWED" && resolution == "" {
		return mcp.NewToolResultError("resolution is required when status is REVIEWED"), nil
	}
	if status == "TO_REVIEW" && resolution != "" {
		return mcp.NewToolResultError("resolution must not be set when status is TO_REVIEW"), nil
	}

	params := url.Values{}
	params.Set("hotspot", hotspotKey)
	params.Set("status", status)
	if resolution != "" {
		params.Set("resolution", resolution)
	}
	if comment != "" {
		params.Set("comment", comment)
	}

	client := mcputils.NewSQClient()
	if err := client.DoPost(ctx, "/api/hotspots/change_status", params, nil); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Change hotspot status failed: %v", err)), nil
	}

	response := ChangeSecurityHotspotStatusToolResponse{
		Success:    true,
		Message:    fmt.Sprintf("Hotspot %s status changed to %s.", hotspotKey, status),
		HotspotKey: hotspotKey,
		NewStatus:  status,
	}
	if resolution != "" {
		r := resolution
		response.NewResolution = &r
	}

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
	}
	return mcp.NewToolResultText(string(jsonBytes)), nil
}
