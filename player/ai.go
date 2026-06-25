package player

import (
	"math/rand"
	"slices"
	"strings"

	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/core/rules"
)

// 難度：以搜尋深度表示，深度越高棋力越強。
// 棋力主要來自評估（位置價值表）與靜默搜尋，而非加深——故深度維持輕量以確保 GUI 反應。
const (
	Easy   = 1
	Medium = 2
	Hard   = 3
)

const (
	mateScore        = 1_000_000 // 將死分值（含步數修正，偏好較快將死）
	infScore         = 1 << 30   // alpha-beta 邊界（可安全取負）
	qMaxPly          = 64        // 靜默搜尋遞迴上限（保險，吃子序列本就收斂）
	nearBestEps      = 16        // near-best 容差：與最佳分相差 ≤ 此值者視為近佳手（< 兵值 100）
	maxCheckExt      = 1         // 將軍延伸累計上限：防止搜尋樹無限膨脹
	killerSlots      = 2         // 每個 ply 保留的 killer move 數量
	killerMaxPly     = 128       // killer 表最大深度（含延伸後的 ply 上限）
)

// pieceValue 為各棋子的子力價值（以 kind 的小寫位元組索引）。
// 帥/將不計值——其安危由將死終局判定處理。
var pieceValue = map[byte]int{
	'r': 900, // 車
	'n': 400, // 馬
	'c': 450, // 炮
	'b': 200, // 相/象
	'a': 200, // 仕/士
	'p': 100, // 兵/卒
	'k': 0,   // 帥/將
}

// pst 為各棋子的位置價值表（piece-square table），以**紅方視角**定義：
// 索引 [rank][file]，rank 0 為紅方底線、rank 9 為黑方底線；數值遠小於子力（約 ±2～30），
// 僅在子力相等時提供策略偏好，不致使「得子/將死」判斷反轉。黑方查表時鏡射 rank。
// 表格皆左右對稱（file f ↔ 8-f），確保開局對稱著手同分、保有變招多樣性。
// 相/象、仕/士（落點受限）與帥/將（安危由終局判定）不設表，視為 0。
var pst = map[byte][board.Ranks][board.Files]int{
	'p': { // 兵：鼓勵過河與向前推進、略偏中路
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{2, 2, 2, 4, 6, 4, 2, 2, 2},
		{10, 10, 12, 16, 18, 16, 12, 10, 10},
		{14, 14, 16, 20, 22, 20, 16, 14, 14},
		{18, 18, 20, 24, 26, 24, 20, 18, 18},
		{16, 16, 18, 20, 22, 20, 18, 16, 16},
		{12, 12, 14, 16, 16, 16, 14, 12, 12},
	},
	'n': { // 馬：偏好中央、前進；邊角與底線扣分
		{-6, -4, -2, -2, -2, -2, -2, -4, -6},
		{-4, 0, 2, 2, 2, 2, 2, 0, -4},
		{-2, 2, 4, 6, 6, 6, 4, 2, -2},
		{-2, 4, 6, 8, 8, 8, 6, 4, -2},
		{0, 4, 8, 10, 10, 10, 8, 4, 0},
		{0, 4, 8, 10, 12, 10, 8, 4, 0},
		{0, 4, 8, 10, 10, 10, 8, 4, 0},
		{-2, 2, 6, 8, 8, 8, 6, 2, -2},
		{-2, 0, 4, 6, 6, 6, 4, 0, -2},
		{-4, -2, 0, 2, 2, 2, 0, -2, -4},
	},
	'c': { // 炮：偏好中路、適度前進
		{0, 0, 2, 2, 4, 2, 2, 0, 0},
		{0, 2, 2, 4, 4, 4, 2, 2, 0},
		{2, 2, 4, 6, 6, 6, 4, 2, 2},
		{2, 4, 6, 6, 8, 6, 6, 4, 2},
		{2, 4, 6, 6, 6, 6, 6, 4, 2},
		{2, 4, 4, 6, 6, 6, 4, 4, 2},
		{0, 2, 4, 4, 6, 4, 4, 2, 0},
		{0, 2, 2, 4, 4, 4, 2, 2, 0},
		{0, 0, 2, 2, 2, 2, 2, 0, 0},
		{0, 0, 0, 2, 2, 2, 0, 0, 0},
	},
	'r': { // 車：偏好中路縱線與前進
		{0, 2, 4, 6, 6, 6, 4, 2, 0},
		{2, 4, 6, 6, 6, 6, 6, 4, 2},
		{2, 4, 6, 8, 8, 8, 6, 4, 2},
		{4, 6, 8, 8, 8, 8, 8, 6, 4},
		{4, 6, 8, 10, 10, 10, 8, 6, 4},
		{6, 8, 10, 10, 12, 10, 10, 8, 6},
		{6, 8, 10, 10, 10, 10, 10, 8, 6},
		{6, 8, 8, 10, 10, 10, 8, 8, 6},
		{4, 6, 8, 8, 8, 8, 8, 6, 4},
		{4, 6, 6, 8, 8, 8, 6, 6, 4},
	},
}

