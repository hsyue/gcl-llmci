package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"golang.org/x/tools/go/analysis"
)

// Config 配置结构
type LLMCiPluginConfig struct {
	FilePatterns []string `json:"file_patterns"`
	Prompt       string   `json:"prompt"`
	APIURL       string   `json:"api_url"`
	APIToken     string   `json:"api_token"`
	Timeout      int      `json:"timeout"`
	Enabled      bool     `json:"enabled"`
	Model        string   `json:"model"`
}

// OpenAIRequest OpenAI API请求结构
type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse OpenAI API响应结构
type OpenAIResponse struct {
	Choices []Choice  `json:"choices"`
	Error   *APIError `json:"error,omitempty"`
}

type Choice struct {
	Message Message `json:"message"`
}

type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

type LLMCiAnalyzer struct {
	config *LLMCiPluginConfig
}

func (a *LLMCiAnalyzer) run(pass *analysis.Pass) (interface{}, error) {

	if a.config.APIToken == "" {
		return nil, fmt.Errorf("API token is required")
	}

	// 编译文件模式正则表达式
	patterns := a.config.FilePatterns
	var regexps []*regexp.Regexp
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		// 将glob模式转换为正则表达式
		regexPattern := globToRegex(pattern)
		regex, err := regexp.Compile(regexPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid file pattern %q: %v", pattern, err)
		}
		regexps = append(regexps, regex)
	}

	if len(regexps) == 0 {
		return nil, fmt.Errorf("no valid file patterns specified")
	}

	// 分析每个文件
	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename
		basename := filepath.Base(filename)

		// 检查文件是否匹配模式
		matched := false
		for _, regex := range regexps {
			if regex.MatchString(basename) || regex.MatchString(filename) {
				matched = true
				break
			}
		}

		if !matched {
			continue
		}

		// 获取文件内容
		content, err := getFileContent(pass.Fset, file)
		if err != nil {
			pass.Reportf(file.Pos(), "failed to get file content: %v", err)
			continue
		}

		// 发送到LLM进行分析
		analysis, err := a.analyzeWithLLM(content, filename)
		if err != nil {
			pass.Reportf(file.Pos(), "LLM analysis failed: %v", err)
			continue
		}

		// 报告分析结果
		if analysis != "" {
			pass.Reportf(file.Pos(), "LLM Analysis for %s:\n%s", basename, analysis)
		}
	}

	return nil, nil
}

// globToRegex 将glob模式转换为正则表达式
func globToRegex(glob string) string {
	// 转义正则表达式特殊字符
	regex := regexp.QuoteMeta(glob)
	// 将glob通配符转换为正则表达式
	regex = strings.ReplaceAll(regex, "\\*", ".*")
	regex = strings.ReplaceAll(regex, "\\?", ".")
	return "^" + regex + "$"
}

// getFileContent 获取文件内容
func getFileContent(fset *token.FileSet, file *ast.File) (string, error) {
	var buf bytes.Buffer
	err := format.Node(&buf, fset, file)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// analyzeWithLLM 使用LLM分析代码
func (a *LLMCiAnalyzer) analyzeWithLLM(content, filename string) (string, error) {
	// 构建请求
	req := OpenAIRequest{
		Model: a.config.Model,
		Messages: []Message{
			{
				Role:    "system",
				Content: a.config.Prompt,
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("文件名: %s\n\n代码内容:\n%s", filename, content),
			},
		},
	}

	// 序列化请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", a.config.APIURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.config.APIToken)

	// 发送请求
	client := &http.Client{
		Timeout: time.Duration(a.config.Timeout) * time.Second,
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	// 解析响应
	var apiResp OpenAIResponse
	err = json.Unmarshal(respBody, &apiResp)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// 检查API错误
	if apiResp.Error != nil {
		return "", fmt.Errorf("API error: %s", apiResp.Error.Message)
	}
	// 检查响应格式
	if len(apiResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response, response: %v, finish", respBody)
	}

	return apiResp.Choices[0].Message.Content, nil
}
