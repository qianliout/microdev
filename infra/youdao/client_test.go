package youdao

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	youdaomodel "microdev/infra/model/youdao"
)

// writeFakeCLI 生成假的 youdaonote：参数逐行追加到 args.log，save 的笔记文件复制为 note.json，再输出 stdout 并以 code 退出。
func writeFakeCLI(t *testing.T, stdout string, code int) (bin, dir string) {
	t.Helper()
	dir = t.TempDir()
	script := "#!/bin/sh\n" +
		"for a in \"$@\"; do echo \"$a\" >> \"" + dir + "/args.log\"; done\n" +
		"if [ \"$1\" = save ]; then cp \"$3\" \"" + dir + "/note.json\"; fi\n" +
		"echo '" + stdout + "'\n" +
		"[ " + string(rune('0'+code)) + " -eq 0 ] || { echo boom >&2; exit " + string(rune('0'+code)) + "; }\n"
	bin = filepath.Join(dir, "youdaonote")
	require.NoError(t, os.WriteFile(bin, []byte(script), 0o755))
	return bin, dir
}

func TestClientCreate(t *testing.T) {
	cases := []struct {
		name     string
		stdout   string
		code     int
		param    *youdaomodel.CreateNoteParam
		wantID   string
		wantErr  string
		wantNote string
	}{
		{name: "top level fileId", stdout: `{"fileId":"f1"}`, param: &youdaomodel.CreateNoteParam{Title: "t", Content: "c"}, wantID: "f1", wantNote: `{"content":"c","title":"t","type":"md"}`},
		{name: "nested data id with folder", stdout: `{"data":{"id":"f2"}}`, param: &youdaomodel.CreateNoteParam{Title: "t", Content: "c", FolderID: "d1"}, wantID: "f2", wantNote: `{"content":"c","parentId":"d1","title":"t","type":"md"}`},
		{name: "missing fileId returns error", stdout: `{"ok":true}`, param: &youdaomodel.CreateNoteParam{Title: "t"}, wantErr: "fileId not found"},
		{name: "non json output returns error", stdout: `created`, param: &youdaomodel.CreateNoteParam{Title: "t"}, wantErr: "parse youdaonote output"},
		{name: "cli failure carries stderr", stdout: ``, code: 2, param: &youdaomodel.CreateNoteParam{Title: "t"}, wantErr: "boom"},
		{name: "empty title rejected before run", stdout: `{"fileId":"x"}`, param: &youdaomodel.CreateNoteParam{}, wantErr: "title is empty"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bin, dir := writeFakeCLI(t, tc.stdout, tc.code)
			id, err := NewClient(bin).Create(context.Background(), tc.param)
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantID, id)
			note, readErr := os.ReadFile(filepath.Join(dir, "note.json"))
			require.NoError(t, readErr)
			assert.JSONEq(t, tc.wantNote, string(note))
		})
	}
}

func TestClientUpdate(t *testing.T) {
	t.Run("passes fileId and content", func(t *testing.T) {
		bin, dir := writeFakeCLI(t, ``, 0)
		err := NewClient(bin).Update(context.Background(), &youdaomodel.UpdateNoteParam{FileID: "f1", Content: "body"})
		require.NoError(t, err)
		args, readErr := os.ReadFile(filepath.Join(dir, "args.log"))
		require.NoError(t, readErr)
		assert.Equal(t, "update\nf1\n-c\nbody\n", string(args))
	})

	t.Run("cli failure returns error", func(t *testing.T) {
		bin, _ := writeFakeCLI(t, ``, 3)
		err := NewClient(bin).Update(context.Background(), &youdaomodel.UpdateNoteParam{FileID: "f1"})
		assert.ErrorContains(t, err, "boom")
	})

	t.Run("empty fileId rejected", func(t *testing.T) {
		err := NewClient("unused").Update(context.Background(), &youdaomodel.UpdateNoteParam{})
		assert.ErrorContains(t, err, "file id is empty")
	})
}

func TestClientBinNotFound(t *testing.T) {
	err := NewClient("youdaonote-not-installed").Update(context.Background(), &youdaomodel.UpdateNoteParam{FileID: "f1"})
	require.Error(t, err)
	assert.ErrorContains(t, err, "install the official youdaonote CLI")
}
