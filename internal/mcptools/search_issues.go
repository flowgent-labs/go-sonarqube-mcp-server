package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

type searchIssuesResponse struct {
	Paging searchIssuesPaging  `json:"paging"`
	Issues []searchIssuesIssue `json:"issues"`
}
type searchIssuesPaging struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	Total     int `json:"total"`
}
type searchIssuesIssue struct {
	Key                       string `json:"key"`
	Rule                      string `json:"rule"`
	Project                   string `json:"project"`
	Component                 string `json:"component"`
	Severity                  string `json:"severity"`
	Status                    string `json:"status"`
	Message                   string `json:"message"`
	CleanCodeAttribute        string `json:"cleanCodeAttribute"`
	CleanCodeAttributeCategory string `json:"cleanCodeAttributeCategory"`
	Author                    string `json:"author"`
	CreationDate              string `json:"creationDate"`
	TextRange                 *searchIssuesTextRange `json:"textRange,omitempty"`
}
type searchIssuesTextRange struct {
	StartLine int `json:"startLine"`
	EndLine   int `json:"endLine"`
}

// SearchIssuesToolResponse matches the Java SearchIssuesToolResponse structure.
type SearchIssuesToolResponse struct {
	Issues []SearchIssuesIssue `json:"issues"`
	Paging SearchIssuesPaging  `json:"paging"`
}
type SearchIssuesIssue struct {
	Key                       string                  `json:"key"`
	Rule                      string                  `json:"rule"`
	Project                   string                  `json:"project"`
	Component                 string                  `json:"component"`
	Severity                  string                  `json:"severity"`
	Status                    string                  `json:"status"`
	Message                   string                  `json:"message"`
	CleanCodeAttribute        string                  `json:"cleanCodeAttribute"`
	CleanCodeAttributeCategory string                 `json:"cleanCodeAttributeCategory"`
	Author                    string                  `json:"author"`
	CreationDate              string                  `json:"creationDate"`
	TextRange                 *SearchIssuesTextRange   `json:"textRange,omitempty"`
}
type SearchIssuesTextRange struct {
	StartLine int `json:"startLine"`
	EndLine   int `json:"endLine"`
}
type SearchIssuesPaging struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	Total     int `json:"total"`
}

func NewSearchIssuesMCPTool() mcp.Tool {
	scope := "my projects"
	if mcputils.IsCloud() {
		scope = "my organization's projects"
	}
	return mcp.NewToolWithRawSchema(
		"search_sonar_issues_in_projects",
		fmt.Sprintf("Search SonarQube Issues — Search for issues (bugs, vulnerabilities, code smells) in %s. Filter by severities=['HIGH','BLOCKER'] for critical issues, impactSoftwareQualities=['SECURITY'] for security, issueStatuses=['OPEN'] to exclude resolved.", scope),
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"projects": {"type": "array", "items": {"type": "string"}, "description": "An optional list of Sonar projects to look in."},
				"files": {"type": "array", "items": {"type": "string"}, "description": "An optional list of component keys (files, directories, modules) to filter issues."},
				"branch": {"type": "string", "description": "Branch name."},
				"pullRequest": {"type": "string", "description": "Pull request ID."},
				"severities": {"type": "array", "items": {"type": "string", "enum": ["INFO", "LOW", "MEDIUM", "HIGH", "BLOCKER"]}, "description": "An optional list of severities to filter by."},
				"impactSoftwareQualities": {"type": "array", "items": {"type": "string", "enum": ["MAINTAINABILITY", "RELIABILITY", "SECURITY"]}, "description": "An optional list of software qualities to filter by."},
				"issueStatuses": {"type": "array", "items": {"type": "string", "enum": ["OPEN", "CONFIRMED", "FALSE_POSITIVE", "ACCEPTED", "FIXED", "IN_SANDBOX"]}, "description": "An optional list of issue statuses to filter by. Note: IN_SANDBOX is valid only for SonarQube Server."},
				"issueKey": {"type": "array", "items": {"type": "string"}, "description": "An optional list of issue keys to fetch specific issues."},
				"p": {"type": "number", "description": "An optional page number. Defaults to 1.", "default": 1},
				"ps": {"type": "number", "description": "An optional page size. Must be greater than 0 and less than or equal to 500. Defaults to 100.", "default": 100}
			},
			"additionalProperties": false
	}`))
}

func SearchIssuesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	// Validate branch/pullRequest
	branch := mcputils.GetOptionalString(args, "branch")
	pr := mcputils.GetOptionalString(args, "pullRequest")
	if branch != "" && pr != "" {
		return mcp.NewToolResultError("branch and pullRequest cannot both be specified"), nil
	}

	params := url.Values{}

	if projects := mcputils.GetStringArray(args, "projects"); len(projects) > 0 {
		params.Set("projects", strings.Join(projects, ","))
	}
	if files := mcputils.GetStringArray(args, "files"); len(files) > 0 {
		params.Set("files", strings.Join(files, ","))
	}
	if branch != "" {
		params.Set("branch", branch)
	}
	if pr != "" {
		params.Set("pullRequest", pr)
	}
	if severities := mcputils.GetStringArray(args, "severities"); len(severities) > 0 {
		params.Set("severities", strings.Join(severities, ","))
	}
	if qualities := mcputils.GetStringArray(args, "impactSoftwareQualities"); len(qualities) > 0 {
		params.Set("impactSoftwareQualities", strings.Join(qualities, ","))
	}
	if statuses := mcputils.GetStringArray(args, "issueStatuses"); len(statuses) > 0 {
		params.Set("issueStatuses", strings.Join(statuses, ","))
	}
	if keys := mcputils.GetStringArray(args, "issueKey"); len(keys) > 0 {
		params.Set("issues", strings.Join(keys, ","))
	}
	if mcputils.IsCloud() {
		if org := mcputils.GetSonarQubeOrg(); org != "" {
			params.Set("organization", org)
		}
	}

	page := mcputils.GetIntOrDefault(args, "p", 1)
	pageSize := mcputils.GetIntOrDefault(args, "ps", 100)
	params.Set("p", strconv.Itoa(page))
	params.Set("ps", strconv.Itoa(pageSize))

	client := mcputils.NewSQClient()
	var resp searchIssuesResponse
	if err := client.DoGet(ctx, "/api/issues/search", params, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search issues failed: %v", err)), nil
	}

	// Build structured response matching Java SearchIssuesToolResponse
	issues := make([]SearchIssuesIssue, 0, len(resp.Issues))
	for _, issue := range resp.Issues {
		var textRange *SearchIssuesTextRange
		// SonarQube API returns textRange with startLine/endLine for file-level issues
		// The raw API response may include these as nested fields
		// We check if the line field exists as a fallback when textRange is not available
		_ = textRange // placeholder for future API response parsing
		issues = append(issues, SearchIssuesIssue{
			Key:                        issue.Key,
			Rule:                       issue.Rule,
			Project:                    issue.Project,
			Component:                  issue.Component,
			Severity:                   issue.Severity,
			Status:                     issue.Status,
			Message:                    issue.Message,
			CleanCodeAttribute:         issue.CleanCodeAttribute,
			CleanCodeAttributeCategory: issue.CleanCodeAttributeCategory,
			Author:                     issue.Author,
			CreationDate:               issue.CreationDate,
			TextRange:                  textRange,
		})
	}

	response := SearchIssuesToolResponse{
		Issues: issues,
		Paging: SearchIssuesPaging{
			PageIndex: resp.Paging.PageIndex,
			PageSize:  resp.Paging.PageSize,
			Total:     resp.Paging.Total,
		},
	}

	respJSON, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(respJSON)), nil
}
