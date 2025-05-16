# 建置階段：使用官方 Golang 映像檔作為建置環境
FROM golang:1.24-alpine AS builder

# 設定工作目錄
WORKDIR /app

# 複製 go.mod 和 go.sum 以便下載依賴
COPY go.mod go.sum ./
RUN go mod download

# 複製所有原始碼
COPY . .

# 編譯程式，輸出成 /app/main（可自行命名）
RUN go build -o main .

# 執行階段：使用輕量的 Alpine Linux 映像
FROM alpine:latest

LABEL author="Fatorin"

# 設定工作目錄
WORKDIR /root/

# 複製編譯好的執行檔
COPY --from=builder /app/main .

# 複製 templates 資料夾
COPY --from=builder /app/templates ./templates

# 容器啟動時執行的指令
CMD ["./main"]
