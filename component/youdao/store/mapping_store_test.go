package store

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	youdaomodel "microdev/infra/model/youdao"
)

func TestMappingStore(t *testing.T) {
	ctx := context.Background()
	dal := NewMappingStore()

	t.Run("missing file loads empty mapping", func(t *testing.T) {
		mapping, err := dal.Load(ctx, &youdaomodel.MappingParam{Root: t.TempDir()})
		require.NoError(t, err)
		assert.NotNil(t, mapping.Notes)
		assert.Empty(t, mapping.Notes)
	})

	t.Run("saved mapping round trips", func(t *testing.T) {
		root := t.TempDir()
		saved := &youdaomodel.Mapping{Notes: map[string]*youdaomodel.MappingEntry{"a/b.md": {FileID: "f1", Hash: "h1"}}}
		require.NoError(t, dal.Save(ctx, &youdaomodel.MappingParam{Root: root, Mapping: saved}))

		loaded, err := dal.Load(ctx, &youdaomodel.MappingParam{Root: root})
		require.NoError(t, err)
		assert.Equal(t, saved, loaded)
	})

	t.Run("corrupted file returns error and is untouched", func(t *testing.T) {
		root := t.TempDir()
		path := mappingPath(root)
		require.NoError(t, os.WriteFile(path, []byte("{bad"), 0o644))

		_, err := dal.Load(ctx, &youdaomodel.MappingParam{Root: root})
		require.Error(t, err)
		assert.ErrorContains(t, err, "corrupted")
		data, readErr := os.ReadFile(path)
		require.NoError(t, readErr)
		assert.Equal(t, "{bad", string(data))
	})

	t.Run("null notes loads as empty map", func(t *testing.T) {
		root := t.TempDir()
		require.NoError(t, os.WriteFile(mappingPath(root), []byte(`{"notes":null}`), 0o644))
		mapping, err := dal.Load(ctx, &youdaomodel.MappingParam{Root: root})
		require.NoError(t, err)
		assert.NotNil(t, mapping.Notes)
	})

	t.Run("empty root and nil mapping are rejected", func(t *testing.T) {
		_, err := dal.Load(ctx, &youdaomodel.MappingParam{})
		assert.Error(t, err)
		err = dal.Save(ctx, &youdaomodel.MappingParam{Root: t.TempDir()})
		assert.ErrorContains(t, err, "mapping is nil")
	})
}
