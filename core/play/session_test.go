package play_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/core/play"
)

func sq(t *testing.T, s string) board.Square {
	t.Helper()
	q, err := board.ParseSquare(s)
	require.NoError(t, err)
	return q
}

func uci(t *testing.T, s string) board.Move {
	t.Helper()
	m, err := board.ParseUCCI(s)
	require.NoError(t, err)
	return m
}

func TestSessionPlayAndRecord(t *testing.T) {
	s := play.NewSession("紅", "黑")
	require.NoError(t, s.Play(uci(t, "h2e2")))
	require.NoError(t, s.Play(uci(t, "h9g7")))

	assert.Equal(t, []string{"h2e2", "h9g7"}, s.Record().Moves)
	assert.Equal(t, board.Red, s.Turn(), "兩手後輪回紅")
}

func TestSessionPlayRejectsIllegal(t *testing.T) {
	s := play.NewSession("紅", "黑")
	err := s.Play(uci(t, "e0e5")) // 帥不能跳
	require.Error(t, err)
	assert.Empty(t, s.Record().Moves, "非法走法不記錄")
}

func TestSessionUndo(t *testing.T) {
	s := play.NewSession("紅", "黑")
	require.NoError(t, s.Play(uci(t, "h2e2")))
	require.True(t, s.Undo())
	assert.Equal(t, board.Red, s.Turn())
	assert.Empty(t, s.Record().Moves)
	assert.False(t, s.Undo(), "無走法時悔棋為無操作")
}

func TestSessionResign(t *testing.T) {
	s := play.NewSession("紅", "黑")
	s.Resign() // 輪紅走，紅認輸 → 黑勝
	out := s.Outcome()
	assert.True(t, out.Over)
	assert.Equal(t, "black", out.Winner)
	assert.Equal(t, "resign", out.Reason)
	assert.Equal(t, "black_win", s.Record().Result)
}

func TestSessionNoPlayAfterGameOver(t *testing.T) {
	s := play.NewSession("紅", "黑")
	s.Resign()
	assert.ErrorIs(t, s.Play(uci(t, "h2e2")), play.ErrGameOver)
	assert.Empty(t, s.Record().Moves)
}
