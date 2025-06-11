package main

import (
	"go/parser"
	"go/token"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestGlobToRegex 测试glob模式转换为正则表达式
func TestGlobToRegex(t *testing.T) {
	tests := []struct {
		name     string
		glob     string
		expected string
	}{
		{
			name:     "simple wildcard",
			glob:     "*.go",
			expected: "^.*\\.go$",
		},
		{
			name:     "question mark",
			glob:     "test?.go",
			expected: "^test.\\.go$",
		},
		{
			name:     "exact match",
			glob:     "main.go",
			expected: "^main\\.go$",
		},
		{
			name:     "multiple wildcards",
			glob:     "*_test*.go",
			expected: "^.*_test.*\\.go$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := globToRegex(tt.glob)
			if result != tt.expected {
				t.Errorf("globToRegex(%q) = %q, want %q", tt.glob, result, tt.expected)
			}
		})
	}
}

// TestGetFileContent 测试获取文件内容功能
func TestGetFileContent(t *testing.T) {
	src := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	content, err := getFileContent(fset, file)
	if err != nil {
		t.Fatalf("getFileContent failed: %v", err)
	}

	if !strings.Contains(content, "package main") {
		t.Errorf("Expected content to contain 'package main', got: %s", content)
	}
	if !strings.Contains(content, "fmt.Println") {
		t.Errorf("Expected content to contain 'fmt.Println', got: %s", content)
	}
}

// TestAnalyzeWithLLM 测试LLM分析功能（使用mock服务器）
func TestAnalyzeWithLLM(t *testing.T) {
	// 创建mock HTTP服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Errorf("Expected Authorization header with Bearer token")
		}

		// 返回mock响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"choices": [{
				"message": {
					"role": "assistant",
					"content": "这是一个简单的Go程序，代码质量良好。建议添加错误处理。"
				}
			}]
		}`))
	}))
	defer server.Close()

	// 设置测试参数
	originalAPIURL := *apiURL
	originalAPIToken := *apiToken
	defer func() {
		*apiURL = originalAPIURL
		*apiToken = originalAPIToken
	}()

	*apiURL = server.URL
	*apiToken = "test-token"

	// 测试分析功能
	content := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`
	filename := "test.go"

	result, err := analyzeWithLLM(content, filename)
	if err != nil {
		t.Fatalf("analyzeWithLLM failed: %v", err)
	}

	expected := "这是一个简单的Go程序，代码质量良好。建议添加错误处理。"
	if result != expected {
		t.Errorf("Expected result %q, got %q", expected, result)
	}
}

// TestAnalyzeWithLLMError 测试LLM分析错误处理
func TestAnalyzeWithLLMError(t *testing.T) {
	// 创建返回错误的mock服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{
			"error": {
				"message": "Invalid API key",
				"type": "invalid_request_error"
			}
		}`))
	}))
	defer server.Close()

	// 设置测试参数
	originalAPIURL := *apiURL
	originalAPIToken := *apiToken
	defer func() {
		*apiURL = originalAPIURL
		*apiToken = originalAPIToken
	}()

	*apiURL = server.URL
	*apiToken = "invalid-token"

	// 测试错误处理
	content := `package main`
	filename := "test.go"

	_, err := analyzeWithLLM(content, filename)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "Invalid API key") {
		t.Errorf("Expected error to contain 'Invalid API key', got: %v", err)
	}
}

// TestAnalyzerBasics 测试分析器基本功能
func TestAnalyzerBasics(t *testing.T) {
	// 测试分析器基本属性
	if Analyzer.Name != "llmci" {
		t.Errorf("Expected analyzer name 'llmci', got %q", Analyzer.Name)
	}

	if Analyzer.Doc == "" {
		t.Error("Analyzer should have documentation")
	}

	// 测试标志设置
	if enabled == nil {
		t.Error("enabled flag should be initialized")
	}

	if apiToken == nil {
		t.Error("apiToken flag should be initialized")
	}

	if filePatterns == nil {
		t.Error("filePatterns flag should be initialized")
	}

	// 测试默认值
	if *filePatterns == "" {
		t.Error("filePatterns should have a default value")
	}

	if *apiURL == "" {
		t.Error("apiURL should have a default value")
	}
}

// TestAnalyzerIntegration 集成测试
func TestAnalyzerIntegration(t *testing.T) {
	// 集成测试 - 这里可以添加更复杂的集成测试
	// 例如测试整个分析流程
	if Analyzer.Name != "llmci" {
		t.Errorf("Expected analyzer name 'llmci', got %q", Analyzer.Name)
	}

	if Analyzer.Doc == "" {
		t.Error("Analyzer should have documentation")
	}

	if Analyzer.Run == nil {
		t.Error("Analyzer should have a Run function")
	}

	// 测试分析器是否正确注册了标志
	flags := Analyzer.Flags
	if flags.Lookup("enabled") == nil {
		t.Error("Analyzer should register 'enabled' flag")
	}

	if flags.Lookup("api-token") == nil {
		t.Error("Analyzer should register 'api-token' flag")
	}

	if flags.Lookup("file-patterns") == nil {
		t.Error("Analyzer should register 'file-patterns' flag")
	}

	// 测试分析器的依赖
	if len(Analyzer.Requires) > 0 {
		t.Logf("Analyzer requires: %v", Analyzer.Requires)
	}

	// 测试分析器的结果类型
	if Analyzer.ResultType != nil {
		t.Logf("Analyzer result type: %v", Analyzer.ResultType)
	}
}

// BenchmarkGlobToRegex 性能测试
func BenchmarkGlobToRegex(b *testing.B) {
	patterns := []string{"*.go", "*_test.go", "test?.go", "**/*.go"}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, pattern := range patterns {
			globToRegex(pattern)
		}
	}
}

// BenchmarkGetFileContent 性能测试
func BenchmarkGetFileContent(b *testing.B) {
	src := `package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: program <arg>")
		return
	}
	fmt.Printf("Hello, %s!\n", os.Args[1])
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "bench.go", src, parser.ParseComments)
	if err != nil {
		b.Fatalf("Failed to parse file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := getFileContent(fset, file)
		if err != nil {
			b.Fatalf("getFileContent failed: %v", err)
		}
	}
}