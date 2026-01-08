package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

// Server は MCP サーバー
type Server struct {
	toolHandler ToolHandler
	reader      *bufio.Reader
	writer      io.Writer
}

// ToolHandler はツールの実行を担当するインターフェース
type ToolHandler interface {
	GetTools() []Tool
	CallTool(name string, arguments map[string]interface{}) CallToolResult
}

// NewServer は新しい MCP サーバーを作成
func NewServer(toolHandler ToolHandler) *Server {
	return &Server{
		toolHandler: toolHandler,
		reader:      bufio.NewReader(os.Stdin),
		writer:      os.Stdout,
	}
}

// Start はサーバーを起動して stdin からリクエストを受け取る
func (s *Server) Start() error {
	log.Println("Redash MCP Server starting...")

	for {
		// 1行読み込み (JSON-RPC リクエスト)
		line, err := s.reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("Client disconnected")
				return nil
			}
			return fmt.Errorf("failed to read request: %w", err)
		}

		// 空行は無視
		if len(line) == 0 || (len(line) == 1 && line[0] == '\n') {
			continue
		}

		log.Printf("Received request: %s", string(line))

		// JSON-RPC リクエストをパース
		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			s.sendError(nil, -32700, "Parse error", err.Error())
			continue
		}

		// リクエストを処理
		s.handleRequest(&req)
	}
}

// handleRequest はリクエストを処理してレスポンスを返す
func (s *Server) handleRequest(req *Request) {
	switch req.Method {
	case "initialize":
		s.handleInitialize(req)
	case "initialized":
		// initialized 通知は応答不要
		log.Println("Client initialized")
	case "tools/list":
		s.handleListTools(req)
	case "tools/call":
		s.handleCallTool(req)
	case "ping":
		s.sendResponse(req.ID, map[string]interface{}{})
	default:
		s.sendError(req.ID, -32601, "Method not found", fmt.Sprintf("Unknown method: %s", req.Method))
	}
}

// handleInitialize は initialize リクエストを処理
func (s *Server) handleInitialize(req *Request) {
	var params InitializeParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, -32602, "Invalid params", err.Error())
		return
	}

	log.Printf("Initialize from client: %s %s", params.ClientInfo.Name, params.ClientInfo.Version)

	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: ServerCapability{
			Tools: &ToolsCapability{
				ListChanged: false,
			},
		},
		ServerInfo: ServerInfo{
			Name:    "redash-mcp-server",
			Version: "1.0.0",
		},
	}

	s.sendResponse(req.ID, result)
}

// handleListTools は tools/list リクエストを処理
func (s *Server) handleListTools(req *Request) {
	tools := s.toolHandler.GetTools()

	result := ListToolsResult{
		Tools: tools,
	}

	s.sendResponse(req.ID, result)
}

// handleCallTool は tools/call リクエストを処理
func (s *Server) handleCallTool(req *Request) {
	var params CallToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, -32602, "Invalid params", err.Error())
		return
	}

	log.Printf("Calling tool: %s with args: %v", params.Name, params.Arguments)

	result := s.toolHandler.CallTool(params.Name, params.Arguments)

	s.sendResponse(req.ID, result)
}

// sendResponse はレスポンスを送信
func (s *Server) sendResponse(id interface{}, result interface{}) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		return
	}

	log.Printf("Sending response: %s", string(data))

	// stdout に書き込み（改行を追加）
	if _, err := fmt.Fprintf(s.writer, "%s\n", data); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

// sendError はエラーレスポンスを送信
func (s *Server) sendError(id interface{}, code int, message string, data interface{}) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	respData, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Failed to marshal error response: %v", err)
		return
	}

	log.Printf("Sending error: %s", string(respData))

	if _, err := fmt.Fprintf(s.writer, "%s\n", respData); err != nil {
		log.Printf("Failed to write error response: %v", err)
	}
}
