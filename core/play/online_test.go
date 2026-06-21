package play_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/core/play"
	"github.com/yuanyu90221/xiangqi-platform/player"
)

// fakeTransport 為 player.MoveTransport 的測試替身：
// Incoming 由 in 通道供應「伺服器確認的走法」；Send 記錄上送並可選擇回聲（模擬權威伺服器
// 把走法廣播回走子者本人）。
type fakeTransport struct {
	in   chan board.Move
	sent []board.Move
	echo bool
}

func newFakeTransport(echo bool) *fakeTransport {
	return &fakeTransport{in: make(chan board.Move, 16), echo: echo}
}

func (f *fakeTransport) Incoming() <-chan board.Move { return f.in }

func (f *fakeTransport) Send(m board.Move) error {
	f.sent = append(f.sent, m)
	if f.echo {
		f.in <- m
	}
	return nil
}

// 編譯期接縫：Controller 與 OnlineController 皆滿足 GUI 所依賴的 play.Driver。
var (
	_ play.Driver = (*play.Controller)(nil)
	_ play.Driver = (*play.OnlineController)(nil)
)

// 伺服器確認的走法（含對手走法）應被套用到本地 Session、推進盤面。
func TestOnlineAppliesServerConfirmedMoves(t *testing.T) {
	tr := newFakeTransport(false)
	oc, _ := play.NewOnlineController("紅", "黑", board.Black, tr) // 人執黑，對手（紅）先走

	tr.in <- uci(t, "h2e2") // 伺服器確認對手（紅）走法
	require.True(t, oc.Step(), "Step 應套用伺服器確認的走法")

	assert.Equal(t, []string{"h2e2"}, oc.Session().Record().Moves)
	assert.Equal(t, board.Black, oc.Turn(), "套用紅方走法後換黑（人類）")
}

// 本地走法只上送、不立即套用；待伺服器回聲才推進盤面（權威性、避免回聲重複套用）。
func TestOnlineLocalMoveSentNotAppliedUntilEcho(t *testing.T) {
	tr := newFakeTransport(false)                              // 先不自動回聲，手動控制時序
	oc, h := play.NewOnlineController("紅", "黑", board.Red, tr) // 人執紅，先手

	require.False(t, oc.Step(), "開局無確認走法可套用")
	iv, ok := oc.CurrentInteractive()
	require.True(t, ok, "紅（人類）回合應可互動")
	require.Equal(t, player.Interactive(h), iv)

	require.Equal(t, player.TapSelected, h.Tap(sq(t, "h2")))
	require.Equal(t, player.TapMoved, h.Tap(sq(t, "e2")))

	require.False(t, oc.Step(), "本地走法只上送，不應立即套用")
	assert.Equal(t, []string{"h2e2"}, movesOf(tr.sent), "走法應已上送伺服器")
	assert.Empty(t, oc.Session().Record().Moves, "尚未收到回聲，盤面不動")
	assert.Equal(t, board.Red, oc.Turn(), "未套用前仍為紅方回合")
	_, ok = oc.CurrentInteractive()
	assert.False(t, ok, "等待伺服器確認期間不接受再次輸入")

	tr.in <- uci(t, "h2e2") // 伺服器回聲確認本地走法
	require.True(t, oc.Step(), "收到回聲後套用本地走法")
	assert.Equal(t, []string{"h2e2"}, oc.Session().Record().Moves)
	assert.Equal(t, board.Black, oc.Turn(), "套用後換黑（對手）")
}

// 互動性僅在「己方回合且未等待確認且未終局」時開放。
func TestOnlineInteractiveOnlyOnHumanTurn(t *testing.T) {
	tr := newFakeTransport(true) // Send 自動回聲，模擬伺服器確認
	oc, h := play.NewOnlineController("紅", "黑", board.Red, tr)

	require.False(t, oc.Step())
	_, ok := oc.CurrentInteractive()
	require.True(t, ok, "紅（人類）回合可互動")

	// 人類走一手 → 上送並（回聲）套用 → 換對手（黑）回合。
	require.Equal(t, player.TapSelected, h.Tap(sq(t, "h2")))
	require.Equal(t, player.TapMoved, h.Tap(sq(t, "e2")))
	oc.Step() // 讀本地走法、上送、回聲入列
	oc.Step() // 套用回聲
	assert.Equal(t, board.Black, oc.Turn())

	_, ok = oc.CurrentInteractive()
	assert.False(t, ok, "對手回合不可互動")
}

func movesOf(ms []board.Move) []string {
	out := make([]string, len(ms))
	for i, m := range ms {
		out[i] = m.String()
	}
	return out
}
