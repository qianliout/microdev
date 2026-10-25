package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"microdev/infra/consts"
	filemodel "microdev/infra/model/file"
)

func TestAnalyze(t *testing.T) {
	dir := t.TempDir()
	plain := filepath.Join(dir, "plain.txt")
	require.NoError(t, os.WriteFile(plain, []byte("not a binary"), 0o644))
	self, err := os.Executable()
	require.NoError(t, err)

	cases := []struct {
		name    string
		path    string
		wantErr string
		check   func(t *testing.T, res *filemodel.Result)
	}{
		{
			name: "go test binary is detected as go executable",
			path: self,
			check: func(t *testing.T, res *filemodel.Result) {
				assert.Equal(t, consts.TrueString, res.IsGoBinary)
				assert.Equal(t, consts.TrueString, res.IsExecutable)
				assert.Equal(t, consts.TrueString, res.HasExecPerms)
				assert.NotEqual(t, consts.Unknown, res.GoVersion)
			},
		},
		{
			name: "plain text file gets unknown format and defaults",
			path: plain,
			check: func(t *testing.T, res *filemodel.Result) {
				assert.Equal(t, consts.Unknown, res.Format)
				assert.Equal(t, consts.FalseString, res.IsGoBinary)
				assert.Equal(t, consts.FalseString, res.HasExecPerms)
				assert.Equal(t, plain, res.Path)
			},
		},
		{name: "missing file returns stat error", path: filepath.Join(dir, "missing"), wantErr: "stat file"},
		{name: "empty path is rejected", path: "  ", wantErr: "path is empty"},
	}
	srv := NewAnalyzerSrv()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := srv.Analyze(context.Background(), &filemodel.AnalyzeParam{Path: tc.path})
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			tc.check(t, res)
		})
	}
}
