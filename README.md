# Redash MCP Server (Go)

Model Context Protocol (MCP) server for integrating Redash with AI assistants like Claude.

これは [suthio/redash-mcp](https://github.com/suthio/redash-mcp) の TypeScript 版を Go 言語で再実装したものです。

## Features

このサーバーは以下の MCP ツールを提供します：

- **execute_query** - 保存済みクエリをIDで実行
  - パラメータ付きクエリにも対応
  - クエリ結果を JSON 形式で返す

- **execute_adhoc_query** - SQL を直接実行
  - データソースIDと SQL を指定
  - 一時的なクエリ実行に便利

## Requirements

- Go 1.25 or later
- Redash インスタンスと API キー

## Installation

### 方法1: ビルド済みバイナリをダウンロード（推奨）

[Releases ページ](https://github.com/shshimamo/redash-mcp-go/releases) から最新版をダウンロードしてください。

#### macOS

```bash
# Apple Silicon (M1/M2/M3)
curl -L https://github.com/shshimamo/redash-mcp-go/releases/latest/download/redash-mcp-go-darwin-arm64 -o redash-mcp-go

# Intel Mac
curl -L https://github.com/shshimamo/redash-mcp-go/releases/latest/download/redash-mcp-go-darwin-amd64 -o redash-mcp-go

# 実行権限を付与して配置
chmod +x redash-mcp-go
sudo mv redash-mcp-go /usr/local/bin/
```

#### Linux

```bash
# amd64
curl -L https://github.com/shshimamo/redash-mcp-go/releases/latest/download/redash-mcp-go-linux-amd64 -o redash-mcp-go

# arm64
curl -L https://github.com/shshimamo/redash-mcp-go/releases/latest/download/redash-mcp-go-linux-arm64 -o redash-mcp-go

# 実行権限を付与して配置
chmod +x redash-mcp-go
sudo mv redash-mcp-go /usr/local/bin/
```

#### Windows

PowerShell で実行：

```powershell
Invoke-WebRequest -Uri "https://github.com/shshimamo/redash-mcp-go/releases/latest/download/redash-mcp-go-windows-amd64.exe" -OutFile "redash-mcp-go.exe"

# PATH が通っているディレクトリに配置
Move-Item redash-mcp-go.exe C:\Windows\System32\
```

### 方法2: ソースからビルド（開発者向け）

```bash
# リポジトリをクローン
git clone https://github.com/shshimamo/redash-mcp-go.git
cd redash-mcp-go

# ビルド
make build

# または直接 go build
go build -o redash-mcp-go .
```

## Usage

### Claude Code (CLI) - プロジェクトローカル設定

プロジェクトディレクトリで使用する場合（推奨）：

1. `.mcp.json.example` をコピー：
```bash
cp .mcp.json.example .mcp.json
```

2. `.mcp.json` を編集して、Redash の情報を設定：
```json
{
  "mcpServers": {
    "redash": {
      "type": "stdio",
      "command": "/usr/local/bin/redash-mcp-go",
      "env": {
        "REDASH_URL": "https://your-redash-instance.com",
        "REDASH_API_KEY": "your-api-key"
      }
    }
  }
}
```

3. このディレクトリで `claude` コマンドを実行すると、自動的に MCP サーバーが読み込まれます

または、コマンドで追加：
```bash
claude mcp add --transport stdio redash --scope project \
  --env REDASH_URL=https://your-redash-instance.com \
  --env REDASH_API_KEY=your-api-key \
  -- /usr/local/bin/redash-mcp-go
```

### Claude Desktop - グローバル設定

#### macOS

`~/Library/Application Support/Claude/claude_desktop_config.json` を編集：

```json
{
  "mcpServers": {
    "redash": {
      "command": "/usr/local/bin/redash-mcp-go",
      "env": {
        "REDASH_URL": "https://your-redash-instance.com",
        "REDASH_API_KEY": "your-api-key"
      }
    }
  }
}
```

#### Linux

`~/.config/Claude/claude_desktop_config.json` を編集：

```json
{
  "mcpServers": {
    "redash": {
      "command": "/usr/local/bin/redash-mcp-go",
      "env": {
        "REDASH_URL": "https://your-redash-instance.com",
        "REDASH_API_KEY": "your-api-key"
      }
    }
  }
}
```

#### Windows

`%APPDATA%\Claude\claude_desktop_config.json` を編集：

```json
{
  "mcpServers": {
    "redash": {
      "command": "C:\\Windows\\System32\\redash-mcp-go.exe",
      "env": {
        "REDASH_URL": "https://your-redash-instance.com",
        "REDASH_API_KEY": "your-api-key"
      }
    }
  }
}
```

## Testing

### 手動テスト

```bash
# initialize リクエストをテスト
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | \
  REDASH_URL="https://your-redash.com" REDASH_API_KEY="your-key" ./redash-mcp-server

# tools/list リクエストをテスト
printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}\n{"jsonrpc":"2.0","id":2,"method":"tools/list"}\n' | \
  REDASH_URL="https://your-redash.com" REDASH_API_KEY="your-key" ./redash-mcp-server
```

## Architecture

### 通信方式

このサーバーは **stdio (標準入出力)** で通信します：

- **stdin**: JSON-RPC 2.0 リクエストを受信
- **stdout**: JSON-RPC 2.0 レスポンスを送信
- **stderr**: ログ出力

### フロー

```
Claude Desktop
    ↓ プロセス起動
redash-mcp-server
    ↓ stdin/stdout (JSON-RPC 2.0)
    ↓ MCP プロトコル
    ↓ HTTP リクエスト
Redash API
```

### Directory Structure

```
.
├── main.go              # エントリーポイント
├── mcp/                 # MCP プロトコル実装
│   ├── types.go        # 型定義 (Request, Response, Tool など)
│   └── server.go       # サーバーロジック (stdin/stdout 通信)
├── redash/             # Redash API クライアント
│   └── client.go       # API 呼び出し、ジョブ待機処理
└── tools/              # MCP ツール実装
    └── tools.go        # execute_query, execute_adhoc_query
```

## How It Works

### MCP とは

MCP (Model Context Protocol) は Anthropic が開発した、AI アシスタントが外部システムと通信するための標準プロトコルです。

### 実装の流れ

1. **サーバー起動**: Claude Desktop がサーバープロセスを起動
2. **initialize**: Claude がサーバーに接続し、機能を確認
3. **tools/list**: 利用可能なツールのリストを取得
4. **tools/call**: ツールを実行（例: execute_query）
5. **結果返却**: Redash API を呼び出して結果を返す

### Redash API の非同期処理

Redash のクエリ実行は非同期です：

1. クエリ実行リクエスト → ジョブIDを取得
2. ジョブIDでステータスをポーリング（最大30秒）
3. 完了したら結果を返す

この処理は `redash/client.go` の `waitForJob` 関数で実装されています。

## Development

```bash
# 依存関係の更新
make deps

# ビルド
make build

# クリーン
make clean

# テスト（今後追加予定）
make test
```

## Troubleshooting

### サーバーが起動しない

- 環境変数 `REDASH_URL` と `REDASH_API_KEY` が設定されているか確認
- stderr のログを確認（Claude Desktop のログに出力されます）

### クエリ実行がタイムアウトする

- Redash のクエリが30秒以内に完了するか確認
- ネットワーク接続を確認

### API キーエラー

- Redash の User Settings → API Key で正しいキーを取得
- キーに余分なスペースや改行が含まれていないか確認

## License

MIT
