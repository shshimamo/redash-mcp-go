.PHONY: build run clean test

# ビルド
build:
	go build -o redash-mcp-server .

# ビルドして実行
run: build
	./redash-mcp-server

# クリーンアップ
clean:
	rm -f redash-mcp-server

# テスト
test:
	go test ./...

# 依存関係の取得
deps:
	go mod tidy
	go mod download

# インストール（$GOPATH/bin にコピー）
install:
	go install .
