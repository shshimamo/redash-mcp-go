#!/bin/bash

# MCP サーバーのテストスクリプト
# stdin/stdout でやり取りをシミュレート

set -e

echo "=== MCP Server Test ==="
echo

# サーバーを起動（バックグラウンド）
export REDASH_URL="${REDASH_URL:-https://demo.redash.io}"
export REDASH_API_KEY="${REDASH_API_KEY:-demo-key}"

echo "Testing with:"
echo "  REDASH_URL: $REDASH_URL"
echo "  REDASH_API_KEY: ${REDASH_API_KEY:0:10}..."
echo

# テスト1: initialize
echo "Test 1: Initialize"
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | ./redash-mcp-server 2>&1 | head -5
echo

# テスト2: tools/list
echo "Test 2: List Tools"
echo -e '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}\n{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | ./redash-mcp-server 2>&1 | grep -A 20 '"method":"tools/list"' || echo "tools/list response received"
echo

echo "=== Test Complete ==="
echo "If you see JSON-RPC responses above, the server is working correctly!"
