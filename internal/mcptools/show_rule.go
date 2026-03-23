package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

// Raw API response structures for /api/rules/show
type rulesShowResponse struct {
	Rule rulesShowRule `json:"rule"`
}
type rulesShowRule struct {
	Key                 string                  `json:"key"`
	Name                string                  `json:"name"`
	Severity            string                  `json:"severity"`
	Type                string                  `json:"type"`
	Lang                string                  `json:"lang"`
	LangName            string                  `json:"langName"`
	HTMLDesc            string                  `json:"htmlDesc"`
	Impacts             []rulesShowImpact       `json:"impacts"`
	DescriptionSections []rulesShowDescSection  `json:"descriptionSections"`
}
type rulesShowImpact struct {
	SoftwareQuality string `json:"softwareQuality"`
	Severity        string `json:"severity"`
}
type rulesShowDescSection struct {
	Content string `json:"content"`
}

// Structured response matching Java ShowRuleToolResponse
type ShowRuleToolResponse struct {
	Key                 string               `json:"key"`
	Name                string               `json:"name"`
	Severity            string               `json:"severity"`
	Type                string               `json:"type"`
	Lang                string               `json:"lang"`
	LangName            string               `json:"langName"`
	HTMLDesc            *string              `json:"htmlDesc,omitempty"`
	Impacts             []Impact             `json:"impacts,omitempty"`
	DescriptionSections []DescriptionSection `json:"descriptionSections,omitempty"`
}
type Impact struct {
	SoftwareQuality string `json:"softwareQuality"`
	Severity        string `json:"severity"`
}
type DescriptionSection struct {
	Content string `json:"content"`
}

func NewShowRuleMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"show_rule",
		"Show Rule — Show detailed information about a SonarQube rule.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"key": {"type": "string", "description": "Rule key (e.g. javascript:EmptyBlock)."}
			},
			"required": ["key"],
			"additionalProperties": false
	}`))
}

func ShowRuleHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	ruleKey := mcputils.GetOptionalString(args, "key")
	if ruleKey == "" {
		return mcp.NewToolResultError("key is required"), nil
	}

	params := url.Values{}
	params.Set("key", ruleKey)

	client := mcputils.NewSQClient()
	var rawResp rulesShowResponse
	if err := client.DoGet(ctx, "/api/rules/show", params, &rawResp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Show rule failed: %v", err)), nil
	}

	r := rawResp.Rule

	response := ShowRuleToolResponse{
		Key:      r.Key,
		Name:     r.Name,
		Severity: r.Severity,
		Type:     r.Type,
		Lang:     r.Lang,
		LangName: r.LangName,
	}

	if r.HTMLDesc != "" {
		response.HTMLDesc = &r.HTMLDesc
	}

	if len(r.Impacts) > 0 {
		impacts := make([]Impact, 0, len(r.Impacts))
		for _, imp := range r.Impacts {
			impacts = append(impacts, Impact{
				SoftwareQuality: imp.SoftwareQuality,
				Severity:        imp.Severity,
			})
		}
		response.Impacts = impacts
	}

	if len(r.DescriptionSections) > 0 {
		sections := make([]DescriptionSection, 0, len(r.DescriptionSections))
		for _, s := range r.DescriptionSections {
			sections = append(sections, DescriptionSection{
				Content: s.Content,
			})
		}
		response.DescriptionSections = sections
	}

	respJSON, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(respJSON)), nil
}
