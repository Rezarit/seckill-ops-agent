package splitter

import (
	"context"
	"strings"

	"github.com/cloudwego/eino/schema"
)

// ChunkStrategy 分块策略枚举
type ChunkStrategy int

const (
	StrategyMarkdown      ChunkStrategy = iota // 按 Markdown 标题分块
	StrategySlidingWindow                      // 滑动窗口分块
	StrategyParagraph                          // 按段落分块
)

// SlidingWindowSplitter 滑动窗口分块器
type SlidingWindowSplitter struct {
	chunkSize   int // 块大小（字符数）
	overlapSize int // 重叠大小（字符数）
}

// NewSlidingWindowSplitter 创建滑动窗口分块器
func NewSlidingWindowSplitter(chunkSize, overlapSize int) *SlidingWindowSplitter {
	return &SlidingWindowSplitter{
		chunkSize:   chunkSize,
		overlapSize: overlapSize,
	}
}

func (s *SlidingWindowSplitter) SplitText(ctx context.Context, text string, meta map[string]any) ([]*schema.Document, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return []*schema.Document{}, nil
	}

	var docs []*schema.Document
	start := 0
	textLen := len(text)
	chunkIndex := 0

	for start < textLen {
		end := min(start+s.chunkSize, textLen)
		chunk := text[start:end]

		if strings.TrimSpace(chunk) != "" {
			docMeta := make(map[string]any)
			for k, v := range meta {
				docMeta[k] = v
			}
			docMeta["chunk_index"] = chunkIndex
			docMeta["start_pos"] = start
			docMeta["end_pos"] = end

			docs = append(docs, &schema.Document{
				Content:  strings.TrimSpace(chunk),
				MetaData: docMeta,
			})
			chunkIndex++
		}

		if start+s.chunkSize >= textLen {
			break
		}
		start += s.chunkSize - s.overlapSize
	}

	return docs, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MarkdownSplitter Markdown 分块器
type MarkdownSplitter struct {
	splitFunc func(text string) []string
}

func NewMarkdownSplitter() *MarkdownSplitter {
	return &MarkdownSplitter{
		splitFunc: splitByHeaders,
	}
}

func (s *MarkdownSplitter) SplitText(ctx context.Context, text string, meta map[string]any) ([]*schema.Document, error) {
	chunks := s.splitFunc(text)
	docs := make([]*schema.Document, 0, len(chunks))

	for i, chunk := range chunks {
		if strings.TrimSpace(chunk) == "" {
			continue
		}

		docMeta := make(map[string]any)
		for k, v := range meta {
			docMeta[k] = v
		}
		docMeta["chunk_index"] = i

		docs = append(docs, &schema.Document{
			Content:  strings.TrimSpace(chunk),
			MetaData: docMeta,
		})
	}

	return docs, nil
}

// ConfigurableSplitter 可配置分块器
type ConfigurableSplitter struct {
	strategy    ChunkStrategy
	chunkSize   int
	overlapSize int
}

// NewConfigurableSplitter 创建可配置分块器
func NewConfigurableSplitter(strategy ChunkStrategy, chunkSize, overlapSize int) *ConfigurableSplitter {
	return &ConfigurableSplitter{
		strategy:    strategy,
		chunkSize:   chunkSize,
		overlapSize: overlapSize,
	}
}

func (s *ConfigurableSplitter) SplitText(ctx context.Context, text string, meta map[string]any) ([]*schema.Document, error) {
	switch s.strategy {
	case StrategySlidingWindow:
		splitter := NewSlidingWindowSplitter(s.chunkSize, s.overlapSize)
		return splitter.SplitText(ctx, text, meta)
	case StrategyMarkdown:
		splitter := NewMarkdownSplitter()
		return splitter.SplitText(ctx, text, meta)
	case StrategyParagraph:
		return splitByParagraphsWithMeta(ctx, text, meta)
	default:
		splitter := NewSlidingWindowSplitter(s.chunkSize, s.overlapSize)
		return splitter.SplitText(ctx, text, meta)
	}
}

// splitByHeaders 按 Markdown 标题分割文本
func splitByHeaders(text string) []string {
	lines := strings.Split(text, "\n")
	var chunks []string
	var currentChunk []string

	for _, line := range lines {
		// 检测标题级别
		level := getHeaderLevel(line)

		if level > 0 {
			// 遇到标题，保存当前块
			if len(currentChunk) > 0 {
				chunk := strings.Join(currentChunk, "\n")
				if strings.TrimSpace(chunk) != "" {
					chunks = append(chunks, chunk)
				}
				currentChunk = nil
			}
			// 标题作为新块的开始
			currentChunk = append(currentChunk, line)
		} else {
			// 普通行，添加到当前块
			currentChunk = append(currentChunk, line)
		}
	}

	// 保存最后一块
	if len(currentChunk) > 0 {
		chunk := strings.Join(currentChunk, "\n")
		if strings.TrimSpace(chunk) != "" {
			chunks = append(chunks, chunk)
		}
	}

	return chunks
}

// getHeaderLevel 返回标题级别（1-6），0 表示非标题
func getHeaderLevel(line string) int {
	trimmed := strings.TrimLeft(line, " \t")
	if !strings.HasPrefix(trimmed, "#") {
		return 0
	}

	level := 0
	for _, c := range trimmed {
		if c == '#' {
			level++
		} else {
			break
		}
	}

	if level > 0 && level <= 6 && len(trimmed) > level && trimmed[level] == ' ' {
		return level
	}
	return 0
}

// SplitByLines 按行数分割（备用简单策略）
func SplitByLines(text string, linesPerChunk int) []string {
	lines := strings.Split(text, "\n")
	var chunks []string
	var currentChunk []string

	for i, line := range lines {
		currentChunk = append(currentChunk, line)
		if (i+1)%linesPerChunk == 0 {
			chunks = append(chunks, strings.Join(currentChunk, "\n"))
			currentChunk = nil
		}
	}

	if len(currentChunk) > 0 {
		chunks = append(chunks, strings.Join(currentChunk, "\n"))
	}

	return chunks
}

// splitByParagraphsWithMeta 按段落分块（带元数据）
func splitByParagraphsWithMeta(ctx context.Context, text string, meta map[string]any) ([]*schema.Document, error) {
	paragraphs := strings.Split(text, "\n\n")
	var docs []*schema.Document

	for i, p := range paragraphs {
		if strings.TrimSpace(p) != "" {
			docMeta := make(map[string]any)
			for k, v := range meta {
				docMeta[k] = v
			}
			docMeta["chunk_index"] = i

			docs = append(docs, &schema.Document{
				Content:  strings.TrimSpace(p),
				MetaData: docMeta,
			})
		}
	}
	return docs, nil
}

// SplitByParagraphs 按空行分割（段落策略）
func SplitByParagraphs(text string) []string {
	paragraphs := strings.Split(text, "\n\n")
	var chunks []string
	for _, p := range paragraphs {
		if strings.TrimSpace(p) != "" {
			chunks = append(chunks, strings.TrimSpace(p))
		}
	}
	return chunks
}
