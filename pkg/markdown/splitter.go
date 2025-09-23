package markdown

import (
	"regexp"
	"strings"
)

// Chunk 表示一个Markdown内容块
type Chunk struct {
	Content string
	Level   int    // 标题层级，0表示非标题块
	Title   string // 标题文本
}

// Splitter Markdown拆分器
type Splitter struct {
	maxChunkSize int
}

// NewSplitter 创建新的Markdown拆分器
func NewSplitter(maxChunkSize int) *Splitter {
	return &Splitter{
		maxChunkSize: maxChunkSize,
	}
}

// SplitByHeadings 按标题层级拆分（简化版本）
func (s *Splitter) SplitByHeadings(content string) []Chunk {
	lines := strings.Split(content, "\n")
	var chunks []Chunk
	var currentChunk strings.Builder
	var currentLevel int
	var currentTitle string

	headingRegex := regexp.MustCompile(`^(#{1,6})\s+(.*)$`)

	for _, line := range lines {
		// 检查是否是标题
		if matches := headingRegex.FindStringSubmatch(line); matches != nil {
			// 保存当前块
			if currentChunk.Len() > 0 {
				chunks = append(chunks, Chunk{
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
			chunks = append(chunks, Chunk{
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
		chunks = append(chunks, Chunk{
			Content: strings.TrimSpace(currentChunk.String()),
			Level:   currentLevel,
			Title:   currentTitle,
		})
	}

	return chunks
}