// pstValue 回傳某棋子在某格的位置價值（黑方鏡射 rank）。
func pstValue(p board.Piece, sq board.Square) int {
	t, ok := pst[p.Kind()]
	if !ok {
		return 0
	}
	file, rank := sq.File(), sq.Rank()
	if p.Color() == board.Black {
		rank = board.Ranks - 1 - rank
	}
	return t[rank][file]
}

// AI 以 negamax + alpha-beta 搜尋實作 Player。
type AI struct {
	Depth   int                          // 搜尋深度（難度）
	visits  map[string]int               // 對局中各盤面（盤面+輪走方）的造訪次數，供重複局面變招
	pick    func(n int) int              // 從 n 個近佳手中挑一個的索引；nil 時用 visits 輪替（可重現）
	killers [killerMaxPly][killerSlots]board.Move // killer move 表：非吃子卻造成 beta cutoff 的著手
}

// NewAI 以難度（搜尋深度）建立 AI；深度至少為 1。
// 預設為可重現選步（近佳手依造訪次數輪替）；若要每局變化開局，呼叫 Seed。
func NewAI(difficulty int) *AI {
	if difficulty < 1 {
		difficulty = 1
	}
	return &AI{Depth: difficulty}
}

// Seed 啟用隨機選步：於近佳手間以給定種子隨機挑選，使開局與後續對應每局不同
// （僅在最佳分 ε 容差內的著手間挑選，不致弱化棋力）。回傳自身以便串接。
func (a *AI) Seed(seed int64) *AI {
	a.pick = rand.New(rand.NewSource(seed)).Intn
	return a
}

// Name 回傳對手名稱，內含難度標籤（如「電腦（普通）」），供 GUI／棋譜辨識棋力。
func (a *AI) Name() string { return "電腦（" + difficultyLabel(a.Depth) + "）" }

// difficultyLabel 將搜尋深度轉成中文難度標籤。
func difficultyLabel(depth int) string {
	switch {
	case depth <= Easy:
		return "簡單"
	case depth >= Hard:
		return "困難"
	default:
		return "普通"
	}
}

// RequestMove 非同步取步：於背景 goroutine 搜尋，完成後將走法送入通道。
// 滿足 play.Player 介面。僅於對局未結束時呼叫（否則搜尋無著手、通道不送出）。
func (a *AI) RequestMove(g *rules.Game) <-chan board.Move {
	ch := make(chan board.Move, 1)
	go func() {
		// 每次搜尋前清除 killer 表，避免上回搜尋殘留干擾新局面
		a.killers = [killerMaxPly][killerSlots]board.Move{}
		if m, err := a.SelectMove(g); err == nil {
			ch <- m
		}
	}()
	return ch
}

