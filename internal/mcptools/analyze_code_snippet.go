package mcptools

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
)

func NewAnalyzeCodeSnippetMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"analyze_code_snippet",
		"SonarQube Code Analysis — Analyze a file or code snippet to identify code quality and security issues. Optionally provide a code snippet to filter issues — only issues within the snippet will be reported (snippet location is auto-detected). Always specify the language and the file scope (MAIN or TEST) for more accurate results.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"projectKey": {"type": "string", "description": "SonarQube project key."},
				"filePath": {"type": "string", "description": "Project-relative path of the file to analyze (e.g., 'src/main/java/MyClass.java')."},
				"fileContent": {"type": "string", "description": "Complete file content to analyze."},
				"codeSnippet": {"type": "string", "description": "Code snippet to filter issues - must match content within the analyzed file."},
				"language": {"type": "string", "description": "Language of the code (e.g., 'java', 'python', 'ts', 'tsx', 'js', 'jsx')."},
				"scope": {"type": "string", "description": "Scope of the file: MAIN or TEST (default: MAIN).", "default": "MAIN"}
			},
			"additionalProperties": false
	}`))
}

func AnalyzeCodeSnippetHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError(
		"analyze_code_snippet requires the SonarQube MCP Server backend service with local analysis capabilities. " +
			"This tool is only available when running the official SonarQube MCP Server JAR (sonarqube-mcp-server). " +
			"The standalone Go MCP server does not include the embedded SonarLint analyzers needed for code analysis. " +
			"Please connect to a SonarQube MCP Server instance instead."), nil
}
