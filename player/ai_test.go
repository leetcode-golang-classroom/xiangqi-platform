package player_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/core/rules"
	"github.com/yuanyu90221/xiangqi-platform/player"
)

// AI 須實作 Player 介面。
var _ player.Player = (*player.AI)(nil)

func TestAICapturesFreeRook(t *testing.T) {
	// 紅車 a0 可沿 a 線吃無保護的黑車 a5（黑僅餘將，無法反吃）。
	g, err := rules.FromFEN("5k3/9/9/9/r8/9/9/9/9/R2K5 w - - 0 1")
	require.NoError(t, err)

	ai := player.NewAI(player.Medium)
	m, err := ai.SelectMove(g)
	require.NoError(t, err)
	assert.Equal(t, "a0a5", m.String(), "應選擇吃掉無保護黑車")
}

func TestAIFindsMateInOne(t *testing.T) {
	// 紅 a1 車進到 a9 將死：a9 控第 9 線、h8 車控第 8 線，黑將 e9 無處可逃。
	// 黑另置一遠方可動卒 i3（永遠有著），排除困斃選項，使 a1a9 將死為唯一致勝著。
	g, err := rules.FromFEN("4k4/7R1/9/9/9/9/8p/9/R8/3K5 w - - 0 1")
	require.NoError(t, err)

	ai := player.NewAI(player.Medium)
	m, err := ai.SelectMove(g)
	require.NoError(t, err)

	ng, err := g.ApplyMove(m)
	require.NoError(t, err)
	res := ng.Result()
	assert.True(t, res.Over, "AI 之著應結束對局")
	assert.Equal(t, "red", res.Winner)
	assert.Equal(t, "checkmate", res.Reason)
}

func TestAIReturnsLegalMoveFromStart(t *testing.T) {
	g := rules.NewGame()
	ai := player.NewAI(player.Medium)
	m, err := ai.SelectMove(g)
	require.NoError(t, err)

	assert.True(t, slices.Contains(g.LegalMoves(), m), "AI 之著須為合法走法")
}

func TestAIErrorsOnFinishedGame(t *testing.T) {
	// 黑將 e9 被 a9 車將死，輪黑走 → 已結束。
	g, err := rules.FromFEN("R3k4/9/9/9/9/9/9/9/9/4K4 b - - 0 1")
	require.NoError(t, err)
	require.True(t, g.Result().Over)

	ai := player.NewAI(player.Easy)
	_, err = ai.SelectMove(g)
	assert.Error(t, err, "對已結束對局選步應回報錯誤")
}

func TestDifficultyMapsToDepth(t *testing.T) {
	easy := player.NewAI(player.Easy)
	med := player.NewAI(player.Medium)
	hard := player.NewAI(player.Hard)
	assert.LessOrEqual(t, easy.Depth, med.Depth)
	assert.LessOrEqual(t, med.Depth, hard.Depth)
	assert.GreaterOrEqual(t, easy.Depth, 1, "深度至少為 1")
}

func TestAINameNonEmpty(t *testing.T) {
	assert.NotEmpty(t, player.NewAI(player.Medium).Name())
}

func TestAIRequestMoveYieldsLegalMove(t *testing.T) {
	g := rules.NewGame()
	m := <-player.NewAI(player.Medium).RequestMove(g) // 背景搜尋完成後送出
	assert.True(t, slices.Contains(g.LegalMoves(), m), "RequestMove 通道應送出合法走法")
}

func TestAIVariesOnRepeatedPosition(t *testing.T) {
	// 紅帥 d0 只有 d0d1、d0e0 兩個等值（皆 0）最佳手；黑將 f9 不照面。
	g, err := rules.FromFEN("5k3/9/9/9/9/9/9/9/9/3K5 w - - 0 1")
	require.NoError(t, err)

	ai := player.NewAI(player.Easy)
	m1, err := ai.SelectMove(g)
	require.NoError(t, err)
	m2, err := ai.SelectMove(g) // 同一 AI 重訪同一盤面 → 應變招
	require.NoError(t, err)

	assert.NotEqual(t, m1.String(), m2.String(), "重複局面應給出不同等值走法")
	for _, m := range []board.Move{m1, m2} {
		assert.True(t, slices.Contains(g.LegalMoves(), m), "變招仍須合法")
	}
}

