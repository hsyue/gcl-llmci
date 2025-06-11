.PHONY: build test clean plugin custom-lint install-deps lint

# 变量定义
GO_VERSION := 1.19
PLUGIN_NAME := llmci
PLUGIN_FILE := $(PLUGIN_NAME).so
CUSTOM_LINT := ./bin/custom-golangci-lint

# 默认目标
all: build test

# 安装依赖
install-deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# 构建插件
build: install-deps
	@echo "Building plugin..."
	go build -v

# 构建Go插件
plugin: install-deps
	@echo "Building Go plugin..."
	go build -buildmode=plugin -o $(PLUGIN_FILE) .

# 构建自定义golangci-lint（需要先安装golangci-lint）
custom-lint:
	@echo "Building custom golangci-lint..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Error: golangci-lint is not installed. Please install it first."; \
		exit 1; \
	fi
	mkdir -p bin
	golangci-lint custom

# 运行测试
test: install-deps
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out

# 运行基准测试
bench: install-deps
	@echo "Running benchmarks..."
	go test -bench=. -benchmem

# 代码覆盖率报告
coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 代码检查
lint: install-deps
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping lint check"; \
	fi

# 格式化代码
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# 清理构建文件
clean:
	@echo "Cleaning up..."
	rm -f $(PLUGIN_FILE)
	rm -f coverage.out coverage.html
	rm -rf bin/
	go clean

# 验证模块
verify:
	@echo "Verifying module..."
	go mod verify

# 安装到GOPATH
install: build
	@echo "Installing to GOPATH..."
	go install

# 创建发布包
package: clean build test
	@echo "Creating release package..."
	mkdir -p dist
	tar -czf dist/$(PLUGIN_NAME)-$(shell git describe --tags --always).tar.gz \
		--exclude='.git' \
		--exclude='dist' \
		--exclude='*.so' \
		--exclude='coverage.*' \
		.

# 显示帮助信息
help:
	@echo "Available targets:"
	@echo "  build        - Build the plugin"
	@echo "  plugin       - Build as Go plugin (.so file)"
	@echo "  custom-lint  - Build custom golangci-lint with plugin"
	@echo "  test         - Run tests"
	@echo "  bench        - Run benchmarks"
	@echo "  coverage     - Generate coverage report"
	@echo "  lint         - Run code linters"
	@echo "  fmt          - Format code"
	@echo "  clean        - Clean build files"
	@echo "  install-deps - Install dependencies"
	@echo "  verify       - Verify module"
	@echo "  install      - Install to GOPATH"
	@echo "  package      - Create release package"
	@echo "  help         - Show this help message"

# 开发环境设置
dev-setup: install-deps
	@echo "Setting up development environment..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin; \
	fi
	@if ! command -v goimports >/dev/null 2>&1; then \
		echo "Installing goimports..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
	fi
	@echo "Development environment ready!"

# 运行示例测试
example: plugin
	@echo "Running example analysis..."
	@if [ -f $(PLUGIN_FILE) ]; then \
		echo "Plugin built successfully: $(PLUGIN_FILE)"; \
		echo "To use the plugin, configure your .golangci.yml file"; \
		echo "and set your API token in the configuration."; \
	else \
		echo "Plugin build failed"; \
		exit 1; \
	fi