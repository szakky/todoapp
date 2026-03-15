# ビルド用のステージ
FROM golang:1.24.4-alpine AS builder

WORKDIR /app

# 依存関係のコピーとダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードのコピーとビルド
COPY . .
RUN go build -o todo-api main.go

# 実行用のステージ (マルチステージビルドで軽量化)
FROM alpine:latest

WORKDIR /app

# ビルドしたバイナリをコピー
COPY --from=builder /app/todo-api .

# ポートのエクスポート
EXPOSE 8080

# アプリケーションの実行
CMD ["./todo-api"]