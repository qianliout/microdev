package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"microdev/infra/consts"
)

func TestLoadConfig(t *testing.T) {
	cases := []struct {
		name    string
		env     map[string]string
		wantErr string
		check   func(t *testing.T, cfg *Config)
	}{
		{name: "missing api keys returns error", env: map[string]string{}, wantErr: consts.EnvDashScopeAPIKey},
		{
			name: "defaults apply when optional env missing",
			env:  map[string]string{consts.EnvAliBailianAPIKey: "bailian"},
			check: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "bailian", cfg.GetAPIKey())
				assert.Equal(t, consts.DefaultModelName, cfg.ModelName)
				assert.Equal(t, consts.DefaultMaxTokens, cfg.MaxTokens)
				assert.Equal(t, consts.DefaultTemperature, cfg.Temperature)
				assert.True(t, cfg.StreamOutput)
				assert.Equal(t, consts.DefaultConcurrency, cfg.Concurrency)
				assert.Equal(t, consts.DefaultSessionDriver, cfg.SessionDriver)
			},
		},
		{
			name: "values are clamped and invalid ones fall back",
			env: map[string]string{
				consts.EnvDashScopeAPIKey: "dash",
				consts.EnvMaxTokens:       "99999",
				consts.EnvConcurrency:     "0",
				consts.EnvTemperature:     "5",
				consts.EnvModelName:       "unknown-model",
				consts.EnvStreamOutput:    "false",
			},
			check: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "dash", cfg.GetAPIKey())
				assert.Equal(t, consts.MaxMaxTokens, cfg.MaxTokens)
				assert.Equal(t, consts.DefaultConcurrency, cfg.Concurrency)
				assert.Equal(t, consts.DefaultTemperature, cfg.Temperature)
				assert.Equal(t, consts.DefaultModelName, cfg.ModelName)
				assert.False(t, cfg.StreamOutput)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, name := range []string{
				consts.EnvDashScopeAPIKey, consts.EnvAliBailianAPIKey, consts.EnvModelName, consts.EnvMaxTokens,
				consts.EnvTemperature, consts.EnvStreamOutput, consts.EnvConcurrency,
			} {
				t.Setenv(name, "")
			}
			for name, value := range tc.env {
				t.Setenv(name, value)
			}
			cfg, err := LoadConfig()
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			tc.check(t, cfg)
		})
	}
}
