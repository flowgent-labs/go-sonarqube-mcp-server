package mcptools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

// ToolEntry pairs an MCP tool definition with its handler function.
type ToolEntry struct {
	Tool    mcp.Tool
	Handler func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
}

// Registry maps tool names to their ToolEntry for dynamic tool discovery.
var Registry = map[string]ToolEntry{
	// Projects
	"search_my_sonarqube_projects": {Tool: NewSearchMyProjectsMCPTool(), Handler: SearchMyProjectsHandler},
	"list_branches":                {Tool: NewListBranchesMCPTool(), Handler: ListBranchesHandler},
	"list_pull_requests":           {Tool: NewListPullRequestsMCPTool(), Handler: ListPullRequestsHandler},

	// Issues
	"search_sonar_issues_in_projects": {Tool: NewSearchIssuesMCPTool(), Handler: SearchIssuesHandler},
	"change_sonar_issue_status":       {Tool: NewChangeIssueStatusMCPTool(), Handler: ChangeIssueStatusHandler},

	// Security Hotspots
	"search_security_hotspots":       {Tool: NewSearchSecurityHotspotsMCPTool(), Handler: SearchSecurityHotspotsHandler},
	"show_security_hotspot":          {Tool: NewShowSecurityHotspotMCPTool(), Handler: ShowSecurityHotspotHandler},
	"change_security_hotspot_status": {Tool: NewChangeSecurityHotspotStatusMCPTool(), Handler: ChangeSecurityHotspotStatusHandler},

	// Quality Gates
	"list_quality_gates":              {Tool: NewListQualityGatesMCPTool(), Handler: ListQualityGatesHandler},
	"get_project_quality_gate_status": {Tool: NewProjectQualityGateStatusMCPTool(), Handler: ProjectQualityGateStatusHandler},

	// Measures
	"get_component_measures":   {Tool: NewGetComponentMeasuresMCPTool(), Handler: GetComponentMeasuresHandler},
	"search_files_by_coverage": {Tool: NewSearchFilesByCoverageMCPTool(), Handler: SearchFilesByCoverageHandler},

	// Coverage
	"get_file_coverage_details": {Tool: NewGetFileCoverageDetailsMCPTool(), Handler: GetFileCoverageDetailsHandler},

	// Rules
	"show_rule": {Tool: NewShowRuleMCPTool(), Handler: ShowRuleHandler},

	// Duplications
	"get_duplications":        {Tool: NewGetDuplicationsMCPTool(), Handler: GetDuplicationsHandler},
	"search_duplicated_files": {Tool: NewSearchDuplicatedFilesMCPTool(), Handler: SearchDuplicatedFilesHandler},

	// Languages
	"list_languages": {Tool: NewListLanguagesMCPTool(), Handler: ListLanguagesHandler},

	// Metrics
	"search_metrics": {Tool: NewSearchMetricsMCPTool(), Handler: SearchMetricsHandler},

	// System
	"ping_system":       {Tool: NewSystemPingMCPTool(), Handler: SystemPingHandler},
	"get_system_status": {Tool: NewSystemStatusMCPTool(), Handler: SystemStatusHandler},
	"get_system_health": {Tool: NewSystemHealthMCPTool(), Handler: SystemHealthHandler},
	"get_system_info":   {Tool: NewSystemInfoMCPTool(), Handler: SystemInfoHandler},
	"get_system_logs":   {Tool: NewSystemLogsMCPTool(), Handler: SystemLogsHandler},

	// Portfolios
	"list_portfolios": {Tool: NewListPortfoliosMCPTool(), Handler: ListPortfoliosHandler},

	// Sources
	"get_raw_source": {Tool: NewGetRawSourceMCPTool(), Handler: GetRawSourceHandler},
	"get_scm_info":   {Tool: NewGetScmInfoMCPTool(), Handler: GetScmInfoHandler},

	// Webhooks
	"list_webhooks":  {Tool: NewListWebhooksMCPTool(), Handler: ListWebhooksHandler},
	"create_webhook": {Tool: NewCreateWebhookMCPTool(), Handler: CreateWebhookHandler},

	// Dependency Risks
	"search_dependency_risks": {Tool: NewSearchDependencyRisksMCPTool(), Handler: SearchDependencyRisksHandler},

	// Enterprises (Cloud-only)
	"list_enterprises": {Tool: NewListEnterprisesMCPTool(), Handler: ListEnterprisesHandler},

	// Analysis
	"analyze_code_snippet":         {Tool: NewAnalyzeCodeSnippetMCPTool(), Handler: AnalyzeCodeSnippetHandler},
	"analyze_file_list":            {Tool: NewAnalyzeFileListMCPTool(), Handler: AnalyzeFileListHandler},
	"toggle_automatic_analysis":    {Tool: NewToggleAutomaticAnalysisMCPTool(), Handler: ToggleAutomaticAnalysisHandler},
	"run_advanced_code_analysis":   {Tool: NewRunAdvancedCodeAnalysisMCPTool(), Handler: RunAdvancedCodeAnalysisHandler},
}
