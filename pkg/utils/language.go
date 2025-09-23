// Package utils 提供通用的工具函数
package utils

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// DetectLanguage 检测文本的主要语言
// 返回 "zh2en" 表示中译英，"en2zh" 表示英译中
func DetectLanguage(text string) string {
	chineseCount := 0
	englishCount := 0
	totalCount := 0

	for _, r := range text {
		if unicode.IsLetter(r) {
			totalCount++
			if IsChinese(r) {
				chineseCount++
			} else if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				englishCount++
			}
		}
	}

	if totalCount == 0 {
		return "zh2en" // 默认中译英
	}

	if float64(chineseCount)/float64(totalCount) > 0.3 {
		return "zh2en" // 中译英
	}
	return "en2zh" // 英译中
}

// IsChinese 判断字符是否为中文字符
func IsChinese(r rune) bool {
	return r >= 0x4e00 && r <= 0x9fff
}

// GenerateOutputPath 根据输入文件路径和内容生成输出文件路径
func GenerateOutputPath(inputPath string, content string) string {
	// 检测内容主要语言
	direction := DetectLanguage(content)

	// 获取文件扩展名和基础名
	ext := filepath.Ext(inputPath)
	baseName := strings.TrimSuffix(inputPath, ext)

	// 根据翻译方向添加后缀
	var suffix string
	if direction == "zh2en" {
		suffix = "_en"
	} else {
		suffix = "_zh"
	}

	return baseName + suffix + ext
}

// IsTranslationExists 检查翻译文件是否已存在（幂等性检查）
func IsTranslationExists(inputPath string, content string) bool {
	outputPath := GenerateOutputPath(inputPath, content)
	_, err := os.Stat(outputPath)
	return err == nil // 文件存在返回true
}

// GetTranslationOutputPath 获取翻译输出路径（确保在同目录下）
func GetTranslationOutputPath(inputPath string, content string) string {
	// 检测内容主要语言
	direction := DetectLanguage(content)

	// 获取目录、文件名和扩展名
	dir := filepath.Dir(inputPath)
	filename := filepath.Base(inputPath)
	ext := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, ext)

	// 根据翻译方向添加后缀
	var suffix string
	if direction == "zh2en" {
		suffix = "_en"
	} else {
		suffix = "_zh"
	}

	// 确保输出文件在同一目录下
	outputFilename := baseName + suffix + ext
	return filepath.Join(dir, outputFilename)
}
