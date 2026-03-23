package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

type a3sAnalysisRequest struct {
	OrganizationKey string `json:"organizationKey"`
	ProjectKey      string `json:"projectKey"`
	BranchName      string `json:"branchName,omitempty"`
	FilePath        string `json:"filePath"`
	FileContent     string `json:"fileContent"`
	FileScope       string `json:"fileScope,omitempty"`
}

func NewRunAdvancedCodeAnalysisMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"run_advanced_code_analysis",
		"SonarQube Advanced Code Analysis — Run advanced code analysis on a single file using SonarQube Cloud's server-side engine. Identifies code quality and security issues, leveraging the project's full analysis context for deeper cross-file detection. Always specify the file scope (MAIN or TEST) for more accurate results.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"projectKey": {"type": "string", "description": "SonarQube project key. Required unless a default is configured via SONARQUBE_PROJECT_KEY."},
				"branchName": {"type": "string", "description": "The branch name used to retrieve the latest analysis context from SonarQube Cloud."},
				"filePath": {"type": "string", "description": "Project-relative path of the file to analyze (e.g., 'src/main/java/MyClass.java')."},
				"fileContent": {"type": "string", "description": "Complete file content to analyze."},
				"fileScope": {"type": "string", "description": "Scope of the file: MAIN or TEST (default: MAIN).", "default": "MAIN"}
			},
			"required": ["projectKey", "branchName", "filePath", "fileContent"],
			"additionalProperties": false
	}`))
}

func RunAdvancedCodeAnalysisHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !mcputils.IsCloud() {
		return mcp.NewToolResultError("run_advanced_code_analysis is only available on SonarQube Cloud."), nil
	}

	org := mcputils.GetSonarQubeOrg()
	if org == "" {
		return mcp.NewToolResultError("SONARQUBE_ORG is required for run_advanced_code_analysis."), nil
	}

	args := request.GetArguments()

	projectKey, err := mcputils.ResolveProjectKey(args, "projectKey")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	branchName := mcputils.GetOptionalString(args, "branchName")
	if branchName == "" {
		return mcp.NewToolResultError("branchName is required."), nil
	}

	filePath := mcputils.GetOptionalString(args, "filePath")
	if filePath == "" {
		return mcp.NewToolResultError("filePath is required."), nil
	}

	fileContent := mcputils.GetOptionalString(args, "fileContent")
	if fileContent == "" {
		return mcp.NewToolResultError("fileContent is required."), nil
	}

	fileScope := mcputils.GetOptionalString(args, "fileScope")
	if fileScope == "" {
		fileScope = "MAIN"
	}

	reqBody := a3sAnalysisRequest{
		OrganizationKey: org,
		ProjectKey:      projectKey,
		BranchName:      branchName,
		FilePath:        filePath,
		FileContent:     fileContent,
		FileScope:       fileScope,
	}

	client := mcputils.NewSQClient()
	var resp map[string]interface{}
	if err := client.DoPostWithBody(ctx, "/a3s-analysis/analyses", url.Values{}, reqBody, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Advanced code analysis failed: %v", err)), nil
	}

	respJSON, _ := json.MarshalIndent(resp, "", "  ")
	return mcp.NewToolResultText(string(respJSON)), nil
}