// SelectMove 選出走法：以 negamax 算出各根節點著手分值，蒐集與最佳分相差 ≤ ε 的近佳手
// （依分數降冪、UCCI 升冪排序，故首位為真正最佳手）。未啟用隨機時於同一對局重訪同一盤面
// 在近佳手間依造訪次數輪替（全新 AI 首次選步可重現）；啟用隨機（Seed）時於近佳手間隨機。
func (a *AI) SelectMove(g *rules.Game) (board.Move, error) {
	if g.Result().Over {
		return board.Move{}, ErrGameOver
	}

	moves := g.LegalMoves()
	orderMoves(g, moves) // 好手優先 → 早早建立高 best，使其餘根節點著手被窗下界剪枝

	type scored struct {
		m board.Move
		s int
	}
	var cand []scored
	best := -infScore
	for _, m := range moves {
		ng, err := g.ApplyMove(m)
		if err != nil {
			continue
		}
		// 僅關心分數高於 (best-ε) 的著手：以 beta=-(best-ε) 讓明顯較差者在子樹內提早剪枝。
		lower := best - nearBestEps
		if best == -infScore {
			lower = -infScore
		}
		s := -a.negamax(ng, a.Depth-1, -infScore, -lower, 1, 0)
		if s > best {
			best = s
		}
		if s > lower { // 分數高於當時門檻者為近佳候選（其值為精確值）
			cand = append(cand, scored{m, s})
		}
	}

	// 以最終 best 過濾近佳手（與最佳分相差 ≤ ε），排序：分數降冪、UCCI 升冪 → 首位為最佳手。
	var list []scored
	for _, e := range cand {
		if e.s >= best-nearBestEps {
			list = append(list, e)
		}
	}
	slices.SortFunc(list, func(x, y scored) int {
		if x.s != y.s {
			return y.s - x.s
		}
		return strings.Compare(x.m.String(), y.m.String())
	})
	bestMoves := make([]board.Move, len(list))
	for i, e := range list {
		bestMoves[i] = e.m
	}

	// 已啟用隨機（Seed）→ 於近佳手間隨機挑選，使每局開局與應對皆有變化。
	if a.pick != nil {
		return bestMoves[a.pick(len(bestMoves))], nil
	}

	// 否則：重訪同一盤面時於近佳手間依造訪次數輪替（可重現、可破解迴圈；首次取最佳手）。
	if a.visits == nil {
		a.visits = make(map[string]int)
	}
	key := posKey(g)
	idx := a.visits[key] % len(bestMoves)
	a.visits[key]++
	return bestMoves[idx], nil
}

// posKey 為盤面 + 輪走方（忽略計步）的局面鍵，用於偵測重複局面。
func posKey(g *rules.Game) string {
	parts := strings.Fields(g.ToFEN())
	if len(parts) >= 2 {
		return parts[0] + " " + parts[1]
	}
	return g.ToFEN()
}

// negamax 以行棋方視角回傳盤面評分（含 alpha-beta 剪枝、將軍延伸、killer moves）。
// extensions 為本路徑已累計的將軍延伸次數，防止無限膨脹。
func (a *AI) negamax(g *rules.Game, depth, alpha, beta, ply, extensions int) int {
	// 自然限著（不需要 LegalMoves）：無吃子 120 半步判和。
	if g.AtNaturalLimit() {
		return 0
	}
	// 呼叫 LegalMoves 一次，同時用於終局偵測與著手迭代。
	moves := g.LegalMoves()
	if len(moves) == 0 {
		if g.InCheck() {
			return -mateScore + ply // 將死：越快越糟
		}
		return 0 // 困斃（和棋）
	}
	if depth == 0 {
		return a.quiesce(g, alpha, beta, ply)
	}
	a.orderMovesKiller(g, moves, ply)
	best := -infScore
	for _, m := range moves {
		ng, err := g.ApplyMove(m)
		if err != nil {
			continue
		}
		// 將軍延伸：走後對方被將軍 → depth+1，但整條路徑總延伸次數不超過 maxCheckExt。
		ext := 0
		if extensions < maxCheckExt && ng.InCheck() {
			ext = 1
		}
		score := -a.negamax(ng, depth-1+ext, -beta, -alpha, ply+1, extensions+ext)
		if score > best {
			best = score
		}
		if best > alpha {
			alpha = best
		}
		if alpha >= beta {
			// Beta 剪枝：若為非吃子著手，記入 killer 表供兄弟節點優先嘗試。
			if g.PieceAt(m.To).IsEmpty() && ply < killerMaxPly {
				a.killers[ply][1] = a.killers[ply][0]
				a.killers[ply][0] = m
			}
			break
		}
	}
	return best
}

