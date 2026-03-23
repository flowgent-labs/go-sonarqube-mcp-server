package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

// Raw API response structures for /api/duplications/list
type dupListResponse struct {
	Duplications []dupListEntry `json:"duplications"`
	Paging       dupListPaging  `json:"paging"`
}
type dupListEntry struct {
	File             string `json:"file"`
	Project          string `json:"project"`
	DuplicatedBlocks int    `json:"duplicatedBlocks"`
	DuplicatedLines  int    `json:"duplicatedLines"`
}
type dupListPaging struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	Total     int `json:"total"`
}

// Structured response matching Java SearchDuplicatedFilesToolResponse
type SearchDuplicatedFilesToolResponse struct {
	Files   []DuplicatedFile `json:"files"`
	Paging  DupPaging        `json:"paging"`
	Summary *Summary         `json:"summary,omitempty"`
}
type DuplicatedFile struct {
	Key                   string  `json:"key"`
	Name                  string  `json:"name"`
	Path                  *string `json:"path,omitempty"`
	DuplicatedLines       *int    `json:"duplicatedLines,omitempty"`
	DuplicatedBlocks      *int    `json:"duplicatedBlocks,omitempty"`
	DuplicatedLinesDensity *string `json:"duplicatedLinesDensity,omitempty"`
}
type Summary struct {
	TotalDuplicatedLines     *int    `json:"totalDuplicatedLines,omitempty"`
	TotalDuplicatedBlocks    *int    `json:"totalDuplicatedBlocks,omitempty"`
	OverallDuplicationDensity *string `json:"overallDuplicationDensity,omitempty"`
}
type DupPaging struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	Total     int `json:"total"`
}

func NewSearchDuplicatedFilesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"search_duplicated_files",
		"Search Duplicated Files — Search for files with code duplications in a project. Auto-fetches all pages by default.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"projectKey": {"type": "string", "description": "SonarQube project key. Required unless a default is configured via SONARQUBE_PROJECT_KEY."},
				"branch": {"type": "string", "description": "Branch name."},
				"pullRequest": {"type": "string", "description": "Pull request key."},
				"pageSize": {"type": "integer", "description": "Page size. Max 500.", "default": 100},
				"pageIndex": {"type": "integer", "description": "1-based page index. Defaults to 1.", "default": 1}
			},
			"additionalProperties": false
	}`))
}

func SearchDuplicatedFilesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	projectKey, err := mcputils.ResolveProjectKey(args, "projectKey")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := url.Values{}
	params.Set("project", projectKey)
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

	client := mcputils.NewSQClient()
	var resp dupListResponse
	if err := client.DoGet(ctx, "/api/duplications/list", params, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search duplicated files failed: %v", err)), nil
	}

	// Build files list
	files := make([]DuplicatedFile, 0, len(resp.Duplications))
	totalLines := 0
	totalBlocks := 0
	for _, d := range resp.Duplications {
		path := d.File
		name := filepath.Base(path)

		df := DuplicatedFile{
			Key:  path,
			Name: name,
		}
		if path != "" {
			df.Path = &path
		}
		if d.DuplicatedLines > 0 {
			dl := d.DuplicatedLines
			df.DuplicatedLines = &dl
		}
		if d.DuplicatedBlocks > 0 {
			db := d.DuplicatedBlocks
			df.DuplicatedBlocks = &db
		}

		totalLines += d.DuplicatedLines
		totalBlocks += d.DuplicatedBlocks

		files = append(files, df)
	}

	// Build summary
	summary := &Summary{}
	if totalLines > 0 {
		summary.TotalDuplicatedLines = &totalLines
	}
	if totalBlocks > 0 {
		summary.TotalDuplicatedBlocks = &totalBlocks
	}

	// Build response
	response := SearchDuplicatedFilesToolResponse{
		Files: files,
		Paging: DupPaging{
			PageIndex: resp.Paging.PageIndex,
			PageSize:  resp.Paging.PageSize,
			Total:     resp.Paging.Total,
		},
		Summary: summary,
	}

	respJSON, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(respJSON)), nil
}
