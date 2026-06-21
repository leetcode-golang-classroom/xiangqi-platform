package play_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/core/play"
	"github.com/yuanyu90221/xiangqi-platform/player"
)

func TestVsComputerHumanPlaysRed(t *testing.T) {
	c, h := play.VsComputer("我", "電腦", board.Red, player.NewAI(player.Easy))

	// 紅（人類）回合：Step 武裝人類，等待點擊。
	require.False(t, c.Step())
	iv, ok := c.CurrentInteractive()
	require.True(t, ok, "開局應為人類（紅）回合")
	require.Equal(t, player.Interactive(h), iv)

	require.Equal(t, player.TapSelected, h.Tap(sq(t, "h2")))
	require.Equal(t, player.TapMoved, h.Tap(sq(t, "e2")))
	require.True(t, c.Step(), "人類送出後套用")
	assert.Equal(t, []string{"h2e2"}, c.Session().Record().Moves)

	// 黑（AI）回合：自動推進。
	_, ok = c.CurrentInteractive()
	assert.False(t, ok, "AI 回合非互動式")
	require.True(t, driveOneMove(c), "AI 自動走一步")
	assert.Len(t, c.Session().Record().Moves, 2)
	assert.Equal(t, board.Red, c.Turn(), "AI 走完換回紅")
}

func TestVsComputerHumanPlaysBlack(t *testing.T) {
	c, h := play.VsComputer("電腦", "我", board.Black, player.NewAI(player.Easy))

	// 開局輪紅＝AI：人類（黑）此時非互動回合，AI 先走。
	_, ok := c.CurrentInteractive()
	assert.False(t, ok, "開局為 AI（紅）回合")
	require.True(t, driveOneMove(c), "AI（紅）先走一步")
	assert.Equal(t, board.Black, c.Turn(), "換黑（人類）")

	// 黑（人類）回合：可選子走子。
	require.False(t, c.Step(), "武裝人類")
	iv, ok := c.CurrentInteractive()
	require.True(t, ok, "現為人類（黑）回合")
	require.Equal(t, player.Interactive(h), iv)
	require.Equal(t, player.TapSelected, h.Tap(sq(t, "h9")))
	require.Equal(t, player.TapMoved, h.Tap(sq(t, "g7")))
	require.True(t, c.Step())
	assert.Len(t, c.Session().Record().Moves, 2, "AI 一步 + 人類一步")
}

func TestAIvsAIPlaysLegalMoves(t *testing.T) {
	s := play.NewSession("紅機", "黑機")
	c := play.NewController(s, player.NewAI(player.Easy), player.NewAI(player.Easy))

	for i := 0; i < 6 && !c.Outcome().Over; i++ {
		before := c.Current()
		require.True(t, driveOneMove(c))
		moves := c.Session().Record().Moves
		last := moves[len(moves)-1]
		legal := false
		for _, lm := range before.LegalMoves() {
			if lm.String() == last {
				legal = true
				break
			}
		}
		assert.True(t, legal, "AI 之著須為合法走法")
	}
	assert.GreaterOrEqual(t, len(c.Session().Record().Moves), 1)
}
