package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

// API response types (matching SonarQube /api/hotspots/show response)
type showHotspotRawResponse struct {
	Key                      string               `json:"key"`
	Component                string               `json:"component"`
	Project                  string               `json:"project"`
	SecurityCategory         string               `json:"securityCategory"`
	VulnerabilityProbability string               `json:"vulnerabilityProbability"`
	Status                   string               `json:"status"`
	Resolution               string               `json:"resolution"`
	Line                     int                  `json:"line"`
	Message                  string               `json:"message"`
	Assignee                 string               `json:"assignee"`
	Author                   string               `json:"author"`
	CreationDate             string               `json:"creationDate"`
	UpdateDate               string               `json:"updateDate"`
	TextRange                *apiTextRange        `json:"textRange"`
	Flows                    []showHotspotFlow    `json:"flows"`
	Comments                 []showHotspotComment `json:"comments"`
	Rule                     showHotspotRule      `json:"rule"`
	CanChangeStatus          bool                 `json:"canChangeStatus"`
}

type showHotspotFlow struct {
	Locations []showHotspotLocation `json:"locations"`
}

type showHotspotLocation struct {
	Component string        `json:"component"`
	TextRange *apiTextRange `json:"textRange"`
	Msg       string        `json:"msg"`
}

type showHotspotComment struct {
	Key       string `json:"key"`
	Login     string `json:"login"`
	HtmlText  string `json:"htmlText"`
	Markdown  string `json:"markdown"`
	Updatable bool   `json:"updatable"`
	CreatedAt string `json:"createdAt"`
}

type showHotspotRule struct {
	Key                      string `json:"key"`
	Name                     string `json:"name"`
	SecurityCategory         string `json:"securityCategory"`
	VulnerabilityProbability string `json:"vulnerabilityProbability"`
	RiskDescription          string `json:"riskDescription"`
	VulnerabilityDescription string `json:"vulnerabilityDescription"`
	FixRecommendations       string `json:"fixRecommendations"`
}

// Tool response types (matching Java ShowSecurityHotspotToolResponse)
type ShowSecurityHotspotToolResponse struct {
	Key                      string     `json:"key"`
	Component                string     `json:"component"`
	Project                  string     `json:"project"`
	SecurityCategory         string     `json:"securityCategory"`
	VulnerabilityProbability string     `json:"vulnerabilityProbability"`
	Status                   string     `json:"status"`
	Resolution               string     `json:"resolution"`
	Line                     int        `json:"line"`
	Message                  string     `json:"message"`
	Assignee                 string     `json:"assignee"`
	Author                   string     `json:"author"`
	CreationDate             string     `json:"creationDate"`
	UpdateDate               string     `json:"updateDate"`
	TextRange                *TextRange `json:"textRange"`
	Flows                    []Flow     `json:"flows"`
	Comments                 []Comment  `json:"comments"`
	Rule                     Rule       `json:"rule"`
	CanChangeStatus          bool       `json:"canChangeStatus"`
}

type Flow struct {
	Locations []Location `json:"locations"`
}

type Location struct {
	Component string     `json:"component"`
	TextRange *TextRange `json:"textRange"`
	Msg       string     `json:"msg"`
}

type Comment struct {
	Key       string `json:"key"`
	Login     string `json:"login"`
	HtmlText  string `json:"htmlText"`
	Markdown  string `json:"markdown"`
	Updatable bool   `json:"updatable"`
	CreatedAt string `json:"createdAt"`
}

type Rule struct {
	Key                      string `json:"key"`
	Name                     string `json:"name"`
	SecurityCategory         string `json:"securityCategory"`
	VulnerabilityProbability string `json:"vulnerabilityProbability"`
	RiskDescription          string `json:"riskDescription"`
	VulnerabilityDescription string `json:"vulnerabilityDescription"`
	FixRecommendations       string `json:"fixRecommendations"`
}

func NewShowSecurityHotspotMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"show_security_hotspot",
		"Show Security Hotspot — Get detailed information about a specific security hotspot (rule details, code context, flows, comments).",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"hotspotKey": {"type": "string", "description": "Security hotspot key."}
			},
			"required": ["hotspotKey"],
			"additionalProperties": false
}`))
}

func ShowSecurityHotspotHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	hotspotKey := mcputils.GetOptionalString(args, "hotspotKey")
	if hotspotKey == "" {
		return mcp.NewToolResultError("hotspotKey is required"), nil
	}

	params := url.Values{}
	params.Set("hotspot", hotspotKey)

	client := mcputils.NewSQClient()

	var raw showHotspotRawResponse
	if err := client.DoGet(ctx, "/api/hotspots/show", params, &raw); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Show hotspot failed: %v", err)), nil
	}

	response := ShowSecurityHotspotToolResponse{
		Key:                      raw.Key,
		Component:                raw.Component,
		Project:                  raw.Project,
		SecurityCategory:         raw.SecurityCategory,
		VulnerabilityProbability: raw.VulnerabilityProbability,
		Status:                   raw.Status,
		Resolution:               raw.Resolution,
		Line:                     raw.Line,
		Message:                  raw.Message,
		Assignee:                 raw.Assignee,
		Author:                   raw.Author,
		CreationDate:             raw.CreationDate,
		UpdateDate:               raw.UpdateDate,
		Rule: Rule{
			Key:                      raw.Rule.Key,
			Name:                     raw.Rule.Name,
			SecurityCategory:         raw.Rule.SecurityCategory,
			VulnerabilityProbability: raw.Rule.VulnerabilityProbability,
			RiskDescription:          raw.Rule.RiskDescription,
			VulnerabilityDescription: raw.Rule.VulnerabilityDescription,
			FixRecommendations:       raw.Rule.FixRecommendations,
		},
		CanChangeStatus: raw.CanChangeStatus,
	}

	if raw.TextRange != nil {
		response.TextRange = &TextRange{
			StartLine:   raw.TextRange.StartLine,
			EndLine:     raw.TextRange.EndLine,
			StartOffset: raw.TextRange.StartOffset,
			EndOffset:   raw.TextRange.EndOffset,
		}
	}

	response.Flows = make([]Flow, 0, len(raw.Flows))
	for _, f := range raw.Flows {
		flow := Flow{
			Locations: make([]Location, 0, len(f.Locations)),
		}
		for _, loc := range f.Locations {
			l := Location{
				Component: loc.Component,
				Msg:       loc.Msg,
			}
			if loc.TextRange != nil {
				l.TextRange = &TextRange{
					StartLine:   loc.TextRange.StartLine,
					EndLine:     loc.TextRange.EndLine,
					StartOffset: loc.TextRange.StartOffset,
					EndOffset:   loc.TextRange.EndOffset,
				}
			}
			flow.Locations = append(flow.Locations, l)
		}
		response.Flows = append(response.Flows, flow)
	}

	response.Comments = make([]Comment, 0, len(raw.Comments))
	for _, c := range raw.Comments {
		response.Comments = append(response.Comments, Comment{
			Key:       c.Key,
			Login:     c.Login,
			HtmlText:  c.HtmlText,
			Markdown:  c.Markdown,
			Updatable: c.Updatable,
			CreatedAt: c.CreatedAt,
		})
	}

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
	}
	return mcp.NewToolResultText(string(jsonBytes)), nil
}