// quiesce 為靜默搜尋：在搜尋深度上限沿吃子序列延伸至安定再評估，消除視界效應。
// 以 stand-pat（靜態評估）為下限；只展開吃子，確保序列嚴格收斂、不爆搜尋量。
func (a *AI) quiesce(g *rules.Game, alpha, beta, ply int) int {
	// 呼叫一次 LegalMoves，同時用於終局偵測與吃子篩選，避免重複計算。
	all := g.LegalMoves()
	if len(all) == 0 {
		if g.InCheck() {
			return -mateScore + ply
		}
		return 0 // 困斃（和棋）
	}
	stand := evaluate(g)
	if ply >= qMaxPly {
		return stand
	}
	if stand >= beta {
		return beta
	}
	if stand > alpha {
		alpha = stand
	}
	for _, m := range all {
		if g.PieceAt(m.To).IsEmpty() {
			continue // 只搜吃子著手
		}
		ng, err := g.ApplyMove(m)
		if err != nil {
			continue
		}
		score := -a.quiesce(ng, -beta, -alpha, ply+1)
		if score >= beta {
			return beta
		}
		if score > alpha {
			alpha = score
		}
	}
	return alpha
}

// orderMoves 就地排序：吃子優先（MVV-LVA），其餘按 UCCI 排序以維持可重現。
func orderMoves(g *rules.Game, moves []board.Move) {
	slices.SortFunc(moves, func(x, y board.Move) int {
		if kx, ky := moveOrderKey(g, x), moveOrderKey(g, y); kx != ky {
			return ky - kx
		}
		return strings.Compare(x.String(), y.String())
	})
}

// orderMovesKiller 就地排序：吃子（MVV-LVA）→ killer moves → 其餘（UCCI 序）。
// killer moves 讓造成過 beta cutoff 的非吃子著手在兄弟節點提前被嘗試，改善剪枝效率。
func (a *AI) orderMovesKiller(g *rules.Game, moves []board.Move, ply int) {
	slices.SortFunc(moves, func(x, y board.Move) int {
		kx := a.killerOrderKey(g, x, ply)
		ky := a.killerOrderKey(g, y, ply)
		if kx != ky {
			return ky - kx
		}
		return strings.Compare(x.String(), y.String())
	})
}

// killerOrderKey 回傳含 killer 優先級的排序鍵：
// 吃子（MVV-LVA 加大偏移）> killer[0] > killer[1] > 其他非吃子。
func (a *AI) killerOrderKey(g *rules.Game, m board.Move, ply int) int {
	if !g.PieceAt(m.To).IsEmpty() {
		return moveOrderKey(g, m) + 100_000 // 吃子永遠排在非吃子前
	}
	if ply < killerMaxPly {
		if m == a.killers[ply][0] {
			return 50_000
		}
		if m == a.killers[ply][1] {
			return 49_999
		}
	}
	return 0
}

// moveOrderKey 為走法排序鍵：吃子為「受吃子值×16 − 攻擊子值」（MVV-LVA），非吃子為 0。
func moveOrderKey(g *rules.Game, m board.Move) int {
	victim := g.PieceAt(m.To)
	if victim.IsEmpty() {
		return 0
	}
	return pieceValue[victim.Kind()]*16 - pieceValue[g.PieceAt(m.From).Kind()]
}

// evaluate 以行棋方視角回傳靜態評分（子力差 + 過河兵加成 + 位置價值）。
func evaluate(g *rules.Game) int {
	red, black := 0, 0
	for s := board.Square(0); s < board.NumSquares; s++ {
		p := g.PieceAt(s)
		if p.IsEmpty() {
			continue
		}
		v := pieceValue[p.Kind()]
		if p.Kind() == 'p' {
			v += pawnBonus(p.Color(), s)
		}
		v += pstValue(p, s)
		if p.Color() == board.Red {
			red += v
		} else {
			black += v
		}
	}
	score := red - black
	if g.Turn() == board.Black {
		score = -score
	}
	return score
}

// pawnBonus 為過河兵加成：兵卒過河後威脅大增。
func pawnBonus(c board.Color, sq board.Square) int {
	rank := sq.Rank()
	if c == board.Red && rank >= 5 { // 紅兵過河（rank 5 起）
		return 100
	}
	if c == board.Black && rank <= 4 { // 黑卒過河（rank 4 止）
		return 100
	}
	return 0
}
