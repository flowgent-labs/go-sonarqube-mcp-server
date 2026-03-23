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

// Raw API response structures for /api/metrics/search
type metricsSearchResponse struct {
	Metrics []metricsSearchEntry `json:"metrics"`
	Paging  metricsSearchPaging  `json:"paging"`
}
type metricsSearchEntry struct {
	ID          int    `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Domain      string `json:"domain"`
	Type        string `json:"type"`
	Hidden      bool   `json:"hidden"`
	Custom      bool   `json:"custom"`
}
type metricsSearchPaging struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	Total     int `json:"total"`
}

// Structured response matching Java SearchMetricsToolResponse
type SearchMetricsToolResponse struct {
	Metrics  []MetricsItem `json:"metrics"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}
type MetricsItem struct {
	ID          int     `json:"id"`
	Key         string  `json:"key"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Domain      *string `json:"domain,omitempty"`
	Type        string  `json:"type"`
	Hidden      bool    `json:"hidden"`
	Custom      bool    `json:"custom"`
}

func NewSearchMetricsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"search_metrics",
		"Search Metrics — Search for available metrics on the SonarQube instance.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"p": {"type": "integer", "description": "Page number. Defaults to 1.", "default": 1},
				"ps": {"type": "integer", "description": "Page size. Max 500. Defaults to 100.", "default": 100}
			},
			"additionalProperties": false
	}`))
}

func SearchMetricsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	page := mcputils.GetIntOrDefault(args, "p", 1)
	pageSize := mcputils.GetIntOrDefault(args, "ps", 100)

	params := url.Values{}
	params.Set("p", strconv.Itoa(page))
	params.Set("ps", strconv.Itoa(pageSize))

	client := mcputils.NewSQClient()
	var resp metricsSearchResponse
	if err := client.DoGet(ctx, "/api/metrics/search", params, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search metrics failed: %v", err)), nil
	}

	metrics := make([]MetricsItem, 0, len(resp.Metrics))
	for _, m := range resp.Metrics {
		item := MetricsItem{
			ID:     m.ID,
			Key:    m.Key,
			Name:   m.Name,
			Type:   m.Type,
			Hidden: m.Hidden,
			Custom: m.Custom,
		}
		if m.Description != "" {
			item.Description = &m.Description
		}
		if m.Domain != "" {
			item.Domain = &m.Domain
		}
		metrics = append(metrics, item)
	}

	response := SearchMetricsToolResponse{
		Metrics:  metrics,
		Total:    resp.Paging.Total,
		Page:     resp.Paging.PageIndex,
		PageSize: resp.Paging.PageSize,
	}

	respJSON, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(respJSON)), nil
}
