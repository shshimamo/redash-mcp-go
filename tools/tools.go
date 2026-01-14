package tools

import (
	"encoding/json"
	"fmt"

	"github.com/shshimamo/redash-mcp-go/mcp"
	"github.com/shshimamo/redash-mcp-go/redash"
)

// Handler は MCP ツールのハンドラー
type Handler struct {
	redashClient *redash.Client
}

// NewHandler は新しいツールハンドラーを作成
func NewHandler(redashClient *redash.Client) *Handler {
	return &Handler{
		redashClient: redashClient,
	}
}

// GetTools は利用可能な MCP ツールのリストを返す
func (h *Handler) GetTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name:        "get_query",
			Description: "Get metadata of a saved Redash query (name, description, SQL, etc.)",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"query_id": {
						Type:        "number",
						Description: "The ID of the query to get",
					},
				},
				Required: []string{"query_id"},
			},
		},
		{
			Name:        "get_dashboard",
			Description: "Get dashboard details including widgets and their associated queries",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"dashboard_id": {
						Type:        "number",
						Description: "The ID of the dashboard to get",
					},
				},
				Required: []string{"dashboard_id"},
			},
		},
		{
			Name:        "execute_query",
			Description: "Execute a saved Redash query by its ID and return the results",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"query_id": {
						Type:        "number",
						Description: "The ID of the query to execute",
					},
					"parameters": {
						Type:        "object",
						Description: "Optional parameters for the query (key-value pairs)",
					},
				},
				Required: []string{"query_id"},
			},
		},
		{
			Name:        "execute_adhoc_query",
			Description: "Execute an ad-hoc SQL query directly",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"query": {
						Type:        "string",
						Description: "The SQL query to execute",
					},
					"data_source_id": {
						Type:        "number",
						Description: "The ID of the data source to use",
					},
				},
				Required: []string{"query", "data_source_id"},
			},
		},
	}
}

// CallTool は指定された MCP ツールを実行
func (h *Handler) CallTool(name string, arguments map[string]interface{}) mcp.CallToolResult {
	switch name {
	case "get_query":
		return h.getQuery(arguments)
	case "get_dashboard":
		return h.getDashboard(arguments)
	case "execute_query":
		return h.executeQuery(arguments)
	case "execute_adhoc_query":
		return h.executeAdhocQuery(arguments)
	default:
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Unknown tool: %s", name),
				},
			},
			IsError: true,
		}
	}
}

// getQuery はクエリのメタデータを取得
func (h *Handler) getQuery(args map[string]interface{}) mcp.CallToolResult {
	// query_id の取得
	queryIDFloat, ok := args["query_id"].(float64)
	if !ok {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: "query_id must be a number",
				},
			},
			IsError: true,
		}
	}
	queryID := int(queryIDFloat)

	// Redash API を呼び出し
	query, err := h.redashClient.GetQuery(queryID)
	if err != nil {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Failed to get query: %v", err),
				},
			},
			IsError: true,
		}
	}

	// JSON として整形して返す
	formatted, err := json.MarshalIndent(query, "", "  ")
	if err != nil {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Failed to format query: %v", err),
				},
			},
			IsError: true,
		}
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: string(formatted),
			},
		},
		IsError: false,
	}
}

// getDashboard はダッシュボードのメタデータを取得
func (h *Handler) getDashboard(args map[string]interface{}) mcp.CallToolResult {
	// dashboard_id の取得
	dashboardIDFloat, ok := args["dashboard_id"].(float64)
	if !ok {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: "dashboard_id must be a number",
				},
			},
			IsError: true,
		}
	}
	dashboardID := int(dashboardIDFloat)

	// Redash API を呼び出し
	dashboard, err := h.redashClient.GetDashboard(dashboardID)
	if err != nil {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Failed to get dashboard: %v", err),
				},
			},
			IsError: true,
		}
	}

	// JSON として整形して返す
	formatted, err := json.MarshalIndent(dashboard, "", "  ")
	if err != nil {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Failed to format dashboard: %v", err),
				},
			},
			IsError: true,
		}
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: string(formatted),
			},
		},
		IsError: false,
	}
}

// executeQuery は保存済みクエリを実行
func (h *Handler) executeQuery(args map[string]interface{}) mcp.CallToolResult {
	// query_id の取得
	queryIDFloat, ok := args["query_id"].(float64)
	if !ok {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: "query_id must be a number",
				},
			},
			IsError: true,
		}
	}
	queryID := int(queryIDFloat)

	// parameters の取得（オプション）
	parameters := make(map[string]interface{})
	if params, ok := args["parameters"].(map[string]interface{}); ok {
		parameters = params
	}

	// Redash API を呼び出し
	result, err := h.redashClient.ExecuteQuery(queryID, parameters)
	if err != nil {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Failed to execute query: %v", err),
				},
			},
			IsError: true,
		}
	}

	// 結果を整形
	formatted, err := h.formatQueryResult(result)
	if err != nil {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Failed to format result: %v", err),
				},
			},
			IsError: true,
		}
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: formatted,
			},
		},
		IsError: false,
	}
}

// executeAdhocQuery はアドホッククエリを実行
func (h *Handler) executeAdhocQuery(args map[string]interface{}) mcp.CallToolResult {
	// query の取得
	query, ok := args["query"].(string)
	if !ok {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: "query must be a string",
				},
			},
			IsError: true,
		}
	}

	// data_source_id の取得
	dataSourceIDFloat, ok := args["data_source_id"].(float64)
	if !ok {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: "data_source_id must be a number",
				},
			},
			IsError: true,
		}
	}
	dataSourceID := int(dataSourceIDFloat)

	// Redash API を呼び出し
	result, err := h.redashClient.ExecuteAdhocQuery(query, dataSourceID)
	if err != nil {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Failed to execute query: %v", err),
				},
			},
			IsError: true,
		}
	}

	// 結果を整形
	formatted, err := h.formatQueryResult(result)
	if err != nil {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Failed to format result: %v", err),
				},
			},
			IsError: true,
		}
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: formatted,
			},
		},
		IsError: false,
	}
}

// formatQueryResult はクエリ結果を読みやすい形式に整形
func (h *Handler) formatQueryResult(result json.RawMessage) (string, error) {
	var data redash.QueryResultData
	if err := json.Unmarshal(result, &data); err != nil {
		return "", fmt.Errorf("failed to parse result: %w", err)
	}

	// JSON として整形して返す
	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format result: %w", err)
	}

	return string(formatted), nil
}
