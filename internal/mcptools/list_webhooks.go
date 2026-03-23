package mcptools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

// API response types
type listWebhooksResponse struct {
	Webhooks []listWebhooksEntry `json:"webhooks"`
}
type listWebhooksEntry struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Secret string `json:"secret"`
}

// Tool response types (matching Java ListWebhooksToolResponse)
type ListWebhooksToolResponse struct {
	Webhooks []Webhook `json:"webhooks"`
}

type Webhook struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	HasSecret bool   `json:"hasSecret"`
}

func NewListWebhooksMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"list_webhooks",
		"List Webhooks — List webhooks for the organization/instance or project. Requires Administer permission.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"projectKey": {"type": "string", "description": "Optional project key to scope webhooks to a specific project."}
			},
			"additionalProperties": false
}`))
}

func ListWebhooksHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	projectKey := mcputils.OptionalProjectKey(args, "projectKey")

	client := mcputils.NewSQClient()

	var resp listWebhooksResponse
	path := "/api/webhooks/list"
	if projectKey != "" {
		path = "/api/project_webhooks/list?project=" + projectKey
	} else if mcputils.IsCloud() {
		if org := mcputils.GetSonarQubeOrg(); org != "" {
			path = "/api/webhooks/list?organization=" + org
		}
	}

	if err := client.DoGet(ctx, path, nil, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("List webhooks failed: %v", err)), nil
	}

	response := ListWebhooksToolResponse{
		Webhooks: make([]Webhook, 0, len(resp.Webhooks)),
	}
	for _, wh := range resp.Webhooks {
		response.Webhooks = append(response.Webhooks, Webhook{
			Key:       wh.Key,
			Name:      wh.Name,
			URL:       wh.URL,
			HasSecret: wh.Secret != "",
		})
	}

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
	}
	return mcp.NewToolResultText(string(jsonBytes)), nil
}
