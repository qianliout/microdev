package service

import (
	"regexp"
	"strings"

	"microdev/component/translate/model"
)

// MarkdownService Markdown处理服务
type MarkdownService struct {
	maxChunkSize int
}

// NewMarkdownService 创建新的Markdown处理服务
func NewMarkdownService(maxChunkSize int) *MarkdownService {
	if maxChunkSize <= 0 {
		maxChunkSize = 2000 // 默认大小
	}
	return &MarkdownService{
		maxChunkSize: maxChunkSize,
	}
}

// SplitByHeadings 按标题层级拆分
func (s *MarkdownService) SplitByHeadings(content string) []model.Chunk {
	lines := strings.Split(content, "\n")
	var chunks []model.Chunk
	var currentChunk strings.Builder
	var currentLevel int
	var currentTitle string

	headingRegex := regexp.MustCompile(`^(#{1,6})\s+(.*)$`)

	for _, line := range lines {
		// 检查是否是标题
		if matches := headingRegex.FindStringSubmatch(line); matches != nil {
			// 保存当前块
			if currentChunk.Len() > 0 {
				chunks = append(chunks, model.Chunk{
					Content: strings.TrimSpace(currentChunk.String()),
					Level:   currentLevel,
					Title:   currentTitle,
				})
				currentChunk.Reset()
			}

			// 解析新标题
			currentLevel = len(matches[1]) // 计算#的数量
			currentTitle = strings.TrimSpace(matches[2])
		}

		// 添加行到当前块
		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n")
		}
		currentChunk.WriteString(line)

		// 检查是否需要因为大小限制而拆分
		if currentChunk.Len() > s.maxChunkSize {
			chunks = append(chunks, model.Chunk{
				Content: strings.TrimSpace(currentChunk.String()),
				Level:   currentLevel,
				Title:   currentTitle,
			})
			currentChunk.Reset()
			currentLevel = 0
			currentTitle = ""
		}
	}

	// 添加最后一个块
	if currentChunk.Len() > 0 {
		chunks = append(chunks, model.Chunk{
			Content: strings.TrimSpace(currentChunk.String()),
			Level:   currentLevel,
			Title:   currentTitle,
		})
	}

	return chunks
}

// SplitBySize 按大小拆分
func (s *MarkdownService) SplitBySize(content string) []string {
	if len(content) <= s.maxChunkSize {
		return []string{content}
	}

	var chunks []string
	runes := []rune(content)

	for i := 0; i < len(runes); i += s.maxChunkSize {
		end := i + s.maxChunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}

	return chunks
}

// IsMarkdownContent 检查是否是Markdown内容
func (s *MarkdownService) IsMarkdownContent(content string) bool {
	// 简单检查是否包含Markdown标记
	markdownPatterns := []string{
		"^#+ ",             // 标题
		"```",              // 代码块
		"\\*\\*",           // 加粗
		"\\[.*\\]\\(.*\\)", // 链接
		"^- ",              // 列表
		"^\\* ",            // 列表
		"^> ",              // 引用
	}

	for _, pattern := range markdownPatterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			return true
		}
	}
	return false
}

// SplitContent 智能切分内容
func (s *MarkdownService) SplitContent(content string) []string {
	// 检查是否是Markdown内容
	if s.IsMarkdownContent(content) {
		chunks := s.SplitByHeadings(content)
		var result []string
		for _, chunk := range chunks {
			result = append(result, chunk.Content)
		}
		return result
	}

	// 对于非Markdown内容，使用简单的段落分割
	return s.splitByParagraphs(content)
}

// splitByParagraphs 按段落分割
func (s *MarkdownService) splitByParagraphs(content string) []string {
	// 按段落分割
	paragraphs := strings.Split(content, "\n\n")

	var chunks []string
	var currentChunk strings.Builder

	for _, paragraph := range paragraphs {
		// 如果当前块加上新段落超过最大字符数，则开始新块
		if currentChunk.Len() > 0 && currentChunk.Len()+len(paragraph) > s.maxChunkSize {
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n\n")
		}
		currentChunk.WriteString(paragraph)
	}

	// 添加最后一块
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}
