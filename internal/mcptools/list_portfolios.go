package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

// Raw API response types for /api/views/list (Server)
type listViewsResponse struct {
	Views  []listViewsEntry `json:"views"`
	Paging listViewsPaging  `json:"paging"`
}
type listViewsEntry struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Desc       string `json:"desc"`
	Qualifier  string `json:"qualifier,omitempty"`
	Visibility string `json:"visibility,omitempty"`
	IsFavorite *bool  `json:"isFavorite,omitempty"`
}
type listViewsPaging struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	Total     int `json:"total"`
}

// Raw API response types for /api/v2/views/list (Cloud)
type cloudViewsResponse struct {
	Views  []cloudViewEntry `json:"views"`
	Paging listViewsPaging  `json:"paging"`
}
type cloudViewEntry struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  *string  `json:"description,omitempty"`
	EnterpriseID *string  `json:"enterpriseId,omitempty"`
	Selection    *string  `json:"selection,omitempty"`
	IsDraft      *bool    `json:"isDraft,omitempty"`
	DraftStage   *int     `json:"draftStage,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

// Tool response types matching Java ListPortfoliosToolResponse
type CloudPortfolio struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  *string  `json:"description,omitempty"`
	EnterpriseID *string  `json:"enterpriseId,omitempty"`
	Selection    *string  `json:"selection,omitempty"`
	IsDraft      *bool    `json:"isDraft,omitempty"`
	DraftStage   *int     `json:"draftStage,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

type ServerPortfolio struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Qualifier  string `json:"qualifier"`
	Visibility string `json:"visibility"`
	IsFavorite *bool  `json:"isFavorite,omitempty"`
}

type ListPortfoliosToolResponse struct {
	Portfolios []interface{} `json:"portfolios"`
	Paging     *Paging       `json:"paging,omitempty"`
}

func NewListPortfoliosMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"list_portfolios",
		"List Portfolios — List portfolios/views. On SonarQube Cloud, provides enterprise portfolios. On Server, provides standard views.",
		json.RawMessage(
			`{
				"type": "object",
				"properties": {
					"q": {"type": "string", "description": "Optional search query to filter by name."},
					"favorite": {"type": "boolean", "description": "Only show favorite portfolios."},
					"pageIndex": {"type": "integer", "description": "1-based page index.", "default": 1},
					"pageSize": {"type": "integer", "description": "Page size (Cloud default: 50, Server default: 100).", "default": 100},
					"enterpriseId": {"type": "string", "description": "Enterprise ID (Cloud only)."},
					"draft": {"type": "boolean", "description": "Only show draft portfolios (Cloud only)."}
				},
				"additionalProperties": false
	}`))
}

func ListPortfoliosHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	params := url.Values{}
	if q := mcputils.GetOptionalString(args, "q"); q != "" {
		params.Set("q", q)
	}
	if mcputils.GetBoolOrDefault(args, "favorite", false) {
		params.Set("favorite", "true")
	}

	// Cloud vs Server defaults
	defaultPageSize := 100
	if mcputils.IsCloud() {
		defaultPageSize = 50
	}
	pageIndex := mcputils.GetIntOrDefault(args, "pageIndex", 1)
	pageSize := mcputils.GetIntOrDefault(args, "pageSize", defaultPageSize)
	params.Set("p", strconv.Itoa(pageIndex))
	params.Set("ps", strconv.Itoa(pageSize))

	client := mcputils.NewSQClient()

	if mcputils.IsCloud() {
		// Cloud uses /api/v2/views/list with enterprise-specific params
		if enterpriseID := mcputils.GetOptionalString(args, "enterpriseId"); enterpriseID != "" {
			params.Set("enterpriseId", enterpriseID)
		}
		if mcputils.GetBoolOrDefault(args, "draft", false) {
			params.Set("draft", "true")
		}
		if org := mcputils.GetSonarQubeOrg(); org != "" {
			params.Set("organization", org)
		}

		var rawResp cloudViewsResponse
		if err := client.DoGet(ctx, "/api/v2/views/list", params, &rawResp); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("List portfolios failed: %v", err)), nil
		}

		portfolios := make([]interface{}, len(rawResp.Views))
		for i, v := range rawResp.Views {
			portfolios[i] = CloudPortfolio{
				ID:           v.ID,
				Name:         v.Name,
				Description:  v.Description,
				EnterpriseID: v.EnterpriseID,
				Selection:    v.Selection,
				IsDraft:      v.IsDraft,
				DraftStage:   v.DraftStage,
				Tags:         v.Tags,
			}
		}

		resp := ListPortfoliosToolResponse{
			Portfolios: portfolios,
			Paging: &Paging{
				PageIndex:   rawResp.Paging.PageIndex,
				PageSize:    rawResp.Paging.PageSize,
				Total:       rawResp.Paging.Total,
				HasNextPage: rawResp.Paging.PageIndex*rawResp.Paging.PageSize < rawResp.Paging.Total,
			},
		}

		jsonBytes, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
		}
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}

	// Server uses /api/views/list
	var rawResp listViewsResponse
	if err := client.DoGet(ctx, "/api/views/list", params, &rawResp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("List portfolios failed: %v", err)), nil
	}

	portfolios := make([]interface{}, len(rawResp.Views))
	for i, v := range rawResp.Views {
		portfolios[i] = ServerPortfolio{
			Key:        v.Key,
			Name:       v.Name,
			Qualifier:  v.Qualifier,
			Visibility: v.Visibility,
			IsFavorite: v.IsFavorite,
		}
	}

	resp := ListPortfoliosToolResponse{
		Portfolios: portfolios,
		Paging: &Paging{
			PageIndex:   rawResp.Paging.PageIndex,
			PageSize:    rawResp.Paging.PageSize,
			Total:       rawResp.Paging.Total,
			HasNextPage: rawResp.Paging.PageIndex*rawResp.Paging.PageSize < rawResp.Paging.Total,
		},
	}

	jsonBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
	}
	return mcp.NewToolResultText(string(jsonBytes)), nil
}
