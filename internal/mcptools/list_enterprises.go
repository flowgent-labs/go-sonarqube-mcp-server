package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

type enterpriseEntry struct {
	ID    string `json:"id"`
	Key   string `json:"key"`
	Name  string `json:"name"`
	Avatar string `json:"avatar"`
	DefaultPortfolioPermissionTemplateID string `json:"defaultPortfolioPermissionTemplateId"`
}

func NewListEnterprisesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"list_enterprises",
		"List SonarQube Cloud Enterprises — List the enterprises available in SonarQube Cloud that you have access to. Use this tool to discover enterprise IDs for other tools.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"enterpriseKey": {"type": "string", "description": "Optional enterprise key to filter results."}
			},
			"additionalProperties": false
	}`))
}

func ListEnterprisesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !mcputils.IsCloud() {
		return mcp.NewToolResultError("list_enterprises is only available on SonarQube Cloud."), nil
	}

	args := request.GetArguments()
	params := url.Values{}
	if ek := mcputils.GetOptionalString(args, "enterpriseKey"); ek != "" {
		params.Set("enterpriseKey", ek)
	}

	client := mcputils.NewSQClient()
	// The enterprises API returns a JSON array directly
	raw, err := client.DoGetRaw(ctx, "/enterprises/enterprises", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("List enterprises failed: %v", err)), nil
	}

	var enterprises []enterpriseEntry
	if err := json.Unmarshal([]byte(raw), &enterprises); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse enterprises: %v", err)), nil
	}

	if len(enterprises) == 0 {
		return mcp.NewToolResultText("No enterprises found."), nil
	}

	text := "Enterprises:\n"
	for _, e := range enterprises {
		text += fmt.Sprintf("- %s (%s) [id=%s]\n", e.Name, e.Key, e.ID)
	}
	return mcp.NewToolResultText(text), nil
}
