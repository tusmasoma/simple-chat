# ベースイメージ
FROM golang:1.21.3

# 作業ディレクトリの設定
WORKDIR /app

# 依存関係をコピー
COPY go.mod ./
#COPY go.sum ./

# 依存関係のインストール
WORKDIR /app
RUN go mod download

# Air をインストール
RUN go install github.com/cosmtrek/air@v1.29.0

WORKDIR /app

# ソースコードをコピー
COPY . .

# Air を使用してアプリケーションを起動する
WORKDIR /app
CMD ["air", "-c", ".air.toml"]