package mcp

import "encoding/json"

// JSON-RPC 2.0 の基本構造
// MCP は JSON-RPC 2.0 プロトコルを使用します

// Request は JSON-RPC 2.0 リクエスト
type Request struct {
	JSONRPC string          `json:"jsonrpc"` // 常に "2.0"
	ID      interface{}     `json:"id"`      // リクエストID (数値または文字列)
	Method  string          `json:"method"`  // 呼び出すメソッド名
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response は JSON-RPC 2.0 レスポンス
type Response struct {
	JSONRPC string      `json:"jsonrpc"` // 常に "2.0"
	ID      interface{} `json:"id"`      // 対応するリクエストID
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Error は JSON-RPC 2.0 エラー
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCP プロトコル固有の型定義

// InitializeParams は initialize リクエストのパラメータ
type InitializeParams struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    Capabilities `json:"capabilities"`
	ClientInfo      ClientInfo   `json:"clientInfo"`
}

type Capabilities struct {
	Roots    *RootsCapability    `json:"roots,omitempty"`
	Sampling *SamplingCapability `json:"sampling,omitempty"`
}

type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type SamplingCapability struct{}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult は initialize レスポンスの結果
type InitializeResult struct {
	ProtocolVersion string           `json:"protocolVersion"`
	Capabilities    ServerCapability `json:"capabilities"`
	ServerInfo      ServerInfo       `json:"serverInfo"`
}

type ServerCapability struct {
	Tools *ToolsCapability `json:"tools,omitempty"`
}

type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Tool は MCP ツールの定義
// AI が呼び出せる機能を表します
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

// ListToolsResult は tools/list の結果
type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}

// CallToolParams は tools/call リクエストのパラメータ
type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// CallToolResult は tools/call の結果
type CallToolResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

type Content struct {
	Type string `json:"type"` // "text" など
	Text string `json:"text"`
}
