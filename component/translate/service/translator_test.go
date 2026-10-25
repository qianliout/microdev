package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"microdev/infra/consts"
)

func TestChunkText(t *testing.T) {
	cases := []struct {
		name       string
		text       string
		size       int64
		wantChunks int
		wantLast   string
	}{
		{name: "empty text yields no chunk", text: "", size: 3, wantChunks: 0},
		{name: "text shorter than size is one chunk", text: "ab", size: 3, wantChunks: 1, wantLast: "ab"},
		{name: "text splits by rune not byte", text: "你好世界啊", size: 2, wantChunks: 3, wantLast: "啊"},
		{name: "non positive size falls back to default", text: strings.Repeat("a", 2001), size: 0, wantChunks: 2, wantLast: "a"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := chunkText(tc.text, tc.size)
			assert.Len(t, got, tc.wantChunks)
			if tc.wantChunks > 0 {
				assert.Equal(t, tc.wantLast, got[len(got)-1])
				assert.Equal(t, tc.text, strings.Join(got, ""))
			}
		})
	}
}

func TestDetectDirection(t *testing.T) {
	cases := []struct {
		name string
		text string
		want string
	}{
		{name: "any han char means zh2en", text: "hello 你", want: consts.DirectionZh2En},
		{name: "english means en2zh", text: "hello", want: consts.DirectionEn2Zh},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, detectDirection(tc.text))
		})
	}
}
