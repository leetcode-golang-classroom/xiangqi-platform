package play_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/core/play"
	"github.com/yuanyu90221/xiangqi-platform/core/rules"
	"github.com/yuanyu90221/xiangqi-platform/player"
)

// botStub 為非互動式（已算好的）取步者：立即送出第一個合法走法。
type botStub struct{ name string }

func (b botStub) Name() string { return b.name }
func (b botStub) RequestMove(g *rules.Game) <-chan board.Move {
	ch := make(chan board.Move, 1)
	ch <- g.LegalMoves()[0]
	return ch
}

// humanStub 為互動式取步者：點擊時送出預設走法。
type humanStub struct {
	name string
	ch   chan board.Move
	move board.Move
}

func (h *humanStub) Name() string { return h.name }
func (h *humanStub) RequestMove(g *rules.Game) <-chan board.Move {
	h.ch = make(chan board.Move, 1)
	return h.ch
}
func (h *humanStub) Selected() (board.Square, bool) { return board.InvalidSquare, false }
func (h *humanStub) Targets() []board.Square        { return nil }
func (h *humanStub) Tap(board.Square) player.TapResult {
	h.ch <- h.move
	return player.TapMoved
}

var (
	_ player.Player      = botStub{}
	_ player.Interactive = (*humanStub)(nil)
)

func driveOneMove(c *play.Controller) bool {
	for !c.Outcome().Over {
		if c.Step() {
			return true
		}
	}
	return false
}

func TestControllerUnifiedLoopAppliesMoves(t *testing.T) {
	s := play.NewSession("紅", "黑")
	c := play.NewController(s, botStub{"r"}, botStub{"b"})

	for i := 0; i < 6 && !c.Outcome().Over; i++ {
		before := c.Current()
		require.True(t, driveOneMove(c), "應套用一步")
		moves := c.Session().Record().Moves
		last := moves[len(moves)-1]
		legal := false
		for _, lm := range before.LegalMoves() {
			if lm.String() == last {
				legal = true
				break
			}
		}
		assert.True(t, legal, "套用的著手須為合法走法")
	}
	assert.GreaterOrEqual(t, len(c.Session().Record().Moves), 1)
}

func TestControllerThinkingOnNonInteractiveTurn(t *testing.T) {
	c := play.NewController(play.NewSession("紅", "黑"), botStub{"r"}, botStub{"b"})
	require.False(t, c.Step(), "首次 Step 武裝取步")
	assert.True(t, c.Thinking(), "等待非互動式取步 → thinking")
	_, ok := c.CurrentInteractive()
	assert.False(t, ok, "bot 非互動式")
}

func TestControllerInteractiveTurn(t *testing.T) {
	h := &humanStub{name: "我", move: uci(t, "h2e2")}
	c := play.NewController(play.NewSession("紅", "黑"), h, botStub{"b"})

	require.False(t, c.Step(), "武裝人類回合")
	iv, ok := c.CurrentInteractive()
	require.True(t, ok, "當前應為互動式（人類）")
	assert.False(t, c.Thinking(), "人類回合不算 thinking")

	require.Equal(t, player.TapMoved, iv.Tap(sq(t, "e2")))
	require.True(t, c.Step(), "人類送出後套用")
	assert.Equal(t, []string{"h2e2"}, c.Session().Record().Moves)
}

func TestControllerUndoResetsPending(t *testing.T) {
	c := play.NewController(play.NewSession("紅", "黑"), botStub{"r"}, botStub{"b"})
	require.True(t, driveOneMove(c))
	require.Len(t, c.Session().Record().Moves, 1)

	assert.True(t, c.Undo())
	assert.Empty(t, c.Session().Record().Moves)
	assert.False(t, c.Thinking(), "悔棋後重置取步")
}

func TestControllerGameOverStops(t *testing.T) {
	c := play.NewController(play.NewSession("紅", "黑"), botStub{"r"}, botStub{"b"})
	c.Resign()
	assert.True(t, c.Outcome().Over)
	assert.False(t, c.Step(), "結束後不再推進")
	_, ok := c.CurrentInteractive()
	assert.False(t, ok)
	assert.False(t, c.Thinking())
}
