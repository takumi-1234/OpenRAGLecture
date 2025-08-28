# Dockerfile

# --- Stage 1: Production Builder (本番用バイナリのビルド) ---
# Goのソースコードをコンパイルするためのステージ
FROM golang:1.25-alpine AS production_builder

# build-base と git をインストール
RUN apk add --no-cache build-base git

# アプリケーションのワーキングディレクトリを設定
WORKDIR /app

# go.mod と go.sum を先にコピーして、依存関係をキャッシュしやすくする
COPY go.mod go.sum ./

# 依存関係をダウンロード
RUN go mod download

# アプリケーションのソースコードをコピー
COPY . .

# アプリケーションをビルド (本番用)
# CGO_ENABLED=0 は静的リンクバイナリを生成し、
# alpineベースの最終イメージで外部ライブラリ依存をなくすために重要です。
# -o /app/main で出力先を指定します。
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/api

# --- Stage 2: Debug Environment (開発用環境、ホットリロード対応) ---
# 'air' を使ったホットリロード開発環境のためのステージ
# docker-compose.yml で `target: debug` を指定された場合に利用されます。
FROM golang:1.25-alpine AS debug

# git をインストール (go mod download や air の内部操作で必要となる場合があるため)
RUN apk add --no-cache git

# アプリケーションのワーキングディレクトリを設定
WORKDIR /app

# go.mod と go.sum をコピーして依存関係をダウンロード
# このステージではソースコードは `docker-compose.yml` の volumes でマウントされるため、
# ここで COPY . . は行いません。
COPY go.mod go.sum ./
RUN go mod download

# ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★
# 修正点:
# ホットリロードツール 'air' を新しい公式パスからインストールします。
# 旧: github.com/cosmtrek/air
# 新: github.com/air-verse/air
# ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★
RUN go install github.com/air-verse/air@latest

# デフォルトコマンドとして 'air' を実行
# air がアプリケーションのビルドと実行、そしてホットリロードを管理します。
# 開発中は `docker-compose.yml` の volumes 設定により、ホストのソースコードが /app にマウントされます。
CMD ["air"]


# --- Stage 3: Production Final (本番用実行イメージ) ---
# ビルドされたバイナリを実行するための、軽量な本番用ステージ
FROM alpine:latest AS production_final

# タイムゾーンデータとSSL証明書をインストール
RUN apk add --no-cache tzdata ca-certificates

# アプリケーションのワーキングディレクトリを設定
WORKDIR /app

# ビルダーステージからコンパイル済みのバイナリのみをコピー
COPY --from=production_builder /app/main .

# 設定ファイルをコピー
# アプリケーションは実行時にこれらのファイルを読み込みます
COPY configs/ /app/configs/

# (オプション) DBマイグレーションファイルをコピー
# アプリケーション起動時にマイグレーションを実行する場合に必要
COPY db/ /app/db/

# アプリケーションがリッスンするポートを公開
EXPOSE 8180

# コンテナ起動時に実行されるコマンド
# /app/main を実行します
CMD ["/app/main"]