package mcptools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

type systemHealthRaw struct {
	Health string                `json:"health"`
	Causes []systemHealthCause   `json:"causes,omitempty"`
	Nodes  []systemHealthNode    `json:"nodes,omitempty"`
}
type systemHealthCause struct {
	Message string `json:"message"`
}
type systemHealthNode struct {
	Name       string              `json:"name"`
	Type       string              `json:"type"`
	Health     string              `json:"health"`
	Host       string              `json:"host"`
	Port       int                 `json:"port"`
	StartedAt  string              `json:"startedAt"`
	Causes     []systemHealthCause `json:"causes,omitempty"`
}

// Structured response matching Java SystemHealthToolResponse
type SystemHealthToolResponse struct {
	Health string                  `json:"health"`
	Causes []SystemHealthCause     `json:"causes,omitempty"`
	Nodes  []SystemHealthNode      `json:"nodes,omitempty"`
}
type SystemHealthCause struct {
	Message string `json:"message"`
}
type SystemHealthNode struct {
	Name      string              `json:"name"`
	Type      string              `json:"type"`
	Health    string              `json:"health"`
	Host      string              `json:"host"`
	Port      int                 `json:"port"`
	StartedAt string              `json:"startedAt"`
	Causes    []SystemHealthCause `json:"causes,omitempty"`
}

func NewSystemHealthMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"get_system_health",
		"Get System Health — Get health status (GREEN, YELLOW, RED) with causes and node details.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {},
			"additionalProperties": false
	}`))
}

func SystemHealthHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := mcputils.NewSQClient()
	var raw systemHealthRaw
	if err := client.DoGet(ctx, "/api/system/health", nil, &raw); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Get system health failed: %v", err)), nil
	}

	causes := make([]SystemHealthCause, 0, len(raw.Causes))
	for _, c := range raw.Causes {
		causes = append(causes, SystemHealthCause{Message: c.Message})
	}

	nodes := make([]SystemHealthNode, 0, len(raw.Nodes))
	for _, n := range raw.Nodes {
		nodeCauses := make([]SystemHealthCause, 0, len(n.Causes))
		for _, c := range n.Causes {
			nodeCauses = append(nodeCauses, SystemHealthCause{Message: c.Message})
		}
		nodes = append(nodes, SystemHealthNode{
			Name:      n.Name,
			Type:      n.Type,
			Health:    n.Health,
			Host:      n.Host,
			Port:      n.Port,
			StartedAt: n.StartedAt,
			Causes:    nodeCauses,
		})
	}

	response := SystemHealthToolResponse{
		Health: raw.Health,
		Causes: causes,
		Nodes:  nodes,
	}
	respJSON, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(respJSON)), nil
}
