package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

// API response types
type createWebhookResponse struct {
	Webhook createWebhookEntry `json:"webhook"`
}
type createWebhookEntry struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Tool response types (matching Java CreateWebhookToolResponse)
type CreateWebhookToolResponse struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	HasSecret bool   `json:"hasSecret"`
}

func NewCreateWebhookMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"create_webhook",
		"Create Webhook — Create a new webhook. Requires Administer permission.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"name": {"type": "string", "description": "Webhook name (max 100 chars)."},
				"url": {"type": "string", "description": "Webhook URL (max 512 chars)."},
				"projectKey": {"type": "string", "description": "Optional project key (max 400 chars)."},
				"secret": {"type": "string", "description": "Optional secret for HMAC validation (16-200 chars)."}
			},
			"required": ["name", "url"],
			"additionalProperties": false
}`))
}

func CreateWebhookHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if mcputils.IsReadOnly() {
		return mcp.NewToolResultError("Cannot create webhook: SONARQUBE_READ_ONLY is enabled"), nil
	}

	args := request.GetArguments()
	name := mcputils.GetOptionalString(args, "name")
	webhookURL := mcputils.GetOptionalString(args, "url")

	if name == "" || webhookURL == "" {
		return mcp.NewToolResultError("Both name and url are required"), nil
	}

	params := url.Values{}
	params.Set("name", name)
	params.Set("url", webhookURL)

	projectKey := mcputils.OptionalProjectKey(args, "projectKey")
	secret := mcputils.GetOptionalString(args, "secret")
	if secret != "" {
		params.Set("secret", secret)
	}

	client := mcputils.NewSQClient()

	var resp createWebhookResponse
	path := "/api/webhooks/create"
	if projectKey != "" {
		path = "/api/project_webhooks/create"
		params.Set("project", projectKey)
	}

	if err := client.DoPost(ctx, path, params, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Create webhook failed: %v", err)), nil
	}

	response := CreateWebhookToolResponse{
		Key:       resp.Webhook.Key,
		Name:      resp.Webhook.Name,
		URL:       resp.Webhook.URL,
		HasSecret: secret != "",
	}

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
	}
	return mcp.NewToolResultText(string(jsonBytes)), nil
}
