package main

import "golang.org/x/tools/go/analysis"

// AnalyzerPlugin 插件入口函数
func AnalyzerPlugin() []*analysis.Analyzer {
	return []*analysis.Analyzer{Analyzer}
}