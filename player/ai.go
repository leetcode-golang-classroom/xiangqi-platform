package player

import (
	"slices"
	"strings"

	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/core/rules"
)

// 難度：以搜尋深度表示，深度越高棋力越強。
const (
	Easy   = 1
	Medium = 2
	Hard   = 3
)

const (
	mateScore = 1_000_000 // 將死分值（含步數修正，偏好較快將死）
	infScore  = 1 << 30   // alpha-beta 邊界（可安全取負）
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

// AI 以 negamax + alpha-beta 搜尋實作 Player。
type AI struct {
	Depth  int            // 搜尋深度（難度）
	visits map[string]int // 對局中各盤面（盤面+輪走方）的造訪次數，供重複局面變招
}

// NewAI 以難度（搜尋深度）建立 AI；深度至少為 1。
func NewAI(difficulty int) *AI {
	if difficulty < 1 {
		difficulty = 1
	}
	return &AI{Depth: difficulty}
}

// Name 回傳對手名稱。
func (a *AI) Name() string { return "電腦" }

// RequestMove 非同步取步：於背景 goroutine 搜尋，完成後將走法送入通道。
// 滿足 play.Player 介面。僅於對局未結束時呼叫（否則搜尋無著手、通道不送出）。
func (a *AI) RequestMove(g *rules.Game) <-chan board.Move {
	ch := make(chan board.Move, 1)
	go func() {
		if m, err := a.SelectMove(g); err == nil {
			ch <- m
		}
	}()
	return ch
}

// SelectMove 選出最佳走法。蒐集所有等值最佳手，並於同一對局中重訪同一盤面時
// 在等值最佳手間輪替（破解迴圈、增加多樣性）；唯一最佳手時不變。
// 全新 AI 於同一盤面的首次選步可重現。
func (a *AI) SelectMove(g *rules.Game) (board.Move, error) {
	if g.Result().Over {
		return board.Move{}, ErrGameOver
	}

	// 以全窗 negamax 取得各根節點走法的精確分值。
	best := -infScore
	for _, m := range g.LegalMoves() {
		ng, err := g.ApplyMove(m)
		if err != nil {
			continue
		}
		if s := -a.negamax(ng, a.Depth-1, -infScore, infScore, 1); s > best {
			best = s
		}
	}

	// 蒐集等值最佳手，依 UCCI 排序以確保可重現。
	var bestMoves []board.Move
	for _, m := range g.LegalMoves() {
		ng, err := g.ApplyMove(m)
		if err != nil {
			continue
		}
		if -a.negamax(ng, a.Depth-1, -infScore, infScore, 1) == best {
			bestMoves = append(bestMoves, m)
		}
	}
	slices.SortFunc(bestMoves, func(x, y board.Move) int {
		return strings.Compare(x.String(), y.String())
	})

	// 重訪同一盤面 → 於等值最佳手間依造訪次數輪替。
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

// negamax 以行棋方視角回傳盤面評分（含 alpha-beta 剪枝）。
func (a *AI) negamax(g *rules.Game, depth, alpha, beta, ply int) int {
	res := g.Result()
	if res.Over {
		switch res.Reason {
		case "checkmate", "stalemate":
			return -mateScore + ply // 行棋方被將死/困斃 → 必負（越快發生越糟）
		default:
			return 0 // 和棋（重複/自然限著）
		}
	}
	if depth == 0 {
		return evaluate(g)
	}
	best := -infScore
	for _, m := range g.LegalMoves() {
		ng, err := g.ApplyMove(m)
		if err != nil {
			continue
		}
		score := -a.negamax(ng, depth-1, -beta, -alpha, ply+1)
		if score > best {
			best = score
		}
		if best > alpha {
			alpha = best
		}
		if alpha >= beta {
			break // beta 剪枝
		}
	}
	return best
}

// evaluate 以行棋方視角回傳靜態評分（子力差 + 過河兵加成）。
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
