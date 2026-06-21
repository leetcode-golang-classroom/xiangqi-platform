package player_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuanyu90221/xiangqi-platform/core/rules"
	"github.com/yuanyu90221/xiangqi-platform/player"
)

// LoopbackTransport 須實作 MoveTransport 接縫。
var _ player.MoveTransport = (*player.LoopbackTransport)(nil)

// 一端 Send 的走法應成為對向端 Incoming 送出的走法（雙向）。
func TestLoopbackForwardsBothDirections(t *testing.T) {
	a, b := player.NewLoopbackPair()

	red := mustMove("h2e2")
	require.NoError(t, a.Send(red))
	assert.Equal(t, red, <-b.Incoming(), "a.Send 應於 b.Incoming 送出")

	black := mustMove("h9g7")
	require.NoError(t, b.Send(black))
	assert.Equal(t, black, <-a.Incoming(), "b.Send 應於 a.Incoming 送出")
}

// 同序輸入得同序輸出（決定性）。
func TestLoopbackPreservesOrder(t *testing.T) {
	a, b := player.NewLoopbackPair()
	seq := []string{"h2e2", "b2e2", "h0g2", "i0h0"}
	for _, s := range seq {
		require.NoError(t, a.Send(mustMove(s)))
	}
	for _, s := range seq {
		assert.Equal(t, mustMove(s), <-b.Incoming())
	}
}

// 兩側各以 RemotePlayer 代表對手，經成對回路交替走子，依序收到對手走法（無伺服器/網路）。
// 資料流：本地一步由本側 transport.Send 送出，於對側對手 RemotePlayer.RequestMove 通道收到。
func TestLoopbackDrivesTwoRemotePlayers(t *testing.T) {
	a, b := player.NewLoopbackPair()
	blackSeenByRed := player.NewRemotePlayer("黑遠端", a) // 紅側看到的黑方對手
	redSeenByBlack := player.NewRemotePlayer("紅遠端", b) // 黑側看到的紅方對手
	g := rules.NewGame()

	// 紅方本地走一步（紅側經 a.Send 送出）→ 黑側的「紅方對手」收到。
	rm := mustMove("h2e2")
	require.NoError(t, a.Send(rm))
	assert.Equal(t, rm, <-redSeenByBlack.RequestMove(g), "黑側應收到紅方走法")

	// 黑方本地回一步（黑側經 b.Send 送出）→ 紅側的「黑方對手」收到。
	bm := mustMove("h9g7")
	require.NoError(t, b.Send(bm))
	assert.Equal(t, bm, <-blackSeenByRed.RequestMove(g), "紅側應收到黑方走法")
}

// 關閉後：兩端 Incoming 通道關閉、Send 回報錯誤、重複關閉無操作（不 panic）。
func TestLoopbackCloseClosesChannelsAndErrorsSend(t *testing.T) {
	a, b := player.NewLoopbackPair()
	require.NoError(t, a.Close())

	_, ok := <-a.Incoming()
	assert.False(t, ok, "關閉後 a.Incoming 應關閉")
	_, ok = <-b.Incoming()
	assert.False(t, ok, "關閉後 b.Incoming 應關閉")

	assert.ErrorIs(t, a.Send(mustMove("h2e2")), player.ErrTransportClosed)
	assert.ErrorIs(t, b.Send(mustMove("h2e2")), player.ErrTransportClosed)

	assert.NoError(t, a.Close(), "重複關閉應為無操作")
}
