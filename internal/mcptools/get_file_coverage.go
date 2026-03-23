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
type sourceLinesResponse struct {
	Sources []sourceLine `json:"sources"`
}
type sourceLine struct {
	Line              int     `json:"line"`
	Code              string  `json:"code"`
	LineHits          *int    `json:"lineHits"`
	Conditions        *int    `json:"conditions"`
	CoveredConditions *int    `json:"coveredConditions"`
	ScmAuthor         *string `json:"scmAuthor"`
	ScmDate           *string `json:"scmDate"`
	ScmRevision       *string `json:"scmRevision"`
	IsNew             *bool   `json:"isNew"`
}

func (s sourceLine) isCoverable() bool { return s.LineHits != nil }
func (s sourceLine) isUncovered() bool { return s.LineHits != nil && *s.LineHits == 0 }

func (s sourceLine) hasPartialBranchCoverage() bool {
	return s.Conditions != nil && *s.Conditions > 0 && s.CoveredConditions != nil && *s.CoveredConditions > 0 && *s.CoveredConditions < *s.Conditions
}
func (s sourceLine) hasNoBranchCoverage() bool {
	return s.Conditions != nil && *s.Conditions > 0 && (s.CoveredConditions == nil || *s.CoveredConditions == 0)
}

// Structured response matching Java GetFileCoverageDetailsToolResponse
type GetFileCoverageDetailsToolResponse struct {
	FileKey                  string                          `json:"fileKey"`
	FilePath                 string                          `json:"filePath,omitempty"`
	Summary                  CoverageSummary                 `json:"summary"`
	UncoveredLines           []UncoveredLine                 `json:"uncoveredLines"`
	PartiallyConditionalLines []PartiallyConditionalLine     `json:"partiallyConditionalLines"`
}
type CoverageSummary struct {
	TotalLines            int     `json:"totalLines"`
	CoverableLines        int     `json:"coverableLines"`
	CoveredLines          int     `json:"coveredLines"`
	UncoveredLines        int     `json:"uncoveredLines"`
	LineCoveragePercent   float64 `json:"lineCoveragePercent"`
	TotalConditions       int     `json:"totalConditions"`
	CoveredConditions     int     `json:"coveredConditions"`
	UncoveredConditions   int     `json:"uncoveredConditions"`
	BranchCoveragePercent float64 `json:"branchCoveragePercent"`
}
type UncoveredLine struct {
	LineNumber int `json:"lineNumber"`
}
type PartiallyConditionalLine struct {
	LineNumber          int `json:"lineNumber"`
	TotalConditions     int `json:"totalConditions"`
	CoveredConditions   int `json:"coveredConditions"`
	UncoveredConditions int `json:"uncoveredConditions"`
}

func NewGetFileCoverageDetailsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"get_file_coverage_details",
		"Get SonarQube File Coverage Details — Get complete line-by-line coverage information for a file, including which exact lines are uncovered and which have partially covered branches. This tool helps identify precisely where to add test coverage. Use after identifying files with low coverage via search_files_by_coverage.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"key": {"type": "string", "description": "File key (e.g. my_project:src/foo/Bar.java)."},
				"branch": {"type": "string", "description": "Branch name."},
				"pullRequest": {"type": "string", "description": "Pull request ID."}
			},
			"required": ["key"],
			"additionalProperties": false
	}`))
}

func GetFileCoverageDetailsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	fileKey := mcputils.GetOptionalString(args, "key")
	if fileKey == "" {
		return mcp.NewToolResultError("key is required"), nil
	}

	branch := mcputils.GetOptionalString(args, "branch")
	pr := mcputils.GetOptionalString(args, "pullRequest")
	if branch != "" && pr != "" {
		return mcp.NewToolResultError("branch and pullRequest cannot both be specified"), nil
	}

	params := url.Values{}
	params.Set("key", fileKey)
	if branch != "" {
		params.Set("branch", branch)
	}
	if pr != "" {
		params.Set("pullRequest", pr)
	}

	client := mcputils.NewSQClient()
	var resp sourceLinesResponse
	if err := client.DoGet(ctx, "/api/sources/lines", params, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to retrieve coverage details: %v", err)), nil
	}

	// Extract file path from key (after the colon)
	filePath := ""
	if idx := strings.Index(fileKey, ":"); idx >= 0 {
		filePath = fileKey[idx+1:]
	}

	// Compute summary statistics
	totalLines := len(resp.Sources)

	coverableCount := 0
	coveredCount := 0
	uncoveredCount := 0
	totalConditions := 0
	coveredConditions := 0

	var uncoveredLines []UncoveredLine
	var partialConditionalLines []PartiallyConditionalLine

	for _, s := range resp.Sources {
		if s.isCoverable() {
			coverableCount++
			if s.isUncovered() {
				uncoveredCount++
				uncoveredLines = append(uncoveredLines, UncoveredLine{LineNumber: s.Line})
			} else {
				coveredCount++
			}
		}

		cond := 0
		covCond := 0
		if s.Conditions != nil {
			cond = *s.Conditions
		}
		if s.CoveredConditions != nil {
			covCond = *s.CoveredConditions
		}
		totalConditions += cond
		coveredConditions += covCond

		if s.hasPartialBranchCoverage() || s.hasNoBranchCoverage() {
			partialConditionalLines = append(partialConditionalLines, PartiallyConditionalLine{
				LineNumber:          s.Line,
				TotalConditions:     cond,
				CoveredConditions:   covCond,
				UncoveredConditions: cond - covCond,
			})
		}
	}

	uncoveredConditions := totalConditions - coveredConditions

	lineCoveragePercent := 100.0
	if coverableCount > 0 {
		lineCoveragePercent = float64(coveredCount) * 100.0 / float64(coverableCount)
	}

	branchCoveragePercent := 100.0
	if totalConditions > 0 {
		branchCoveragePercent = float64(coveredConditions) * 100.0 / float64(totalConditions)
	}

	response := GetFileCoverageDetailsToolResponse{
		FileKey:  fileKey,
		FilePath: filePath,
		Summary: CoverageSummary{
			TotalLines:            totalLines,
			CoverableLines:        coverableCount,
			CoveredLines:          coveredCount,
			UncoveredLines:        uncoveredCount,
			LineCoveragePercent:   lineCoveragePercent,
			TotalConditions:       totalConditions,
			CoveredConditions:     coveredConditions,
			UncoveredConditions:   uncoveredConditions,
			BranchCoveragePercent: branchCoveragePercent,
		},
		UncoveredLines:            uncoveredLines,
		PartiallyConditionalLines: partialConditionalLines,
	}

	respJSON, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(respJSON)), nil
}
