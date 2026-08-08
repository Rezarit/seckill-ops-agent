package chat_pipeline

import (
	"strings"
	"testing"
)

// genChineseText 生成指定 rune 数量的中文测试文本
func genChineseText(n int) string {
	var b strings.Builder
	base := "春眠不觉晓处处闻啼鸟夜来风雨声花落知多少"
	for b.Len() < n*3 { // 每个中文字符占 3 字节，留出余量
		b.WriteString(base)
	}
	runes := []rune(b.String())
	if len(runes) > n {
		runes = runes[:n]
	}
	return string(runes)
}

func TestTruncateByRune(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLen    int
		shouldCut bool // 是否应该被截断
	}{
		{"纯英文短于限制", "hello world", 500, false},
		{"纯英文长于限制", string(make([]byte, 600)), 500, true},
		{"纯中文短于限制", "你好世界", 500, false},
		{"纯中文长于限制", genChineseText(600), 500, true},
	}

	const suffix = "...[truncated]"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateByRune(tt.input, tt.maxLen)
			originalRunes := []rune(tt.input)
			resultRunes := []rune(result)

			if tt.shouldCut {
				// 长字符串：结果应由 maxLen 个原文 rune + suffix 组成
				if len(resultRunes) != tt.maxLen+len([]rune(suffix)) {
					t.Errorf("截断后 rune 数应为 %d (maxLen=%d + suffix=%d)，实际 %d",
						tt.maxLen+len([]rune(suffix)), tt.maxLen, len([]rune(suffix)), len(resultRunes))
				}
			} else {
				// 短字符串：结果应与原文完全相同
				if result != tt.input {
					t.Errorf("短字符串不应被修改: 原文=%q 结果=%q", tt.input, result)
				}
			}

			// 验证前缀匹配（原文字符不被破坏）
			checkLen := tt.maxLen
			if len(originalRunes) < checkLen {
				checkLen = len(originalRunes)
			}
			for i := 0; i < checkLen; i++ {
				if resultRunes[i] != originalRunes[i] {
					t.Errorf("第 %d 个字符不匹配: 原文=%c (U+%04X) 截断后=%c (U+%04X)",
						i, originalRunes[i], originalRunes[i], resultRunes[i], resultRunes[i])
				}
			}
		})
	}
}
