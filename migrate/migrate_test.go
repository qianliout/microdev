package migrate

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"microdev/infra/consts"
)

func TestStatements(t *testing.T) {
	cases := []struct {
		name       string
		driver     string
		wantCount  int
		wantPrefix string
	}{
		{name: "sqlite has two tables and two indexes", driver: consts.DBDriverSQLite, wantCount: 4, wantPrefix: "CREATE TABLE IF NOT EXISTS prompt_session"},
		{name: "mysql has two tables with inline indexes", driver: consts.DBDriverMySQL, wantCount: 2, wantPrefix: "CREATE TABLE IF NOT EXISTS prompt_session"},
		{name: "unsupported driver returns nil", driver: "postgres", wantCount: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Statements(tc.driver)
			assert.Len(t, got, tc.wantCount)
			if tc.wantCount > 0 {
				assert.Contains(t, got[0], tc.wantPrefix)
				assert.NotContains(t, got[0], "--")
			}
		})
	}
}
