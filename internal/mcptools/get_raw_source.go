package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

// GetRawSourceToolResponse matches the Java GetRawSourceToolResponse structure.
type GetRawSourceToolResponse struct {
	FileKey    string `json:"fileKey"`
	SourceCode string `json:"sourceCode"`
}

func NewGetRawSourceMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"get_raw_source",
		"Get SonarQube Raw Source Code — Get source code as raw text. Requires 'See Source Code' permission on file.",
		json.RawMessage(`{
			"type": "object",
			"properties": {
				"key": {"type": "string", "description": "File key (e.g. my_project:src/foo/Bar.php)."},
				"branch": {"type": "string", "description": "Branch name."},
				"pullRequest": {"type": "string", "description": "Pull request ID."}
			},
			"required": ["key"],
			"additionalProperties": false
		}`),
	)
}

func GetRawSourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	raw, err := client.DoGetRaw(ctx, "/api/sources/raw", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to retrieve source code: %v", err)), nil
	}

	response := GetRawSourceToolResponse{
		FileKey:    fileKey,
		SourceCode: raw,
	}

	respJSON, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(respJSON)), nil
}
