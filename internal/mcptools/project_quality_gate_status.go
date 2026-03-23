package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

// Raw API response types for /api/qualitygates/project_status
type projectStatusResponse struct {
	ProjectStatus projectStatusEntry `json:"projectStatus"`
}
type projectStatusEntry struct {
	Status            string                   `json:"status"`
	Conditions        []projectStatusCondition `json:"conditions"`
	IgnoredConditions *bool                    `json:"ignoredConditions,omitempty"`
}
type projectStatusCondition struct {
	MetricKey      string `json:"metricKey"`
	Comparator     string `json:"comparator"`
	ErrorThreshold string `json:"errorThreshold"`
	Status         string `json:"status"`
	ActualValue    string `json:"actualValue"`
}

// Tool response types matching Java ProjectStatusToolResponse
type ProjectStatusToolResponse struct {
	Status            string      `json:"status"`
	Conditions        []Condition `json:"conditions"`
	IgnoredConditions *bool       `json:"ignoredConditions,omitempty"`
}

type Condition struct {
	MetricKey      string  `json:"metricKey"`
	Status         string  `json:"status"`
	ErrorThreshold *string `json:"errorThreshold,omitempty"`
	ActualValue    *string `json:"actualValue,omitempty"`
}

func NewProjectQualityGateStatusMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"get_project_quality_gate_status",
		"Get Project Quality Gate Status — Get the quality gate status for a project. Requires analysisId, projectId, or projectKey.",
		json.RawMessage(
			`{
				"type": "object",
				"properties": {
					"analysisId": {"type": "string", "description": "Analysis ID."},
					"projectId": {"type": "string", "description": "Project ID."},
					"projectKey": {"type": "string", "description": "SonarQube project key."},
					"branch": {"type": "string", "description": "Branch name."},
					"pullRequest": {"type": "string", "description": "Pull request key."}
				},
				"additionalProperties": false
	}`))
}

func ProjectQualityGateStatusHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	analysisID := mcputils.GetOptionalString(args, "analysisId")
	projectID := mcputils.GetOptionalString(args, "projectId")
	projectKey := mcputils.OptionalProjectKey(args, "projectKey")

	// Validation: at least one of analysisId, projectId, or projectKey must be provided
	if analysisID == "" && projectID == "" && projectKey == "" {
		return mcp.NewToolResultError("At least one of analysisId, projectId, or projectKey must be provided"), nil
	}

	// Validation: projectId cannot be used with branch or pullRequest
	branch := mcputils.GetOptionalString(args, "branch")
	pullRequest := mcputils.GetOptionalString(args, "pullRequest")
	if projectID != "" && (branch != "" || pullRequest != "") {
		return mcp.NewToolResultError("projectId cannot be used together with branch or pullRequest"), nil
	}

	params := url.Values{}
	if analysisID != "" {
		params.Set("analysisId", analysisID)
	}
	if projectID != "" {
		params.Set("projectId", projectID)
	}
	if projectKey != "" {
		params.Set("projectKey", projectKey)
	}
	if branch != "" {
		params.Set("branch", branch)
	}
	if pullRequest != "" {
		params.Set("pullRequest", pullRequest)
	}

	client := mcputils.NewSQClient()
	var rawResp projectStatusResponse
	if err := client.DoGet(ctx, "/api/qualitygates/project_status", params, &rawResp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Get project quality gate status failed: %v", err)), nil
	}

	ps := rawResp.ProjectStatus
	conditions := make([]Condition, len(ps.Conditions))
	for i, c := range ps.Conditions {
		cond := Condition{
			MetricKey: c.MetricKey,
			Status:    c.Status,
		}
		if c.ErrorThreshold != "" {
			cond.ErrorThreshold = &c.ErrorThreshold
		}
		if c.ActualValue != "" {
			cond.ActualValue = &c.ActualValue
		}
		conditions[i] = cond
	}

	resp := ProjectStatusToolResponse{
		Status:            ps.Status,
		Conditions:        conditions,
		IgnoredConditions: ps.IgnoredConditions,
	}

	jsonBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
	}
	return mcp.NewToolResultText(string(jsonBytes)), nil
}
