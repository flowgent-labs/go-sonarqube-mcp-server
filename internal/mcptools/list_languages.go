package mcptools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

// Raw API response structures for /api/languages/list
type languagesListResponse struct {
	Languages []languagesListEntry `json:"languages"`
}
type languagesListEntry struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// Structured response matching Java ListLanguagesToolResponse
type ListLanguagesToolResponse struct {
	Languages []LanguageItem `json:"languages"`
}
type LanguageItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

func NewListLanguagesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"list_languages",
		"List Languages — List all programming languages supported by this SonarQube instance.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"q": {"type": "string", "description": "Optional pattern to match language keys or names."}
			},
			"additionalProperties": false
	}`))
}

func ListLanguagesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	q := mcputils.GetOptionalString(args, "q")

	client := mcputils.NewSQClient()
	var resp languagesListResponse
	if err := client.DoGet(ctx, "/api/languages/list", nil, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("List languages failed: %v", err)), nil
	}

	languages := make([]LanguageItem, 0, len(resp.Languages))
	for _, l := range resp.Languages {
		if q != "" && l.Key != q && l.Name != q {
			continue
		}
		languages = append(languages, LanguageItem{
			Key:  l.Key,
			Name: l.Name,
		})
	}

	response := ListLanguagesToolResponse{
		Languages: languages,
	}

	respJSON, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(respJSON)), nil
}
