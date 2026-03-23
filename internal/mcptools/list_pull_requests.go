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
type pullRequestsResponse struct {
	PullRequests []pullRequestEntry `json:"pullRequests"`
}
type pullRequestEntry struct {
	Key    string `json:"key"`
	Title  string `json:"title"`
	Branch string `json:"branch"`
}

// Structured response matching Java ListPullRequestsToolResponse
type ListPullRequestsToolResponse struct {
	ProjectKey        string                    `json:"projectKey"`
	TotalPullRequests int                       `json:"totalPullRequests"`
	PullRequests      []ListPullRequestsPR      `json:"pullRequests"`
}
type ListPullRequestsPR struct {
	Key    string `json:"key"`
	Title  string `json:"title"`
	Branch string `json:"branch"`
}

func NewListPullRequestsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"list_pull_requests",
		"List SonarQube Pull Requests — List all pull requests for a project. Use this tool to discover available pull requests and their corresponding branch names before analyzing their coverage, issues, or quality. Returns the pull request key/ID and source branch for each PR, which can be used with other tools that accept a pullRequest parameter. For long-lived branches (main, develop), use list_branches instead.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"projectKey": {"type": "string", "description": "SonarQube project key. Required unless a default is configured via SONARQUBE_PROJECT_KEY."}
			},
			"additionalProperties": false
	}`))
}

func ListPullRequestsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	projectKey, err := mcputils.ResolveProjectKey(args, "projectKey")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := url.Values{}
	params.Set("project", projectKey)

	client := mcputils.NewSQClient()
	var resp pullRequestsResponse
	if err := client.DoGet(ctx, "/api/project_pull_requests/list", params, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("List pull requests failed: %v", err)), nil
	}

	pullRequests := make([]ListPullRequestsPR, 0, len(resp.PullRequests))
	for _, pr := range resp.PullRequests {
		pullRequests = append(pullRequests, ListPullRequestsPR{
			Key:    pr.Key,
			Title:  pr.Title,
			Branch: pr.Branch,
		})
	}

	response := ListPullRequestsToolResponse{
		ProjectKey:        projectKey,
		TotalPullRequests: len(pullRequests),
		PullRequests:      pullRequests,
	}

	respJSON, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(respJSON)), nil
}
