# 构建阶段
FROM golang:1.24-alpine AS builder
WORKDIR /app

# 将 go.mod 复制到工作目录，并下载依赖
COPY go.mod ./
RUN go mod download

# 将项目源码复制到容器中
COPY . .

# 编译应用程序，生成静态二进制文件
RUN go build -o airnotfy .

# 运行阶段
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

# 从构建阶段复制编译好的二进制文件和模板目录
COPY --from=builder /app/airnotfy .

# 启动应用
CMD ["./airnotfy"]