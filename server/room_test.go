package server_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuanyu90221/xiangqi-platform/core/rules"
	"github.com/yuanyu90221/xiangqi-platform/server"
)

// fakeConn 為 server.Conn 的測試替身，記錄伺服器送來的 envelope。
type fakeConn struct {
	mu   sync.Mutex
	sent []server.Envelope
}

func (c *fakeConn) Send(env server.Envelope) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sent = append(c.sent, env)
	return nil
}

func (c *fakeConn) all() []server.Envelope {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]server.Envelope(nil), c.sent...)
}

func (c *fakeConn) ofType(t server.MsgType) []server.Envelope {
	var out []server.Envelope
	for _, e := range c.all() {
		if e.Type == t {
			out = append(out, e)
		}
	}
	return out
}

var _ server.Conn = (*fakeConn)(nil)

// 合法走子：套用並向雙方廣播 move_applied（盤面／輪走方更新）。
func TestRoomLegalMoveBroadcastsBoth(t *testing.T) {
	red, black := &fakeConn{}, &fakeConn{}
	r := server.NewRoom("g-1", rules.NewGame(), red, black)

	r.HandleMove(red, "h2e2") // 紅炮平，合法

	for name, c := range map[string]*fakeConn{"red": red, "black": black} {
		ma := c.ofType(server.TypeMoveApplied)
		require.Len(t, ma, 1, "%s 應收到一則 move_applied", name)
		assert.Equal(t, "g-1", ma[0].GameID)
		var p server.MoveAppliedPayload
		require.NoError(t, server.DecodePayload(ma[0], &p))
		assert.Equal(t, "h2e2", p.Move)
		assert.Equal(t, "black", p.Turn, "套用後應輪黑")
		assert.NotEmpty(t, p.Fen)
	}
	assert.Empty(t, red.ofType(server.TypeError), "合法走子不應有 error")
}

// 非法走子：僅回送出方 error、不廣播、狀態不變。
func TestRoomIllegalMoveErrorsSenderOnly(t *testing.T) {
	red, black := &fakeConn{}, &fakeConn{}
	r := server.NewRoom("g-1", rules.NewGame(), red, black)

	r.HandleMove(red, "b0c0") // 馬橫走一格，非法

	assert.Len(t, red.ofType(server.TypeError), 1, "送出方應收到 error")
	assert.Empty(t, red.ofType(server.TypeMoveApplied), "非法走子不應廣播")
	assert.Empty(t, black.all(), "對手不應收到任何訊息")
}

// 非當前回合方走子：回 error 並忽略，狀態不變。
func TestRoomWrongTurnRejected(t *testing.T) {
	red, black := &fakeConn{}, &fakeConn{}
	r := server.NewRoom("g-1", rules.NewGame(), red, black) // 開局輪紅

	r.HandleMove(black, "h7e7") // 黑於紅回合搶走

	assert.Len(t, black.ofType(server.TypeError), 1, "非當前回合方應收到 error")
	assert.Empty(t, black.ofType(server.TypeMoveApplied))
	assert.Empty(t, red.all(), "另一方不應收到任何訊息")
}

// 將死後：move_applied 後廣播 game_over，其後拒絕進一步走子。
func TestRoomGameOverBroadcastsAndRejectsFurther(t *testing.T) {
	g, err := rules.FromFEN("4k4/7R1/9/9/9/9/8p/9/R8/3K5 w - - 0 1")
	require.NoError(t, err)
	red, black := &fakeConn{}, &fakeConn{}
	r := server.NewRoom("g-1", g, red, black)

	r.HandleMove(red, "a1a9") // 紅車進底將死

	for name, c := range map[string]*fakeConn{"red": red, "black": black} {
		require.Len(t, c.ofType(server.TypeMoveApplied), 1, "%s 應收到 move_applied", name)
		go_ := c.ofType(server.TypeGameOver)
		require.Len(t, go_, 1, "%s 應收到 game_over", name)
		var p server.GameOverPayload
		require.NoError(t, server.DecodePayload(go_[0], &p))
		assert.Equal(t, "red", p.Result)
		assert.Equal(t, "checkmate", p.Reason)
	}

	// 局終後黑方再走應被拒、不再廣播。
	r.HandleMove(black, "i3i4")
	assert.NotEmpty(t, black.ofType(server.TypeError), "局終後走子應被拒")
	assert.Len(t, black.ofType(server.TypeMoveApplied), 1, "局終後不應再有 move_applied")
}