func TestAIUniqueBestUnaffectedByRepeat(t *testing.T) {
	// 唯一最佳手：白吃黑車 a0a5。連選多次皆應相同。
	g, err := rules.FromFEN("5k3/9/9/9/r8/9/9/9/9/R2K5 w - - 0 1")
	require.NoError(t, err)

	ai := player.NewAI(player.Medium)
	for i := 0; i < 3; i++ {
		m, err := ai.SelectMove(g)
		require.NoError(t, err)
		assert.Equal(t, "a0a5", m.String(), "唯一最佳手不受重訪影響")
	}
}

func TestSeededAIVariesOpening(t *testing.T) {
	// 自開局起，不同種子的 AI 應能給出不同（但皆合法）的開局著手，避免每局同一套路。
	g := rules.NewGame()
	seen := map[string]bool{}
	for _, seed := range []int64{1, 2, 3, 4, 5, 6, 7, 8} {
		m, err := player.NewAI(player.Easy).Seed(seed).SelectMove(g)
		require.NoError(t, err)
		assert.True(t, slices.Contains(g.LegalMoves(), m), "隨機開局仍須合法")
		seen[m.String()] = true
	}
	assert.Greater(t, len(seen), 1, "不同種子應產生多種開局，不應每局同一招")
}

func TestFreshAIFirstMoveReproducible(t *testing.T) {
	g := rules.NewGame()
	m1, err := player.NewAI(player.Medium).SelectMove(g)
	require.NoError(t, err)
	m2, err := player.NewAI(player.Medium).SelectMove(g)
	require.NoError(t, err)
	assert.Equal(t, m1.String(), m2.String(), "全新 AI 同盤面首手應可重現")
}

func TestAIAvoidsLosingCapture(t *testing.T) {
	// 紅車 a5 可橫吃黑兵 e5，但黑車 e9 守在 e 線，吃後必被回吃淨虧一車。
	// 以深度 1 凸顯靜默搜尋：若無 quiesce，葉節點靜態評估會誤判吃兵得分；
	// 有 quiesce 則看穿回吃 → AI 不選送子吃法。
	g, err := rules.FromFEN("4rk3/9/9/9/R3p4/9/9/9/9/3K5 w - - 0 1")
	require.NoError(t, err)
	require.True(t, slices.Contains(g.LegalMoves(), mustMove("a5e5")), "前提：a5e5 為合法吃子")

	ai := player.NewAI(1) // 深度 1：葉節點即進入靜默搜尋
	m, err := ai.SelectMove(g)
	require.NoError(t, err)
	assert.NotEqual(t, "a5e5", m.String(), "靜默搜尋應看穿回吃，避免送車吃兵")
}

func TestPSTPrefersDevelopment(t *testing.T) {
	// 等子力下，紅馬 b0 可走 a2/c2/d1；位置表以 c2（中央前進）最高 → 全新 AI 應選 b0c2。
	g, err := rules.FromFEN("3k5/9/9/9/9/9/9/9/9/1N2K4 w - - 0 1")
	require.NoError(t, err)

	m, err := player.NewAI(player.Easy).SelectMove(g)
	require.NoError(t, err)
	assert.Equal(t, "b0c2", m.String(), "位置評估應偏好較佳發展位置")
}

func TestNearBestStillUniqueOnTactics(t *testing.T) {
	// 得子（吃一車）遠超 ε 容差，即使啟用隨機也必為唯一選擇，不被 near-best 變招稀釋。
	g, err := rules.FromFEN("5k3/9/9/9/r8/9/9/9/9/R2K5 w - - 0 1")
	require.NoError(t, err)

	for _, seed := range []int64{1, 7, 42} {
		m, err := player.NewAI(player.Medium).Seed(seed).SelectMove(g)
		require.NoError(t, err)
		assert.Equal(t, "a0a5", m.String(), "戰術得子手不受 ε 容差/隨機影響")
	}
}

func TestAINameReflectsDifficulty(t *testing.T) {
	assert.Contains(t, player.NewAI(player.Easy).Name(), "簡單")
	assert.Contains(t, player.NewAI(player.Medium).Name(), "普通")
	assert.Contains(t, player.NewAI(player.Hard).Name(), "困難")
}

// mustMove 解析 UCCI 走法，測試輔助用。
func mustMove(s string) board.Move {
	m, err := board.ParseUCCI(s)
	if err != nil {
		panic(err)
	}
	return m
}

// 編譯期確認介面方法簽章與 board.Move 對應。
var _ = func() board.Move { return board.Move{} }
