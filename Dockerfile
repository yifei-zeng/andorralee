# 多阶段构建 - 使用官方Go镜像作为构建环境
# 支持多架构构建 (x86_64/amd64 和 ARM64/aarch64)
FROM --platform=$BUILDPLATFORM golang:1.23.4-alpine AS builder

# 设置构建参数
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

# 设置工作目录
WORKDIR /app

# 安装构建依赖
RUN apk --no-cache add git ca-certificates tzdata

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用 - 支持交叉编译
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -a -installsuffix cgo -ldflags '-extldflags "-static"' \
    -o andorralee ./cmd/main.go

# 运行时镜像 - 使用更兼容的基础镜像
# Alpine Linux 在银河麒麟系统上有更好的兼容性
FROM alpine:3.19

# 安装运行时依赖
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    curl \
    bash \
    && rm -rf /var/cache/apk/*

# 设置时区为中国时区
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 创建非root用户以提高安全性
RUN addgroup -g 1000 appgroup && \
    adduser -D -u 1000 -G appgroup appuser

# 创建工作目录
WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/andorralee /app/
RUN chmod +x /app/andorralee

# 复制必要的配置文件和静态资源
COPY --from=builder /app/static /app/static
COPY --from=builder /app/scripts /app/scripts

# 创建必要的目录并设置权限
RUN mkdir -p /app/dm_home /app/logs /app/data && \
    chown -R appuser:appgroup /app

# 暴露应用端口
EXPOSE 8081

# 设置环境变量 - 针对银河麒麟系统优化
ENV MYSQL_HOST=mysql \
    MYSQL_PORT=3306 \
    MYSQL_USER=root \
    MYSQL_PASSWORD=123456 \
    MYSQL_DATABASE=andorralee \
    DAMENG_HOST=dameng \
    DAMENG_PORT=5236 \
    DAMENG_USER=SYSDBA \
    DAMENG_PASSWORD=Dm123456 \
    DAMENG_DATABASE=DOCKER_OPS \
    DM_HOME=/app/dm_home \
    GIN_MODE=release \
    GOOS=linux

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8081/health || exit 1

# 切换到非root用户
USER appuser

# 启动应用
CMD ["/app/andorralee"]