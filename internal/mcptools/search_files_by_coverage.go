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

// Raw API response structures for /api/measures/component (project summary)
type coverageProjectSummaryResponse struct {
	Component coverageProjectComp `json:"component"`
}
type coverageProjectComp struct {
	Measures []coverageSummaryMeasure `json:"measures"`
}
type coverageSummaryMeasure struct {
	Metric string `json:"metric"`
	Value  string `json:"value"`
}

// Raw API response structures for /api/measures/component_tree (file list)
type coverageComponentTreeResponse struct {
	BaseComponent coverageBaseComp   `json:"baseComponent"`
	Components    []coverageFileComp `json:"components"`
	Paging        coverageTreePaging `json:"paging"`
}
type coverageBaseComp struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}
type coverageFileComp struct {
	Key      string                   `json:"key"`
	Path     string                   `json:"path"`
	Name     string                   `json:"name"`
	Qualifier string                  `json:"qualifier"`
	Measures []coverageFileMeasure    `json:"measures"`
}
type coverageFileMeasure struct {
	Metric string `json:"metric"`
	Value  string `json:"value"`
}
type coverageTreePaging struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	Total     int `json:"total"`
}

// Structured response matching Java SearchFilesByCoverageToolResponse
type SearchFilesByCoverageToolResponse struct {
	ProjectKey     string              `json:"projectKey"`
	TotalFiles     int                 `json:"totalFiles"`
	FilesReturned  int                 `json:"filesReturned"`
	PageIndex      int                 `json:"pageIndex"`
	PageSize       int                 `json:"pageSize"`
	ProjectSummary *ProjectSummary     `json:"projectSummary,omitempty"`
	Files          []FileWithCoverage  `json:"files"`
}
type ProjectSummary struct {
	Coverage       *float64 `json:"coverage,omitempty"`
	LinesToCover   *int     `json:"linesToCover,omitempty"`
	UncoveredLines *int     `json:"uncoveredLines,omitempty"`
}
type FileWithCoverage struct {
	Key                   string   `json:"key"`
	Path                  string   `json:"path"`
	Coverage              *float64 `json:"coverage,omitempty"`
	LineCoverage          *float64 `json:"lineCoverage,omitempty"`
	BranchCoverage        *float64 `json:"branchCoverage,omitempty"`
	LinesToCover          *int     `json:"linesToCover,omitempty"`
	UncoveredLines        *int     `json:"uncoveredLines,omitempty"`
	ConditionsToCover     *int     `json:"conditionsToCover,omitempty"`
	UncoveredConditions   *int     `json:"uncoveredConditions,omitempty"`
}

func NewSearchFilesByCoverageMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"search_files_by_coverage",
		"Search Files by Coverage — Search files sorted by coverage ascending (worst first). Helps identify files needing test improvements.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"projectKey": {"type": "string", "description": "SonarQube project key. Required unless a default is configured via SONARQUBE_PROJECT_KEY."},
				"branch": {"type": "string", "description": "Branch name."},
				"pullRequest": {"type": "string", "description": "Pull request key."},
				"maxCoverage": {"type": "number", "description": "Maximum coverage percentage (0-100). Only files below this threshold."},
				"pageIndex": {"type": "integer", "description": "1-based page index. Defaults to 1.", "default": 1},
				"pageSize": {"type": "integer", "description": "Page size. Max 500.", "default": 100}
			},
			"additionalProperties": false
	}`))
}

func parseNullableFloat(s string) *float64 {
	if s == "" {
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &v
}

func parseNullableInt(s string) *int {
	if s == "" {
		return nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &v
}

func SearchFilesByCoverageHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	projectKey, err := mcputils.ResolveProjectKey(args, "projectKey")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client := mcputils.NewSQClient()

	// ----- FIRST API CALL: Project summary from /api/measures/component -----
	var projectSummary *ProjectSummary
	summaryParams := url.Values{}
	summaryParams.Set("component", projectKey)
	summaryParams.Set("metricKeys", "coverage,lines_to_cover,uncovered_lines")
	if branch := mcputils.GetOptionalString(args, "branch"); branch != "" {
		summaryParams.Set("branch", branch)
	}
	if pr := mcputils.GetOptionalString(args, "pullRequest"); pr != "" {
		summaryParams.Set("pullRequest", pr)
	}

	var summaryResp coverageProjectSummaryResponse
	if err := client.DoGet(ctx, "/api/measures/component", summaryParams, &summaryResp); err == nil {
		ps := &ProjectSummary{}
		for _, m := range summaryResp.Component.Measures {
			switch m.Metric {
			case "coverage":
				ps.Coverage = parseNullableFloat(m.Value)
			case "lines_to_cover":
				ps.LinesToCover = parseNullableInt(m.Value)
			case "uncovered_lines":
				ps.UncoveredLines = parseNullableInt(m.Value)
			}
		}
		projectSummary = ps
	}

	// ----- SECOND API CALL: File list from /api/measures/component_tree -----
	params := url.Values{}
	params.Set("component", projectKey)
	params.Set("metricKeys", "coverage,line_coverage,branch_coverage,lines_to_cover,uncovered_lines,conditions_to_cover,uncovered_conditions")
	params.Set("s", "metricPeriod")
	params.Set("asc", "true") // worst coverage first

	pageIndex := mcputils.GetIntOrDefault(args, "pageIndex", 1)
	pageSize := mcputils.GetIntOrDefault(args, "pageSize", 100)
	params.Set("p", strconv.Itoa(pageIndex))
	params.Set("ps", strconv.Itoa(pageSize))

	if branch := mcputils.GetOptionalString(args, "branch"); branch != "" {
		params.Set("branch", branch)
	}
	if pr := mcputils.GetOptionalString(args, "pullRequest"); pr != "" {
		params.Set("pullRequest", pr)
	}

	var treeResp coverageComponentTreeResponse
	if err := client.DoGet(ctx, "/api/measures/component_tree", params, &treeResp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search files by coverage failed: %v", err)), nil
	}

	maxCov := -1.0
	if v, ok := args["maxCoverage"]; ok {
		if f, ok := v.(float64); ok {
			maxCov = f
		}
	}

	files := make([]FileWithCoverage, 0, len(treeResp.Components))
	for _, comp := range treeResp.Components {
		fwc := FileWithCoverage{
			Key:  comp.Key,
			Path: comp.Path,
		}

		for _, m := range comp.Measures {
			switch m.Metric {
			case "coverage":
				fwc.Coverage = parseNullableFloat(m.Value)
			case "line_coverage":
				fwc.LineCoverage = parseNullableFloat(m.Value)
			case "branch_coverage":
				fwc.BranchCoverage = parseNullableFloat(m.Value)
			case "lines_to_cover":
				fwc.LinesToCover = parseNullableInt(m.Value)
			case "uncovered_lines":
				fwc.UncoveredLines = parseNullableInt(m.Value)
			case "conditions_to_cover":
				fwc.ConditionsToCover = parseNullableInt(m.Value)
			case "uncovered_conditions":
				fwc.UncoveredConditions = parseNullableInt(m.Value)
			}
		}

		// Apply maxCoverage filter
		if maxCov >= 0 && fwc.Coverage != nil && *fwc.Coverage > maxCov {
			continue
		}

		files = append(files, fwc)
	}

	response := SearchFilesByCoverageToolResponse{
		ProjectKey:    projectKey,
		TotalFiles:    treeResp.Paging.Total,
		FilesReturned: len(files),
		PageIndex:     treeResp.Paging.PageIndex,
		PageSize:      treeResp.Paging.PageSize,
		ProjectSummary: projectSummary,
		Files:         files,
	}

	respJSON, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(respJSON)), nil
}
