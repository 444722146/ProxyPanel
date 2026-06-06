# ProxyPanel 后端Docker镜像
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /build

# 安装必要的工具
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# 复制依赖文件
COPY backend/go.mod backend/go.sum ./

#下载依赖
RUN go mod download

# 复制源代码
COPY backend/ ./

# 编译应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o proxypanel .

# 运行镜像
FROM alpine:latest

# 安装必要的运行依赖
RUN apk add --no-cache \
    nginx \
    sqlite \
    ca-certificates \
    tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建必要的目录
RUN mkdir -p \
    /app/data \
    /app/templates \
    /etc/nginx/sites-enabled \
    /var/log/nginx \
    /run/nginx

# 复制编译好的应用
COPY --from=builder /build/proxypanel /app/
COPY --from=builder /build/templates/nginx.conf.tmpl /app/templates/

# 复制Nginx配置
COPY nginx/nginx.conf /etc/nginx/nginx.conf

# 设置工作目录
WORKDIR /app

# 暴露端口
EXPOSE 5000 80 443

# 启动脚本
COPY scripts/start.sh /app/
RUN chmod +x /app/start.sh

# 启动服务
CMD ["/app/start.sh"]