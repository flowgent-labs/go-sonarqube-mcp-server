package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

// Raw API response structures
type projectBranchesResponse struct {
	Branches []projectBranchEntry `json:"branches"`
}
type projectBranchEntry struct {
	Name         string              `json:"name"`
	IsMain       bool                `json:"isMain"`
	Type         string              `json:"type"`
	Status       *projectBranchStatus `json:"status,omitempty"`
	AnalysisDate string              `json:"analysisDate"`
	BranchID     string              `json:"branchId"`
}
type projectBranchStatus struct {
	QualityGateStatus string `json:"qualityGateStatus"`
}

// Structured response matching Java ListBranchesToolResponse
type ListBranchesToolResponse struct {
	ProjectKey    string                  `json:"projectKey"`
	TotalBranches int                     `json:"totalBranches"`
	Branches      []ListBranchesBranch    `json:"branches"`
}
type ListBranchesBranch struct {
	Name             string `json:"name"`
	IsMain           bool   `json:"isMain"`
	Type             string `json:"type"`
	QualityGateStatus string `json:"qualityGateStatus,omitempty"`
	AnalysisDate     string `json:"analysisDate,omitempty"`
	BranchID         string `json:"branchId"`
}

// isLongLivedBranchType filters branches to only long-lived (BRANCH) type.
func isLongLivedBranchType(t string) bool {
	return t == "BRANCH" || t == "LONG"
}

func parseQualityGateStatus(s string) string {
	switch s {
	case "OK":
		return "OK"
	case "WARN":
		return "WARN"
	case "ERROR":
		return "ERROR"
	default:
		return s
	}
}

func NewListBranchesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"list_branches",
		"List SonarQube Branches — List long-lived branches for a project (e.g. main, develop, master). Use returned branch names as the branch parameter on other tools (e.g. get_project_quality_gate_status, get_component_measures). For pull requests, use list_pull_requests instead.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"projectKey": {"type": "string", "description": "SonarQube project key. Required unless a default is configured via SONARQUBE_PROJECT_KEY."}
			},
			"additionalProperties": false
	}`))
}

func ListBranchesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	projectKey, err := mcputils.ResolveProjectKey(args, "projectKey")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := url.Values{}
	params.Set("project", projectKey)

	client := mcputils.NewSQClient()
	var resp projectBranchesResponse
	if err := client.DoGet(ctx, "/api/project_branches/list", params, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("List branches failed: %v", err)), nil
	}

	// Filter long-lived branches and build structured response
	branches := make([]ListBranchesBranch, 0)
	for _, b := range resp.Branches {
		if !isLongLivedBranchType(b.Type) {
			continue
		}
		qgStatus := ""
		if b.Status != nil {
			qgStatus = parseQualityGateStatus(b.Status.QualityGateStatus)
		}
		branches = append(branches, ListBranchesBranch{
			Name:             b.Name,
			IsMain:           b.IsMain,
			Type:             b.Type,
			QualityGateStatus: qgStatus,
			AnalysisDate:     b.AnalysisDate,
			BranchID:         b.BranchID,
		})
	}

	response := ListBranchesToolResponse{
		ProjectKey:    projectKey,
		TotalBranches: len(branches),
		Branches:      branches,
	}

	respJSON, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(respJSON)), nil
}
