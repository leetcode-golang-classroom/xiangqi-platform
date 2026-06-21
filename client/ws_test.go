package client_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuanyu90221/xiangqi-platform/client"
	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/player"
	"github.com/yuanyu90221/xiangqi-platform/server"
)

func mustMove(t *testing.T, ucci string) board.Move {
	t.Helper()
	m, err := board.ParseUCCI(ucci)
	require.NoError(t, err)
	return m
}

// 起一台 localhost WS server（Hub），行程內兩客戶端對接後回傳 (紅, 黑) WSTransport。
func dialPair(t *testing.T, ctx context.Context) (red, black *client.WSTransport) {
	t.Helper()
	hub := server.NewHub()
	srv := httptest.NewServer(hub.Handler())
	t.Cleanup(srv.Close)
	url := "ws" + strings.TrimPrefix(srv.URL, "http")

	ta, err := client.Dial(ctx, url)
	require.NoError(t, err)
	t.Cleanup(func() { _ = ta.Close() })
	tb, err := client.Dial(ctx, url)
	require.NoError(t, err)
	t.Cleanup(func() { _ = tb.Close() })

	ca, err := ta.WaitMatched(ctx)
	require.NoError(t, err)
	cb, err := tb.WaitMatched(ctx)
	require.NoError(t, err)
	require.NotEqual(t, ca, cb, "雙方執色應相異")

	if ca == "red" {
		return ta, tb
	}
	return tb, ta
}

// Send 上送 move、伺服器確認的對手走法於 Incoming 送出。
func TestWSTransportSendReceivesMoveApplied(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	red, black := dialPair(t, ctx)

	require.NoError(t, red.Send(mustMove(t, "h2e2")))

	select {
	case m := <-black.Incoming():
		assert.Equal(t, "h2e2", m.String(), "對手應於 Incoming 收到伺服器確認的走法")
	case <-ctx.Done():
		t.Fatal("等待對手走法逾時")
	}
}

// WSTransport 滿足 MoveTransport，可包成 RemotePlayer 與回路互換（核心不改）。
func TestWSTransportWrapsRemotePlayer(t *testing.T) {
	var _ player.MoveTransport = (*client.WSTransport)(nil) // 編譯期接縫保證

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	red, black := dialPair(t, ctx)

	rp := player.NewRemotePlayer("對手", black) // 黑方以 RemotePlayer 接入既有迴圈
	require.NoError(t, red.Send(mustMove(t, "h2e2")))

	select {
	case m := <-rp.RequestMove(nil):
		assert.Equal(t, "h2e2", m.String())
	case <-ctx.Done():
		t.Fatal("RemotePlayer 等待走法逾時")
	}
}
