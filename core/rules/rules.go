// Package rules 為象棋規則引擎與對局狀態（RuleEngine / Game）。
//
// 職責：產生合法走法、判斷合法性、判定勝負（將死/困斃）。
// ApplyMove 採不可變語意：回傳新的 Game，不改動原狀態。
//
// 走法產生分兩層：先產生擬合法走法（走子規則 + 蹩馬腿/塞象眼/炮架），
// 再過濾「走後使己方被將軍（含飛將）」者。
package rules

import (
	"fmt"
	"slices"
	"sync"

	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/core/notation"
)

// boardPool 複用 Board 物件，避免 LegalMoves 每個擬合法著手都做 heap 分配。
var boardPool = sync.Pool{New: func() any { return &board.Board{} }}

const startposFEN = "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"

// Result 為對局結果。Winner 為 "red"/"black"，和局或未結束為空字串。
type Result struct {
	Over   bool   `json:"over"`
	Winner string `json:"winner"`
	Reason string `json:"reason"`
}

// Game 為一局的權威狀態。
type Game struct {
	board    *board.Board
	turn     board.Color
	halfmove int
	fullmove int
}

// NewGame 建立開局狀態。
func NewGame() *Game {
	g, _ := FromFEN(startposFEN)
	return g
}

// FromFEN 由 FEN 還原對局狀態。
func FromFEN(fen string) (*Game, error) {
	b, turn, half, full, err := notation.ParseFEN(fen)
	if err != nil {
		return nil, err
	}
	return &Game{board: b, turn: turn, halfmove: half, fullmove: full}, nil
}

// ToFEN 將對局狀態輸出為 FEN。
func (g *Game) ToFEN() string {
	return notation.EncodeFEN(g.board, g.turn, g.halfmove, g.fullmove)
}

// Turn 回傳輪到走子的一方。
func (g *Game) Turn() board.Color { return g.turn }

// PieceAt 回傳指定棋格上的棋子（空格為 board.Empty）。唯讀，不影響不可變性。
func (g *Game) PieceAt(sq board.Square) board.Piece { return g.board.Get(sq) }

// LegalMoves 產生當前所有合法走法。
func (g *Game) LegalMoves() []board.Move {
	if g.board == nil {
		return nil
	}
	// 從 pool 借一塊 Board 作為走子暫存，避免每個擬合法著手都做 heap 分配。
	tmp := boardPool.Get().(*board.Board)
	defer boardPool.Put(tmp)

	var moves []board.Move
	for s := board.Square(0); s < board.NumSquares; s++ {
		p := g.board.Get(s)
		if p.IsEmpty() || p.Color() != g.turn {
			continue
		}
		for _, to := range pseudoTargets(g.board, s) {
			*tmp = *g.board // 複製 90 bytes（無 heap 分配）
			tmp.Set(to, p)
			tmp.Set(s, board.Empty)
			if !inCheck(tmp, g.turn) {
				moves = append(moves, board.Move{From: s, To: to})
			}
		}
	}
	return moves
}

// IsLegal 判斷一步是否合法。
func (g *Game) IsLegal(m board.Move) bool {
	return slices.Contains(g.LegalMoves(), m)
}

// ApplyMove 套用一步合法走法，回傳新狀態（不可變）。
func (g *Game) ApplyMove(m board.Move) (*Game, error) {
	if !g.IsLegal(m) {
		return nil, fmt.Errorf("rules: 非法走法 %s", m)
	}
	captured := !g.board.Get(m.To).IsEmpty()
	isPawn := g.board.Get(m.From).Kind() == 'p'

	ng := &Game{
		board:    applied(g.board, m),
		turn:     g.turn.Opposite(),
		halfmove: g.halfmove + 1,
		fullmove: g.fullmove,
	}
	if captured || isPawn {
		ng.halfmove = 0
	}
	if g.turn == board.Black {
		ng.fullmove = g.fullmove + 1
	}
	return ng, nil
}

// NaturalLimitPlies 為自然限著上限（半回合數）：連續這麼多步無吃子即判和。
const NaturalLimitPlies = 120

// Result 判定當前對局結果。
func (g *Game) Result() Result {
	if g.board == nil {
		return Result{}
	}
	if len(g.LegalMoves()) == 0 {
		// 無合法走法：被將軍為將死，否則為困斃；兩者皆由輪走方判負。
		winner := g.turn.Opposite().String()
		if inCheck(g.board, g.turn) {
			return Result{Over: true, Winner: winner, Reason: "checkmate"}
		}
		return Result{Over: true, Winner: winner, Reason: "stalemate"}
	}
	if g.halfmove >= NaturalLimitPlies {
		return Result{Over: true, Reason: "natural_limit"} // 和棋，無勝方
	}
	return Result{Over: false}
}

// InCheck 回傳目前輪走方是否處於被將軍狀態。
func (g *Game) InCheck() bool {
	return inCheck(g.board, g.turn)
}

// AtNaturalLimit 回傳是否已達無吃子限著上限（和棋），不需呼叫 LegalMoves。
func (g *Game) AtNaturalLimit() bool {
	return g.halfmove >= NaturalLimitPlies
}

// PseudoTargets 回傳指定格的擬合法目的格（棋子移動規則，不過濾自將）。供測試用。
func (g *Game) PseudoTargets(sq board.Square) []board.Square {
	if g.board == nil {
		return nil
	}
	return pseudoTargets(g.board, sq)
}

// positionKey 回傳僅含盤面與輪走方的鍵（忽略計步），供重複局面偵測。
func (g *Game) positionKey() string {
	return notation.EncodeFEN(g.board, g.turn, 0, 0)
}

// ToChinese 將一步走法轉成中文記譜（顯示用）。尚未實作。
func (g *Game) ToChinese(m board.Move) string {
	return notation.ToChinese(g.board, m, g.turn)
}

// applied 回傳套用走法後的新盤面（不改動原盤面）。
func applied(b *board.Board, m board.Move) *board.Board {
	nb := b.Clone()
	nb.Set(m.To, nb.Get(m.From))
	nb.Set(m.From, board.Empty)
	return nb
}
