package player_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/core/rules"
	"github.com/yuanyu90221/xiangqi-platform/player"
)

// Human 須同時滿足 player.Player 與 player.Interactive。
var (
	_ player.Player      = (*player.Human)(nil)
	_ player.Interactive = (*player.Human)(nil)
)

func psq(t *testing.T, s string) board.Square {
	t.Helper()
	q, err := board.ParseSquare(s)
	require.NoError(t, err)
	return q
}

func TestHumanSelectShowsTargets(t *testing.T) {
	h := player.NewHuman("我")
	g := rules.NewGame()
	h.RequestMove(g)

	assert.Equal(t, player.TapSelected, h.Tap(psq(t, "h2")))
	selected, ok := h.Selected()
	require.True(t, ok)
	assert.Equal(t, psq(t, "h2"), selected)

	var want []board.Square
	for _, m := range g.LegalMoves() {
		if m.From == psq(t, "h2") {
			want = append(want, m.To)
		}
	}
	assert.ElementsMatch(t, want, h.Targets())
	assert.NotEmpty(t, want)
}

func TestHumanTapTargetEmitsMove(t *testing.T) {
	h := player.NewHuman("我")
	ch := h.RequestMove(rules.NewGame())

	require.Equal(t, player.TapSelected, h.Tap(psq(t, "h2")))
	require.Equal(t, player.TapMoved, h.Tap(psq(t, "e2")))

	select {
	case m := <-ch:
		assert.Equal(t, "h2e2", m.String())
	default:
		t.Fatal("應於通道送出走法")
	}
	_, ok := h.Selected()
	assert.False(t, ok, "送出後清除選取")
}

func TestHumanReselect(t *testing.T) {
	h := player.NewHuman("我")
	h.RequestMove(rules.NewGame())
	require.Equal(t, player.TapSelected, h.Tap(psq(t, "h2")))

	assert.Equal(t, player.TapSelected, h.Tap(psq(t, "b2")))
	selected, ok := h.Selected()
	require.True(t, ok)
	assert.Equal(t, psq(t, "b2"), selected)
}

func TestHumanTapEmptyClears(t *testing.T) {
	h := player.NewHuman("我")
	h.RequestMove(rules.NewGame())
	require.Equal(t, player.TapSelected, h.Tap(psq(t, "h2")))

	assert.Equal(t, player.TapCleared, h.Tap(psq(t, "a5")))
	_, ok := h.Selected()
	assert.False(t, ok)
}

func TestHumanTapWhenNotArmedIgnored(t *testing.T) {
	h := player.NewHuman("我") // 未 RequestMove
	assert.Equal(t, player.TapIgnored, h.Tap(psq(t, "h2")))
}

func TestHumanTapOpponentIgnored(t *testing.T) {
	h := player.NewHuman("我")
	h.RequestMove(rules.NewGame()) // 開局輪紅
	assert.Equal(t, player.TapIgnored, h.Tap(psq(t, "h7")), "黑子非可選")
}
