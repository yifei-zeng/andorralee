# Andorralee蜜罐诱捕系统 Makefile

# 版本信息
VERSION := v2.0
BUILD_TIME := $(shell date +%Y%m%d%H%M%S)
GIT_COMMIT := $(shell git rev-parse --short HEAD)

# 构建配置
BINARY_NAME := andorralee
BUILD_DIR := ./build
DOCKER_IMAGE := andorralee:$(VERSION)

# Go配置
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# 构建标志
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

.PHONY: all build clean test deps docker run help

# 默认目标
all: deps test build

# 显示帮助信息
help:
	@echo "Andorralee蜜罐诱捕系统构建工具"
	@echo ""
	@echo "可用命令:"
	@echo "  build      - 构建应用程序"
	@echo "  clean      - 清理构建文件"
	@echo "  test       - 运行测试"
	@echo "  deps       - 下载依赖"
	@echo "  docker     - 构建Docker镜像"
	@echo "  run        - 运行应用程序"
	@echo "  dev        - 开发模式运行"
	@echo "  lint       - 代码检查"
	@echo "  deploy     - 部署到生产环境"
	@echo "  help       - 显示此帮助信息"

# 下载依赖
deps:
	@echo "正在下载依赖..."
	$(GOMOD) download
	$(GOMOD) tidy

# 构建应用程序
build: deps
	@echo "正在构建应用程序..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/main.go
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

# 构建Windows版本
build-windows: deps
	@echo "正在构建Windows版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).exe ./cmd/main.go
	@echo "Windows版本构建完成: $(BUILD_DIR)/$(BINARY_NAME).exe"

# 构建Linux版本
build-linux: deps
	@echo "正在构建Linux版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux ./cmd/main.go
	@echo "Linux版本构建完成: $(BUILD_DIR)/$(BINARY_NAME)-linux"

# 交叉编译所有平台
build-all: build-windows build-linux build

# 清理构建文件
clean:
	@echo "正在清理构建文件..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	@echo "清理完成"

# 运行测试
test:
	@echo "正在运行测试..."
	$(GOTEST) -v ./...

# 运行测试并生成覆盖率报告
test-coverage:
	@echo "正在运行测试并生成覆盖率报告..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

# 代码检查
lint:
	@echo "正在进行代码检查..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "正在安装golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	golangci-lint run

# 运行应用程序
run: build
	@echo "正在启动应用程序..."
	$(BUILD_DIR)/$(BINARY_NAME)

# 开发模式运行
dev:
	@echo "开发模式启动..."
	$(GOCMD) run ./cmd/main.go

# 构建Docker镜像
docker:
	@echo "正在构建Docker镜像..."
	docker build -t $(DOCKER_IMAGE) .
	@echo "Docker镜像构建完成: $(DOCKER_IMAGE)"

# 构建并运行Docker容器
docker-run: docker
	@echo "正在启动Docker容器..."
	docker-compose up -d

# 停止Docker容器
docker-stop:
	@echo "正在停止Docker容器..."
	docker-compose down

# 查看Docker日志
docker-logs:
	@echo "查看Docker容器日志..."
	docker-compose logs -f

# 数据库初始化
db-init:
	@echo "正在初始化数据库..."
	@if [ -f "deployment/create_tables.sql" ]; then \
		mysql -u root -p < deployment/create_tables.sql; \
		echo "数据库初始化完成"; \
	else \
		echo "错误: 找不到数据库初始化脚本"; \
	fi

# 数据库迁移
db-migrate:
	@echo "正在执行数据库迁移..."
	@if [ -f "db_migration.sql" ]; then \
		mysql -u root -p < db_migration.sql; \
		echo "数据库迁移完成"; \
	else \
		echo "错误: 找不到数据库迁移脚本"; \
	fi

# 生成API文档
docs:
	@echo "正在生成API文档..."
	@if ! command -v swag >/dev/null 2>&1; then \
		echo "正在安装swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	swag init -g cmd/main.go -o docs/swagger

# 安装前端依赖
frontend-deps:
	@echo "正在安装前端依赖..."
	@if [ -d "frontend/node_modules" ]; then \
		cd frontend && npm install; \
	else \
		echo "警告: 前端项目尚未初始化"; \
	fi

# 构建前端
frontend-build: frontend-deps
	@echo "正在构建前端..."
	@if [ -d "frontend/src" ]; then \
		cd frontend && npm run build; \
	else \
		echo "警告: 前端项目尚未初始化"; \
	fi

# 开发模式启动前端
frontend-dev:
	@echo "前端开发模式启动..."
	@if [ -d "frontend/src" ]; then \
		cd frontend && npm start; \
	else \
		echo "警告: 前端项目尚未初始化"; \
	fi

# 部署到生产环境
deploy: build-linux docker
	@echo "正在部署到生产环境..."
	@echo "1. 构建完成"
	@echo "2. Docker镜像已准备"
	@echo "3. 请在目标环境执行 docker-compose 命令:"
	@echo "   docker-compose -f docker-compose.yml up -d"
	@echo "   # 如需差异化配置，可自定义 override 文件"

# 快速部署 (开发环境)
deploy-dev: docker-run
	@echo "开发环境部署完成"
	@echo "应用程序访问地址: http://localhost:8848"

# 健康检查
health-check:
	@echo "正在执行健康检查..."
	@curl -f http://localhost:8848/health || echo "服务不可用"

# 查看服务状态
status:
	@echo "=== 服务状态 ==="
	@echo "Go应用进程:"
	@ps aux | grep $(BINARY_NAME) || echo "未找到Go应用进程"
	@echo ""
	@echo "Docker容器:"
	@docker ps | grep andorralee || echo "未找到Docker容器"
	@echo ""
	@echo "端口占用:"
	@netstat -tlnp | grep :8848 || echo "8848端口未被占用"

# 日志查看
logs:
	@echo "查看应用日志..."
	@if [ -f "logs/andorralee.log" ]; then \
		tail -f logs/andorralee.log; \
	else \
		echo "日志文件不存在"; \
	fi

# 性能测试
benchmark:
	@echo "正在执行性能测试..."
	$(GOTEST) -bench=. -benchmem ./...

# 安全扫描
security-scan:
	@echo "正在执行安全扫描..."
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "正在安装gosec..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	gosec ./...

# 完整的CI检查
ci: deps lint test security-scan build
	@echo "=== CI检查完成 ==="
	@echo "✓ 依赖检查"
	@echo "✓ 代码规范"
	@echo "✓ 单元测试"
	@echo "✓ 安全扫描"
	@echo "✓ 构建成功"

# 发布版本
release: ci build-all docker
	@echo "=== 版本发布 $(VERSION) ==="
	@echo "✓ 多平台构建完成"
	@echo "✓ Docker镜像已构建"
	@echo "✓ 准备发布版本: $(VERSION)"

# 清理所有资源
purge: clean docker-stop
	@echo "正在清理所有资源..."
	@docker rmi $(DOCKER_IMAGE) 2>/dev/null || true
	@docker system prune -f
	@echo "资源清理完成"
