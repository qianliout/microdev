package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"microdev/infra/consts"
)

func TestLoadYoudaoConfig(t *testing.T) {
	home := t.TempDir()
	cases := []struct {
		name    string
		env     map[string]string
		wantDir string
		wantBin string
	}{
		{name: "defaults expand home", env: map[string]string{}, wantDir: filepath.Join(home, "youdao-notes"), wantBin: consts.DefaultYoudaoBin},
		{name: "blank values fall back", env: map[string]string{consts.EnvYoudaoDir: "  ", consts.EnvYoudaoBin: " "}, wantDir: filepath.Join(home, "youdao-notes"), wantBin: consts.DefaultYoudaoBin},
		{name: "tilde path expands", env: map[string]string{consts.EnvYoudaoDir: "~/notes"}, wantDir: filepath.Join(home, "notes"), wantBin: consts.DefaultYoudaoBin},
		{name: "absolute dir and bin kept", env: map[string]string{consts.EnvYoudaoDir: "/data/notes", consts.EnvYoudaoBin: "/opt/yd"}, wantDir: "/data/notes", wantBin: "/opt/yd"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("HOME", home)
			t.Setenv(consts.EnvYoudaoDir, tc.env[consts.EnvYoudaoDir])
			t.Setenv(consts.EnvYoudaoBin, tc.env[consts.EnvYoudaoBin])
			cfg := LoadYoudaoConfig()
			assert.Equal(t, tc.wantDir, cfg.Dir)
			assert.Equal(t, tc.wantBin, cfg.Bin)
		})
	}
}
