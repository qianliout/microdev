package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"microdev/infra/consts"
)

func TestDetectLanguage(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "pure chinese translates to english", input: "你好，世界", want: consts.DirectionZh2En},
		{name: "pure english translates to chinese", input: "hello world", want: consts.DirectionEn2Zh},
		{name: "no letters defaults to zh2en", input: "123 !?", want: consts.DirectionZh2En},
		{name: "chinese ratio below threshold stays en2zh", input: "hello world 你好", want: consts.DirectionEn2Zh},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, DetectLanguage(tc.input))
		})
	}
}

func TestHasChinese(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "mixed text with one han char", input: "hello 你", want: true},
		{name: "english only", input: "hello", want: false},
		{name: "empty text", input: "", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, HasChinese(tc.input))
		})
	}
}

func TestGenerateOutputPath(t *testing.T) {
	cases := []struct {
		name    string
		path    string
		content string
		want    string
	}{
		{name: "chinese content gets _en suffix", path: "a/doc.md", content: "你好", want: "a/doc_en.md"},
		{name: "english content gets _zh suffix", path: "a/doc.md", content: "hello", want: "a/doc_zh.md"},
		{name: "file without extension", path: "doc", content: "hello", want: "doc_zh"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, GenerateOutputPath(tc.path, tc.content))
			assert.Equal(t, filepath.Clean(tc.want), GetTranslationOutputPath(tc.path, tc.content))
		})
	}
}

func TestReadInput(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "in.txt")
	require.NoError(t, os.WriteFile(file, []byte("from file"), 0o600))

	cases := []struct {
		name    string
		arg     string
		want    string
		wantErr string
	}{
		{name: "existing file is read", arg: file, want: "from file"},
		{name: "plain text is returned as is", arg: "just text", want: "just text"},
		{name: "missing path is treated as text", arg: filepath.Join(dir, "missing.txt"), want: filepath.Join(dir, "missing.txt")},
		{name: "blank input returns error", arg: "   ", wantErr: "empty input"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ReadInput(tc.arg)
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
