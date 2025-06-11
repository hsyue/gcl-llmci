# golangci-lint LLM 插件

这是一个 golangci-lint 插件，能够将指定的 Go 代码文件发送给 LLM（大语言模型）进行分析，提供代码质量建议、潜在问题检测和最佳实践推荐。

## 功能特性

- 🎯 **文件模式匹配**：支持 glob 模式和正则表达式来指定要分析的文件
- 🤖 **LLM 集成**：支持 OpenAI API 和其他兼容的 API 服务
- ⚙️ **灵活配置**：可配置提示词、API 地址、超时时间等
- 🚀 **高性能**：基于 go/analysis 框架，支持并行分析
- 🔧 **易于集成**：支持 golangci-lint 的 Module Plugin System

## 安装和使用

### 方法一：使用 Module Plugin System（推荐）

1. 创建 `.custom-gcl.yml` 配置文件：

```yaml
version: v1.55.0
name: custom-golangci-lint
destination: ./bin/
plugins:
  - module: 'github.com/golangci/llmci'
    path: ./
```

2. 构建自定义的 golangci-lint：

```bash
golangci-lint custom
```

### 方法二：使用 Go Plugin System

1. 构建插件：

```bash
go build -buildmode=plugin -o llmci.so .
```

2. 在 `.golangci.yml` 中配置：

```yaml
version: "2"

linters-settings:
  custom:
    llmci:
      path: ./llmci.so
      description: "使用LLM分析Go代码文件"
      original-url: github.com/golangci/llmci
      settings:
        file-patterns: "*.go,*_test.go"
        prompt: "请分析这个Go代码文件，重点关注：1. 代码质量和最佳实践 2. 潜在的bug和安全问题 3. 性能优化建议 4. 代码可读性和维护性"
        api-url: "https://api.openai.com/v1/chat/completions"
        api-token: "your-openai-api-token-here"
        timeout: 30
        enabled: true

linters:
  enable:
    - llmci
```

## 配置选项

| 选项 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `file-patterns` | string | `"*.go"` | 文件匹配模式，支持多个模式用逗号分隔 |
| `prompt` | string | 默认提示词 | 发送给 LLM 的系统提示词 |
| `api-url` | string | `"https://api.openai.com/v1/chat/completions"` | LLM API 地址 |
| `api-token` | string | `""` | API 访问令牌（必需） |
| `timeout` | int | `30` | API 请求超时时间（秒） |
| `enabled` | bool | `true` | 是否启用插件 |

## 文件模式匹配

支持以下模式：

- `*.go` - 匹配所有 .go 文件
- `*_test.go` - 匹配所有测试文件
- `main.go` - 精确匹配
- `src/**/*.go` - 递归匹配（需要完整路径）
- `test?.go` - 单字符通配符

多个模式可以用逗号分隔：`"*.go,*_test.go,cmd/**/*.go"`

## API 兼容性

插件设计为与 OpenAI Chat Completions API 兼容，也支持其他兼容的 API 服务，如：

- OpenAI GPT-3.5/GPT-4
- Azure OpenAI Service
- 本地部署的兼容服务
- 其他支持相同 API 格式的服务

## 示例输出

```
example.go:1:1: LLM Analysis for example.go:
这个Go代码文件存在以下问题和改进建议：

1. **错误处理缺失**：BadFunction中忽略了os.Open的错误，这可能导致程序崩溃
2. **未使用变量**：unusedVar变量被声明但未使用
3. **硬编码值**：循环中的100应该作为常量或参数
4. **空指针解引用**：ptr变量可能为nil，直接解引用会导致panic

建议：
- 始终检查和处理错误
- 移除未使用的变量
- 使用常量替代魔法数字
- 在解引用指针前检查是否为nil
```

## 开发和测试

### 运行测试

```bash
go test -v
```

### 运行基准测试

```bash
go test -bench=.
```

### 构建

```bash
go build
```

## 项目结构

```
.
├── llmci.go              # 主要插件代码
├── plugin.go             # 插件入口
├── llmci_test.go         # 测试文件
├── testdata/             # 测试数据
│   └── example.go        # 示例测试文件
├── go.mod                # Go 模块文件
├── .golangci.yml         # golangci-lint 配置示例
├── .custom-gcl.yml       # 自定义构建配置
└── README.md             # 项目文档
```

## 注意事项

1. **API 密钥安全**：请确保不要将 API 密钥提交到版本控制系统中
2. **网络依赖**：插件需要网络连接来访问 LLM API
3. **成本控制**：LLM API 调用可能产生费用，建议合理配置文件模式和超时时间
4. **隐私考虑**：代码内容会发送到外部 API，请确保符合您的隐私和安全要求

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License