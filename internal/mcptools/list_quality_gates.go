package mcptools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

// Raw API response structures
type qualityGatesListResponse struct {
	QualityGates []qualityGateEntry `json:"qualitygates"`
}
type qualityGateEntry struct {
	ID                   int                        `json:"id"`
	Name                 string                     `json:"name"`
	IsDefault            bool                       `json:"isDefault"`
	IsBuiltIn            bool                       `json:"isBuiltIn"`
	Conditions           []qualityGateCondition     `json:"conditions,omitempty"`
	CaycStatus           string                     `json:"caycStatus,omitempty"`
	HasStandardConditions bool                      `json:"hasStandardConditions"`
	HasMQRConditions     bool                       `json:"hasMQRConditions"`
	IsAiCodeSupported    bool                       `json:"isAiCodeSupported"`
}
type qualityGateCondition struct {
	Metric string `json:"metric"`
	Op     string `json:"op"`
	Error  string `json:"error"`
}

// Structured response matching Java ListQualityGatesToolResponse
type ListQualityGatesToolResponse struct {
	QualityGates []ListQualityGatesQG `json:"qualityGates"`
}
type ListQualityGatesQG struct {
	ID                   int                         `json:"id"`
	Name                 string                      `json:"name"`
	IsDefault            bool                        `json:"isDefault"`
	IsBuiltIn            bool                        `json:"isBuiltIn"`
	Conditions           []ListQualityGatesCondition `json:"conditions,omitempty"`
	CaycStatus           string                      `json:"caycStatus,omitempty"`
	HasStandardConditions bool                       `json:"hasStandardConditions"`
	HasMQRConditions     bool                        `json:"hasMQRConditions"`
	IsAiCodeSupported    bool                        `json:"isAiCodeSupported"`
}
type ListQualityGatesCondition struct {
	Metric string `json:"metric"`
	Op     string `json:"op"`
	Error  int    `json:"error"`
}

func NewListQualityGatesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"list_quality_gates",
		"List SonarQube Quality Gates — List all quality gates.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {},
			"additionalProperties": false
	}`))
}

func ListQualityGatesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := mcputils.NewSQClient()
	var resp qualityGatesListResponse
	if err := client.DoGet(ctx, "/api/qualitygates/list", nil, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("List quality gates failed: %v", err)), nil
	}

	qualityGates := make([]ListQualityGatesQG, 0, len(resp.QualityGates))
	for _, qg := range resp.QualityGates {
		conditions := make([]ListQualityGatesCondition, 0, len(qg.Conditions))
		for _, c := range qg.Conditions {
			errVal := 0
			fmt.Sscanf(c.Error, "%d", &errVal)
			conditions = append(conditions, ListQualityGatesCondition{
				Metric: c.Metric,
				Op:     c.Op,
				Error:  errVal,
			})
		}
		qualityGates = append(qualityGates, ListQualityGatesQG{
			ID:                    qg.ID,
			Name:                  qg.Name,
			IsDefault:             qg.IsDefault,
			IsBuiltIn:             qg.IsBuiltIn,
			Conditions:            conditions,
			CaycStatus:            qg.CaycStatus,
			HasStandardConditions: qg.HasStandardConditions,
			HasMQRConditions:      qg.HasMQRConditions,
			IsAiCodeSupported:     qg.IsAiCodeSupported,
		})
	}

	response := ListQualityGatesToolResponse{QualityGates: qualityGates}
	respJSON, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(respJSON)), nil
}
