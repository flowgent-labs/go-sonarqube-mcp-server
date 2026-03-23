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

// Raw API response types for /api/sources/scm
type scmInfoRawResponse struct {
	Scm []scmInfoRawLine `json:"scm"`
}

type scmInfoRawLine struct {
	Line     int    `json:"line"`
	Author   string `json:"author"`
	Date     string `json:"date"`
	Revision string `json:"revision"`
}

// Tool response types matching Java GetScmInfoToolResponse
type GetScmInfoToolResponse struct {
	ScmLines []ScmLine `json:"scmLines"`
}

type ScmLine struct {
	LineNumber int    `json:"lineNumber"`
	Author     string `json:"author"`
	DateTime   string `json:"datetime"`
	Revision   string `json:"revision"`
}

func NewGetScmInfoMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"get_scm_info",
		"Get SCM Info — Get SCM information (author, date, revision) per line of a source file.",
		json.RawMessage(
			`{
				"type": "object",
				"properties": {
					"key": {"type": "string", "description": "File key (e.g. my_project:src/foo/Bar.php)."},
					"commits_by_line": {"type": "boolean", "description": "Show commits per line."},
					"from": {"type": "integer", "description": "First line (1-based)."},
					"to": {"type": "integer", "description": "Last line (1-based)."}
				},
				"required": ["key"],
				"additionalProperties": false
	}`))
}

func GetScmInfoHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	fileKey := mcputils.GetOptionalString(args, "key")
	if fileKey == "" {
		return mcp.NewToolResultError("key is required"), nil
	}

	params := url.Values{}
	params.Set("key", fileKey)
	if mcputils.GetBoolOrDefault(args, "commits_by_line", false) {
		params.Set("commits_by_line", "true")
	}
	if v, ok := args["from"]; ok {
		if f, ok := v.(float64); ok {
			params.Set("from", strconv.Itoa(int(f)))
		}
	}
	if v, ok := args["to"]; ok {
		if f, ok := v.(float64); ok {
			params.Set("to", strconv.Itoa(int(f)))
		}
	}

	client := mcputils.NewSQClient()
	var rawResp scmInfoRawResponse
	if err := client.DoGet(ctx, "/api/sources/scm", params, &rawResp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Get SCM info failed: %v", err)), nil
	}

	resp := GetScmInfoToolResponse{
		ScmLines: make([]ScmLine, len(rawResp.Scm)),
	}
	for i, line := range rawResp.Scm {
		resp.ScmLines[i] = ScmLine{
			LineNumber: line.Line,
			Author:     line.Author,
			DateTime:   line.Date,
			Revision:   line.Revision,
		}
	}

	jsonBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
	}
	return mcp.NewToolResultText(string(jsonBytes)), nil
}
