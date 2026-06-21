package server_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuanyu90221/xiangqi-platform/server"
)

func joinRaw(t *testing.T, gameID string) []byte {
	t.Helper()
	raw, err := server.Encode(server.TypeJoin, gameID, nil)
	require.NoError(t, err)
	return raw
}

// 兩客戶端配對成局：建一房、指派紅/黑、各回 matched（含初始 FEN）。
func TestHubMatchesTwoClients(t *testing.T) {
	h := server.NewHub()
	a, b := &fakeConn{}, &fakeConn{}

	require.NoError(t, h.Handle(a, joinRaw(t, ""))) // 先到，等待
	assert.Empty(t, a.ofType(server.TypeMatched), "第一人應尚未配對")
	require.NoError(t, h.Handle(b, joinRaw(t, ""))) // 後到，配成

	ma := a.ofType(server.TypeMatched)
	mb := b.ofType(server.TypeMatched)
	require.Len(t, ma, 1)
	require.Len(t, mb, 1)

	gameID := ma[0].GameID
	assert.NotEmpty(t, gameID)
	assert.Equal(t, gameID, mb[0].GameID, "雙方應在同一房")

	var pa, pb server.MatchedPayload
	require.NoError(t, server.DecodePayload(ma[0], &pa))
	require.NoError(t, server.DecodePayload(mb[0], &pb))
	assert.NotEqual(t, pa.Color, pb.Color, "雙方執色應相異")
	assert.Contains(t, []string{"red", "black"}, pa.Color)
	assert.NotEmpty(t, pa.InitialFen)
	assert.Equal(t, pa.InitialFen, pb.InitialFen)
}

// 斷線後以 gameId 重連：回 state_sync（目前 FEN 與走法序列），狀態保留。
func TestHubReconnectRestoresState(t *testing.T) {
	h := server.NewHub()
	a, b := &fakeConn{}, &fakeConn{}
	require.NoError(t, h.Handle(a, joinRaw(t, "")))
	require.NoError(t, h.Handle(b, joinRaw(t, "")))

	gameID := a.ofType(server.TypeMatched)[0].GameID
	var pa server.MatchedPayload
	require.NoError(t, server.DecodePayload(a.ofType(server.TypeMatched)[0], &pa))

	// 找出紅方並走一步，建立可還原的歷史。
	red, _ := a, b
	if pa.Color != "red" {
		red, _ = b, a
	}
	moveRaw, err := server.Encode(server.TypeMove, gameID, server.MovePayload{Move: "h2e2"})
	require.NoError(t, err)
	require.NoError(t, h.Handle(red, moveRaw))

	// 紅方斷線後以原 gameId 重連。
	h.Disconnect(red)
	rc := &fakeConn{}
	require.NoError(t, h.Handle(rc, joinRaw(t, gameID)))

	ss := rc.ofType(server.TypeStateSync)
	require.Len(t, ss, 1, "重連應收到 state_sync")
	var p server.StateSyncPayload
	require.NoError(t, server.DecodePayload(ss[0], &p))
	assert.NotEmpty(t, p.Fen)
	assert.Equal(t, []string{"h2e2"}, p.Moves, "走法序列應保留")
}
