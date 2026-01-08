package main

import (
	"log"
	"os"

	"github.com/shshimamo/redash-mcp-go/mcp"
	"github.com/shshimamo/redash-mcp-go/redash"
	"github.com/shshimamo/redash-mcp-go/tools"
)

func main() {
	// ログを stderr に出力（stdout は MCP 通信に使用）
	log.SetOutput(os.Stderr)

	// 環境変数から設定を取得
	redashURL := os.Getenv("REDASH_URL")
	redashAPIKey := os.Getenv("REDASH_API_KEY")

	if redashURL == "" || redashAPIKey == "" {
		log.Fatal("REDASH_URL and REDASH_API_KEY environment variables are required")
	}

	log.Printf("Connecting to Redash at: %s", redashURL)

	// Redash クライアントを作成
	redashClient := redash.NewClient(redashURL, redashAPIKey)

	// ツールハンドラーを作成
	toolHandler := tools.NewHandler(redashClient)

	// MCP サーバーを作成
	server := mcp.NewServer(toolHandler)

	// サーバーを起動（stdin/stdout で通信）
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
