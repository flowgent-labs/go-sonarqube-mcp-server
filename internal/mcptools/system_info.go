package mcptools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

type SystemInfoToolResponse struct {
	Sections []SystemInfoSection `json:"sections"`
}
type SystemInfoSection struct {
	Name       string                 `json:"name"`
	Attributes map[string]interface{} `json:"attributes"`
}

func NewSystemInfoMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"get_system_info",
		"Get System Info — Get detailed system configuration (JVM, DB, search indexes, settings). Requires Administer permission.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {},
			"additionalProperties": false
	}`))
}

func SystemInfoHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := mcputils.NewSQClient()
	var raw map[string]interface{}
	if err := client.DoGet(ctx, "/api/system/info", nil, &raw); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Get system info failed: %v", err)), nil
	}

	sections := make([]SystemInfoSection, 0)
	for _, name := range []string{"System", "Database", "Bundled Plugins", "Installed Plugins",
		"Web JVM State", "Web Database Connection", "Web Logging",
		"Compute Engine Tasks", "Compute Engine JVM State", "Compute Engine Database Connection", "Compute Engine Logging",
		"Search State", "Search Indexes", "ALMs", "Server Push Connections", "Settings"} {
		if v, ok := raw[name]; ok {
			if m, ok := v.(map[string]interface{}); ok && len(m) > 0 {
				sections = append(sections, SystemInfoSection{Name: name, Attributes: m})
			}
		}
	}

	response := SystemInfoToolResponse{Sections: sections}
	respJSON, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(respJSON)), nil
}
