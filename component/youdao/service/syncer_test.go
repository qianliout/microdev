package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"microdev/component/youdao/store"
	"microdev/infra/consts"
	youdaomodel "microdev/infra/model/youdao"
)

type fakeNoteClient struct {
	creates   []*youdaomodel.CreateNoteParam
	updates   []*youdaomodel.UpdateNoteParam
	createErr error
	updateErr error
}

func (f *fakeNoteClient) Create(_ context.Context, param *youdaomodel.CreateNoteParam) (string, error) {
	if f.createErr != nil {
		return "", f.createErr
	}
	f.creates = append(f.creates, param)
	return "id-" + param.Title, nil
}

func (f *fakeNoteClient) Update(_ context.Context, param *youdaomodel.UpdateNoteParam) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.updates = append(f.updates, param)
	return nil
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func actions(summary *youdaomodel.SyncSummary) map[string]string {
	res := make(map[string]string, len(summary.Results))
	for _, r := range summary.Results {
		res[r.Path] = r.Action
	}
	return res
}

func TestSyncDirectory(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.md"), "A")
	writeFile(t, filepath.Join(dir, "sub", "b.MD"), "B")
	writeFile(t, filepath.Join(dir, "note.txt"), "ignored")
	writeFile(t, filepath.Join(dir, ".hidden", "c.md"), "ignored")
	writeFile(t, filepath.Join(dir, ".d.md"), "ignored")

	cli := &fakeNoteClient{}
	syncer := NewSyncerSrv(store.NewMappingStore(), cli)
	param := func() *youdaomodel.SyncParam { return &youdaomodel.SyncParam{Dir: dir, Folder: "f9"} }

	first, err := syncer.Sync(ctx, param())
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"a.md": consts.SyncActionCreate, "sub/b.MD": consts.SyncActionCreate}, actions(first))
	assert.EqualValues(t, 2, first.Created)
	require.Len(t, cli.creates, 2)
	assert.Equal(t, "f9", cli.creates[0].FolderID)
	assert.Equal(t, "b", cli.creates[1].Title)

	second, err := syncer.Sync(ctx, param())
	require.NoError(t, err)
	assert.EqualValues(t, 2, second.Skipped)
	assert.Len(t, cli.creates, 2)
	assert.Empty(t, cli.updates)

	writeFile(t, filepath.Join(dir, "a.md"), "A2")
	third, err := syncer.Sync(ctx, param())
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"a.md": consts.SyncActionUpdate, "sub/b.MD": consts.SyncActionSkip}, actions(third))
	require.Len(t, cli.updates, 1)
	assert.Equal(t, "id-a", cli.updates[0].FileID)
	assert.Equal(t, "A2", cli.updates[0].Content)
	assert.Equal(t, "folder ignored, note already exists in cloud", third.Results[0].Note)
}

func TestSyncDryRun(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.md"), "A")
	cli := &fakeNoteClient{}
	syncer := NewSyncerSrv(store.NewMappingStore(), cli)

	summary, err := syncer.Sync(context.Background(), &youdaomodel.SyncParam{Dir: dir, DryRun: true})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"a.md": consts.SyncActionCreate}, actions(summary))
	assert.Empty(t, cli.creates)
	_, statErr := os.Stat(filepath.Join(dir, consts.YoudaoMappingFile))
	assert.True(t, os.IsNotExist(statErr))
}

