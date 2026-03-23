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
type duplicationsShowResponse struct {
	Duplications []dupBlockGroup   `json:"duplications"`
	Files        map[string]dupFileInfo `json:"files"`
}
type dupBlockGroup struct {
	Blocks []dupBlock `json:"blocks"`
}
type dupBlock struct {
	From int    `json:"from"`
	Size int    `json:"size"`
	Ref  string `json:"_ref"`
}
type dupFileInfo struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// Structured response matching Java GetDuplicationsToolResponse
type GetDuplicationsToolResponse struct {
	Duplications []GetDuplicationsDuplication `json:"duplications"`
	Files        []GetDuplicationsFileInfo    `json:"files"`
}
type GetDuplicationsDuplication struct {
	Blocks []GetDuplicationsBlock `json:"blocks"`
}
type GetDuplicationsBlock struct {
	From     int    `json:"from"`
	Size     int    `json:"size"`
	FileName string `json:"fileName"`
	FileKey  string `json:"fileKey"`
}
type GetDuplicationsFileInfo struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

func NewGetDuplicationsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"get_duplications",
		"Get SonarQube Code Duplications — Get duplications for a file. Requires Browse permission on file's project.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"key": {"type": "string", "description": "File key (e.g. my_project:src/foo/Bar.php)."},
				"branch": {"type": "string", "description": "Branch name."},
				"pullRequest": {"type": "string", "description": "Pull request ID."}
			},
			"required": ["key"],
			"additionalProperties": false
	}`))
}

func GetDuplicationsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	var resp duplicationsShowResponse
	if err := client.DoGet(ctx, "/api/duplications/show", params, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Get duplications failed: %v", err)), nil
	}

	// Build structured response matching Java GetDuplicationsToolResponse
	duplications := make([]GetDuplicationsDuplication, 0, len(resp.Duplications))
	for _, dg := range resp.Duplications {
		blocks := make([]GetDuplicationsBlock, 0, len(dg.Blocks))
		for _, block := range dg.Blocks {
			fileName := ""
			fileKey := ""
			if fi, ok := resp.Files[block.Ref]; ok {
				fileName = fi.Name
				fileKey = fi.Key
			}
			blocks = append(blocks, GetDuplicationsBlock{
				From:     block.From,
				Size:     block.Size,
				FileName: fileName,
				FileKey:  fileKey,
			})
		}
		duplications = append(duplications, GetDuplicationsDuplication{Blocks: blocks})
	}

	files := make([]GetDuplicationsFileInfo, 0, len(resp.Files))
	for _, fi := range resp.Files {
		files = append(files, GetDuplicationsFileInfo{
			Key:  fi.Key,
			Name: fi.Name,
		})
	}

	response := GetDuplicationsToolResponse{
		Duplications: duplications,
		Files:        files,
	}

	respJSON, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(respJSON)), nil
}
