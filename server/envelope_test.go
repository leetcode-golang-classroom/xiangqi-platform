package server_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuanyu90221/xiangqi-platform/server"
)

// move envelope 編碼後再解碼應等值（型別/gameId/UCCI 走法）。
func TestEnvelopeMoveRoundTrip(t *testing.T) {
	raw, err := server.Encode(server.TypeMove, "g-1", server.MovePayload{Move: "h2e2"})
	require.NoError(t, err)

	env, err := server.Decode(raw)
	require.NoError(t, err)
	assert.Equal(t, server.TypeMove, env.Type)
	assert.Equal(t, "g-1", env.GameID)

	mp, err := server.DecodeMove(env)
	require.NoError(t, err)
	assert.Equal(t, "h2e2", mp.Move)
}

// matched envelope 往返保留 color 與 initialFen。
func TestEnvelopeMatchedRoundTrip(t *testing.T) {
	const fen = "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	raw, err := server.Encode(server.TypeMatched, "g-1", server.MatchedPayload{Color: "red", InitialFen: fen})
	require.NoError(t, err)

	env, err := server.Decode(raw)
	require.NoError(t, err)
	assert.Equal(t, server.TypeMatched, env.Type)

	var mp server.MatchedPayload
	require.NoError(t, server.DecodePayload(env, &mp))
	assert.Equal(t, "red", mp.Color)
	assert.Equal(t, fen, mp.InitialFen)
}

// 未知型別解碼應回報錯誤（不靜默忽略/不 panic）。
func TestEnvelopeDecodeUnknownType(t *testing.T) {
	_, err := server.Decode([]byte(`{"type":"frobnicate","gameId":"g-1"}`))
	assert.Error(t, err, "未知 type 應回報錯誤")
}

// 損壞 JSON 應回報錯誤而非 panic。
func TestEnvelopeDecodeMalformed(t *testing.T) {
	_, err := server.Decode([]byte(`{not json`))
	assert.Error(t, err)
}