func TestSyncSingleFile(t *testing.T) {
	dir := t.TempDir()
	outside := t.TempDir()
	writeFile(t, filepath.Join(dir, "in", "a.md"), "A")
	writeFile(t, filepath.Join(outside, "o.md"), "O")
	writeFile(t, filepath.Join(dir, "x.txt"), "X")

	cases := []struct {
		name      string
		param     *youdaomodel.SyncParam
		wantErr   string
		wantTitle string
		wantRoot  string
		wantKey   string
	}{
		{name: "file inside dir shares dir mapping", param: &youdaomodel.SyncParam{Path: filepath.Join(dir, "in", "a.md"), Dir: dir}, wantTitle: "a", wantRoot: dir, wantKey: "in/a.md"},
		{name: "file outside dir uses own directory", param: &youdaomodel.SyncParam{Path: filepath.Join(outside, "o.md"), Dir: dir, Title: "custom"}, wantTitle: "custom", wantRoot: outside, wantKey: "o.md"},
		{name: "non markdown file is rejected", param: &youdaomodel.SyncParam{Path: filepath.Join(dir, "x.txt"), Dir: dir}, wantErr: "is not a .md file"},
		{name: "title with directory is rejected", param: &youdaomodel.SyncParam{Path: dir, Dir: dir, Title: "t"}, wantErr: "single file"},
		{name: "missing path hints env", param: &youdaomodel.SyncParam{Path: filepath.Join(dir, "nope"), Dir: dir}, wantErr: consts.EnvYoudaoDir},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cli := &fakeNoteClient{}
			mappingStore := store.NewMappingStore()
			syncer := NewSyncerSrv(mappingStore, cli)

			summary, err := syncer.Sync(context.Background(), tc.param)
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Len(t, cli.creates, 1)
			assert.Equal(t, tc.wantTitle, cli.creates[0].Title)
			assert.Equal(t, tc.wantKey, summary.Results[0].Path)
			mapping, loadErr := mappingStore.Load(context.Background(), &youdaomodel.MappingParam{Root: tc.wantRoot})
			require.NoError(t, loadErr)
			assert.Contains(t, mapping.Notes, tc.wantKey)
		})
	}
}

func TestSyncFailures(t *testing.T) {
	ctx := context.Background()

	t.Run("one failure does not stop others and leaves mapping untouched", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "a.md"), "A")
		writeFile(t, filepath.Join(dir, "b.md"), "B")
		cli := &fakeNoteClient{createErr: errors.New("api down")}
		syncer := NewSyncerSrv(store.NewMappingStore(), cli)

		summary, err := syncer.Sync(ctx, &youdaomodel.SyncParam{Dir: dir})
		require.NoError(t, err)
		assert.EqualValues(t, 2, summary.Failed)
		assert.ErrorContains(t, summary.Results[0].Err, "api down")
		_, statErr := os.Stat(filepath.Join(dir, consts.YoudaoMappingFile))
		assert.True(t, os.IsNotExist(statErr))
	})

	t.Run("update failure keeps old hash so next run retries", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "a.md"), "A")
		cli := &fakeNoteClient{}
		syncer := NewSyncerSrv(store.NewMappingStore(), cli)
		_, err := syncer.Sync(ctx, &youdaomodel.SyncParam{Dir: dir})
		require.NoError(t, err)

		writeFile(t, filepath.Join(dir, "a.md"), "A2")
		cli.updateErr = errors.New("gone")
		failed, err := syncer.Sync(ctx, &youdaomodel.SyncParam{Dir: dir})
		require.NoError(t, err)
		assert.EqualValues(t, 1, failed.Failed)

		cli.updateErr = nil
		retry, err := syncer.Sync(ctx, &youdaomodel.SyncParam{Dir: dir})
		require.NoError(t, err)
		assert.EqualValues(t, 1, retry.Updated)
	})

	t.Run("corrupted mapping aborts sync", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "a.md"), "A")
		writeFile(t, filepath.Join(dir, consts.YoudaoMappingFile), "{bad")
		cli := &fakeNoteClient{}
		syncer := NewSyncerSrv(store.NewMappingStore(), cli)

		_, err := syncer.Sync(ctx, &youdaomodel.SyncParam{Dir: dir})
		require.Error(t, err)
		assert.ErrorContains(t, err, "corrupted")
		assert.Empty(t, cli.creates)
	})

	t.Run("empty directory yields no results", func(t *testing.T) {
		syncer := NewSyncerSrv(store.NewMappingStore(), &fakeNoteClient{})
		summary, err := syncer.Sync(ctx, &youdaomodel.SyncParam{Dir: t.TempDir()})
		require.NoError(t, err)
		assert.Empty(t, summary.Results)
	})

	t.Run("empty dir param is rejected", func(t *testing.T) {
		syncer := NewSyncerSrv(store.NewMappingStore(), &fakeNoteClient{})
		_, err := syncer.Sync(ctx, &youdaomodel.SyncParam{})
		assert.ErrorContains(t, err, "default dir is empty")
	})
}
