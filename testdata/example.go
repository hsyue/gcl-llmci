package testdata

import (
	"fmt"
	"os"
)

// BadFunction 这是一个有问题的函数示例，用于测试LLM分析
func BadFunction() {
	// 没有错误处理
	file, _ := os.Open("nonexistent.txt")
	defer file.Close()

	// 未使用的变量
	unusedVar := "this is not used"

	// 硬编码的值
	for i := 0; i < 100; i++ {
		fmt.Println("Iteration:", i)
	}

	// 可能的空指针解引用
	var ptr *string
	fmt.Println(*ptr)
}

// GoodFunction 这是一个相对较好的函数示例
func GoodFunction(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close file: %v\n", closeErr)
		}
	}()

	// 处理文件内容...
	return nil
}

// ComplexFunction 复杂函数，用于测试更深入的分析
func ComplexFunction(data []string) map[string]int {
	result := make(map[string]int)
	
	// 嵌套循环，可能有性能问题
	for i := 0; i < len(data); i++ {
		for j := 0; j < len(data); j++ {
			if data[i] == data[j] && i != j {
				result[data[i]]++
			}
		}
	}
	
	return result
}