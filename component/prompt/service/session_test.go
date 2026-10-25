package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"microdev/component/prompt/store"
	"microdev/infra/consts"
	promptmodel "microdev/infra/model/prompt"
	"microdev/infra/rdb"
	"microdev/migrate"
)

func newTestSessionSrv(t *testing.T) *SessionSrv {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := rdb.Open(consts.DBDriverSQLite, dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	require.NoError(t, db.ExecAll(context.Background(), migrate.Statements(db.Driver())))
	return NewSessionSrv(store.NewSessionDao(db), store.NewMessageDao(db))
}

func addMessages(t *testing.T, srv *SessionSrv, sessionID int64, role string, count int) {
	t.Helper()
	for i := 0; i < count; i++ {
		param := &promptmodel.AddMessageParam{SessionID: sessionID, Role: role, Content: fmt.Sprintf("%s-%d", role, i)}
		require.NoError(t, srv.AddMessage(context.Background(), param))
	}
}

func TestSessionSrvAddMessageValidation(t *testing.T) {
	srv := newTestSessionSrv(t)
	cases := []struct {
		name    string
		param   *promptmodel.AddMessageParam
		wantErr string
	}{
		{name: "nil param", param: nil, wantErr: "sessionID"},
		{name: "non positive session id", param: &promptmodel.AddMessageParam{Role: consts.RoleUser}, wantErr: "sessionID"},
		{name: "unsupported role", param: &promptmodel.AddMessageParam{SessionID: 1, Role: "system"}, wantErr: "unsupported role"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := srv.AddMessage(context.Background(), tc.param)
			require.Error(t, err)
			assert.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestSessionSrvStatsAndContext(t *testing.T) {
	ctx := context.Background()
	srv := newTestSessionSrv(t)
	session, err := srv.CreateSession(ctx, &promptmodel.CreateSessionParam{})
	require.NoError(t, err)

	addMessages(t, srv, session.ID, consts.RoleUser, 1)
	addMessages(t, srv, session.ID, consts.RoleAssistant, 1)

	stats, err := srv.GetSessionStats(ctx, &promptmodel.SessionIDParam{SessionID: session.ID})
	require.NoError(t, err)
	assert.Equal(t, consts.TrueString, stats.SessionActive)
	assert.EqualValues(t, 2, stats.TotalMessages)
	assert.EqualValues(t, 1, stats.UserMessages)
	assert.EqualValues(t, 1, stats.AssistantMessages)
	assert.Equal(t, consts.FalseString, stats.HasSummary)

	sessionContext, err := srv.GetContext(ctx, &promptmodel.SessionIDParam{SessionID: session.ID})
	require.NoError(t, err)
	assert.Contains(t, sessionContext, "用户: user-0")
	assert.Contains(t, sessionContext, "助手: assistant-0")
	assert.Less(t, indexOf(sessionContext, "user-0"), indexOf(sessionContext, "assistant-0"))

	missing, err := srv.GetSessionStats(ctx, &promptmodel.SessionIDParam{SessionID: session.ID + 100})
	require.NoError(t, err)
	assert.Equal(t, consts.FalseString, missing.SessionActive)
}

func TestSessionSrvCompressMemory(t *testing.T) {
	ctx := context.Background()
	srv := newTestSessionSrv(t)
	session, err := srv.CreateSession(ctx, &promptmodel.CreateSessionParam{})
	require.NoError(t, err)

	addMessages(t, srv, session.ID, consts.RoleUser, int(consts.SessionMaxMessages)+1)

	stats, err := srv.GetSessionStats(ctx, &promptmodel.SessionIDParam{SessionID: session.ID})
	require.NoError(t, err)
	assert.EqualValues(t, consts.SessionMaxMessages/2, stats.TotalMessages)
	assert.Equal(t, consts.TrueString, stats.HasSummary)

	sessionContext, err := srv.GetContext(ctx, &promptmodel.SessionIDParam{SessionID: session.ID})
	require.NoError(t, err)
	assert.Contains(t, sessionContext, "会话摘要")
	assert.Contains(t, sessionContext, "user-8")
}

func TestSessionSrvDeleteAndCleanup(t *testing.T) {
	ctx := context.Background()
	srv := newTestSessionSrv(t)
	first, err := srv.CreateSession(ctx, &promptmodel.CreateSessionParam{})
	require.NoError(t, err)
	second, err := srv.CreateSession(ctx, &promptmodel.CreateSessionParam{})
	require.NoError(t, err)
	addMessages(t, srv, first.ID, consts.RoleUser, 2)

	infos, total, err := srv.SearchSession(ctx, &promptmodel.SearchSessionParam{})
	require.NoError(t, err)
	assert.EqualValues(t, 2, total)
	assert.Len(t, infos, 2)

	require.NoError(t, srv.DeleteSession(ctx, &promptmodel.SessionIDParam{SessionID: first.ID}))
	stats, err := srv.GetSessionStats(ctx, &promptmodel.SessionIDParam{SessionID: first.ID})
	require.NoError(t, err)
	assert.Equal(t, consts.FalseString, stats.SessionActive)

	future := time.Now().Add(time.Hour).UTC().UnixMilli()
	require.NoError(t, srv.CleanupExpiredSession(ctx, &promptmodel.CleanupSessionParam{ExpiredBefore: future}))
	_, total, err = srv.SearchSession(ctx, &promptmodel.SearchSessionParam{ID: second.ID})
	require.NoError(t, err)
	assert.EqualValues(t, 0, total)

	err = srv.CleanupExpiredSession(ctx, &promptmodel.CleanupSessionParam{})
	require.Error(t, err)
	assert.ErrorContains(t, err, "expiredBefore")
}

func indexOf(text, sub string) int {
	for i := 0; i+len(sub) <= len(text); i++ {
		if text[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
