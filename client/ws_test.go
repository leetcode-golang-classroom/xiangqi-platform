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
	"github.com/yuanyu90221/xiangqi-platform/core/play"
	"github.com/yuanyu90221/xiangqi-platform/player"
	"github.com/yuanyu90221/xiangqi-platform/server"
)

func mustMove(t *testing.T, ucci string) board.Move {
	t.Helper()
	m, err := board.ParseUCCI(ucci)
	require.NoError(t, err)
	return m
}

func mustSquare(t *testing.T, s string) board.Square {
	t.Helper()
	q, err := board.ParseSquare(s)
	require.NoError(t, err)
	return q
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

// 端對端：兩個 OnlineController 經真實 Hub 對局——紅方人類落子只上送，待伺服器確認回聲，
// 雙方本地盤面才同步推進。這是 GUI 線上流程（去除渲染後）的最強自動化代理。
func TestOnlineControllersPlayOverWS(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	redTr, blackTr := dialPair(t, ctx)

	ocR, hR := play.NewOnlineController("你", "對手", board.Red, redTr)
	ocB, _ := play.NewOnlineController("對手", "你", board.Black, blackTr)
	pump := func() { ocR.Step(); ocB.Step() }

	// 紅方先手：推進至紅（人類）可互動。
	require.Eventually(t, func() bool {
		pump()
		_, ok := ocR.CurrentInteractive()
		return ok
	}, 2*time.Second, 5*time.Millisecond, "紅方應可互動")

	// 紅方人類落子 h2e2（只上送，不立即套用）。
	require.Equal(t, player.TapSelected, hR.Tap(mustSquare(t, "h2")))
	require.Equal(t, player.TapMoved, hR.Tap(mustSquare(t, "e2")))

	// 推進至雙方皆收到伺服器確認、盤面同步。
	require.Eventually(t, func() bool {
		pump()
		return len(ocR.Session().Record().Moves) == 1 && len(ocB.Session().Record().Moves) == 1
	}, 2*time.Second, 5*time.Millisecond, "雙方應同步套用伺服器確認的走法")

	assert.Equal(t, []string{"h2e2"}, ocR.Session().Record().Moves)
	assert.Equal(t, []string{"h2e2"}, ocB.Session().Record().Moves)
	assert.Equal(t, board.Black, ocR.Turn(), "紅走完換黑")
	assert.Equal(t, board.Black, ocB.Turn())

	// 換黑方（人類在 B 端）可互動，紅方不可。
	_, okB := ocB.CurrentInteractive()
	assert.True(t, okB, "換黑方人類回合，B 端應可互動")
	_, okR := ocR.CurrentInteractive()
	assert.False(t, okR, "非紅方回合，R 端不可互動")
}
