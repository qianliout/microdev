// Package utils 提供通用的工具函数。
package utils

import (
	"path/filepath"
	"strings"
	"unicode"

	"microdev/infra/consts"
)

type langCount struct {
	letters int64
	chinese int64
}

func (vi *langCount) add(r rune) {
	if !unicode.IsLetter(r) {
		return
	}
	vi.letters++
	if IsChinese(r) {
		vi.chinese++
	}
}

// DetectLanguage 按汉字占字母总数的比例判断翻译方向，无字母时默认中译英。
func DetectLanguage(text string) string {
	count := &langCount{}
	for _, r := range text {
		count.add(r)
	}
	if count.letters == 0 {
		return consts.DirectionZh2En
	}
	ratio := float64(count.chinese) / float64(count.letters)
	if ratio > consts.ChineseRatioThreshold {
		return consts.DirectionZh2En
	}
	return consts.DirectionEn2Zh
}

// HasChinese 只要文本含任意汉字即返回 true。
func HasChinese(text string) bool {
	return strings.IndexFunc(text, IsChinese) >= 0
}

func IsChinese(r rune) bool {
	return r >= 0x4e00 && r <= 0x9fff
}

// suffixFor 返回翻译方向对应的文件名后缀。
func suffixFor(direction string) string {
	switch direction {
	case consts.DirectionZh2En:
		return consts.SuffixZh2En
	default:
		return consts.SuffixEn2Zh
	}
}

// GenerateOutputPath 根据输入文件路径和内容生成翻译输出文件路径。
func GenerateOutputPath(inputPath string, content string) string {
	suffix := suffixFor(DetectLanguage(content))
	ext := filepath.Ext(inputPath)
	baseName := strings.TrimSuffix(inputPath, ext)
	return baseName + suffix + ext
}

// IsTranslationExists 用于幂等检查：目标翻译文件已存在时返回 true。
func IsTranslationExists(inputPath string, content string) bool {
	outputPath := GenerateOutputPath(inputPath, content)
	return fileExists(outputPath)
}

// GetTranslationOutputPath 返回与输入文件同目录的翻译输出路径。
func GetTranslationOutputPath(inputPath string, content string) string {
	suffix := suffixFor(DetectLanguage(content))
	dir := filepath.Dir(inputPath)
	filename := filepath.Base(inputPath)
	ext := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, ext)
	return filepath.Join(dir, baseName+suffix+ext)
}
