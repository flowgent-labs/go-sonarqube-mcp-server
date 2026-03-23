package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

// Raw API response structures
type measuresComponentResponse struct {
	Component measuresComponent `json:"component"`
}
type measuresComponent struct {
	Key         string                  `json:"key"`
	Name        string                  `json:"name"`
	Qualifier   string                  `json:"qualifier"`
	Description string                  `json:"description"`
	Language    string                  `json:"language"`
	Path        string                  `json:"path"`
	Measures    []measuresMeasureEntry  `json:"measures"`
}
type measuresMeasureEntry struct {
	Metric    string `json:"metric"`
	Value     string `json:"value"`
	BestValue bool   `json:"bestValue"`
}

// Structured response matching Java GetComponentMeasuresToolResponse
type GetComponentMeasuresToolResponse struct {
	Component Component       `json:"component"`
	Measures  []Measure       `json:"measures"`
	Metrics   []Metric        `json:"metrics,omitempty"`
}
type Component struct {
	Key         string  `json:"key"`
	Name        string  `json:"name"`
	Qualifier   string  `json:"qualifier"`
	Description *string `json:"description,omitempty"`
	Language    *string `json:"language,omitempty"`
	Path        *string `json:"path,omitempty"`
}
type Measure struct {
	Metric string  `json:"metric"`
	Value  *string `json:"value,omitempty"`
}
type Metric struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Domain      string `json:"domain,omitempty"`
	Type        string `json:"type"`
	Hidden      bool   `json:"hidden"`
	Custom      bool   `json:"custom"`
}

func NewGetComponentMeasuresMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"get_component_measures",
		"Get Component Measures — Get project measures like ncloc, complexity, violations, coverage, etc.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"projectKey": {"type": "string", "description": "SonarQube project key. Required unless a default is configured via SONARQUBE_PROJECT_KEY."},
				"branch": {"type": "string", "description": "Long-lived branch name."},
				"pullRequest": {"type": "string", "description": "Pull request key."},
				"metricKeys": {"type": "array", "items": {"type": "string"}, "description": "Metric keys (e.g. ncloc, complexity, violations, coverage)."}
			},
			"additionalProperties": false
	}`))
}

func GetComponentMeasuresHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	projectKey, err := mcputils.ResolveProjectKey(args, "projectKey")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	metricKeys := mcputils.GetStringArray(args, "metricKeys")

	params := url.Values{}
	params.Set("component", projectKey)
	if len(metricKeys) > 0 {
		params.Set("metricKeys", strings.Join(metricKeys, ","))
	}
	if branch := mcputils.GetOptionalString(args, "branch"); branch != "" {
		params.Set("branch", branch)
	}
	if pr := mcputils.GetOptionalString(args, "pullRequest"); pr != "" {
		params.Set("pullRequest", pr)
	}

	client := mcputils.NewSQClient()
	var resp measuresComponentResponse
	if err := client.DoGet(ctx, "/api/measures/component", params, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Get component measures failed: %v", err)), nil
	}

	// Build component
	c := resp.Component
	component := Component{
		Key:       c.Key,
		Name:      c.Name,
		Qualifier: c.Qualifier,
	}
	if c.Description != "" {
		component.Description = &c.Description
	}
	if c.Language != "" {
		component.Language = &c.Language
	}
	if c.Path != "" {
		component.Path = &c.Path
	}

	// Build measures
	measures := make([]Measure, 0, len(c.Measures))
	for _, m := range c.Measures {
		measure := Measure{Metric: m.Metric}
		if m.Value != "" {
			measure.Value = &m.Value
		}
		measures = append(measures, measure)
	}

	response := GetComponentMeasuresToolResponse{
		Component: component,
		Measures:  measures,
	}

	respJSON, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(respJSON)), nil
}
