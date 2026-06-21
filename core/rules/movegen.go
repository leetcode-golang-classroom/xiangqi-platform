package rules

import (
	"slices"

	"github.com/yuanyu90221/xiangqi-platform/core/board"
)

// targetFn 為單一棋子的走法產生策略：給盤面、起點與顏色，回傳擬合法目的格
// （含蹩馬腿/塞象眼/炮架，但不過濾「走後自將」）。
type targetFn func(b *board.Board, from board.Square, color board.Color) []board.Square

// pieceTargets 為棋子種類 → 走法策略的註冊表（Strategy）。
// 新增變體棋子只需在此註冊一筆，無需改動 pseudoTargets。
var pieceTargets = map[byte]targetFn{
	'r': rookTargets,
	'c': cannonTargets,
	'n': horseTargets,
	'b': elephantTargets,
	'a': advisorTargets,
	'k': kingTargets,
	'p': pawnTargets,
}

// pseudoTargets 依棋子種類查表分派至對應策略。
func pseudoTargets(b *board.Board, from board.Square) []board.Square {
	p := b.Get(from)
	if p.IsEmpty() {
		return nil
	}
	if fn := pieceTargets[p.Kind()]; fn != nil {
		return fn(b, from, p.Color())
	}
	return nil
}

var (
	orthoDirs    = [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	diagDirs     = [][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
	elephantDirs = [][2]int{{2, 2}, {2, -2}, {-2, 2}, {-2, -2}}
)

// canLand 回報該格可否落子（在盤內且為空或敵子）。
func canLand(b *board.Board, sq board.Square, color board.Color) bool {
	if sq == board.InvalidSquare {
		return false
	}
	t := b.Get(sq)
	return t.IsEmpty() || t.Color() != color
}

// rookTargets 車：沿四正方向滑動，遇敵子可吃，遇子即止。
func rookTargets(b *board.Board, from board.Square, color board.Color) []board.Square {
	var out []board.Square
	f, r := from.File(), from.Rank()
	for _, d := range orthoDirs {
		nf, nr := f, r
		for {
			nf, nr = nf+d[0], nr+d[1]
			sq := board.MakeSquare(nf, nr)
			if sq == board.InvalidSquare {
				break
			}
			t := b.Get(sq)
			if t.IsEmpty() {
				out = append(out, sq)
				continue
			}
			if t.Color() != color {
				out = append(out, sq)
			}
			break
		}
	}
	return out
}

// cannonTargets 炮：空線直走；隔一個炮架後可吃第一個敵子。
func cannonTargets(b *board.Board, from board.Square, color board.Color) []board.Square {
	var out []board.Square
	f, r := from.File(), from.Rank()
	for _, d := range orthoDirs {
		nf, nr := f, r
		jumped := false
		for {
			nf, nr = nf+d[0], nr+d[1]
			sq := board.MakeSquare(nf, nr)
			if sq == board.InvalidSquare {
				break
			}
			t := b.Get(sq)
			if !jumped {
				if t.IsEmpty() {
					out = append(out, sq)
				} else {
					jumped = true // 找到炮架
				}
				continue
			}
			if !t.IsEmpty() {
				if t.Color() != color {
					out = append(out, sq) // 炮架後第一個敵子可吃
				}
				break
			}
		}
	}
	return out
}

// horseTargets 馬：日字 + 蹩馬腿。
func horseTargets(b *board.Board, from board.Square, color board.Color) []board.Square {
	var out []board.Square
	f, r := from.File(), from.Rank()
	cands := []struct{ df, dr, lf, lr int }{
		{1, 2, 0, 1}, {-1, 2, 0, 1}, {1, -2, 0, -1}, {-1, -2, 0, -1},
		{2, 1, 1, 0}, {2, -1, 1, 0}, {-2, 1, -1, 0}, {-2, -1, -1, 0},
	}
	for _, c := range cands {
		leg := board.MakeSquare(f+c.lf, r+c.lr)
		if leg == board.InvalidSquare || !b.Get(leg).IsEmpty() {
			continue // 蹩馬腿
		}
		if to := board.MakeSquare(f+c.df, r+c.dr); canLand(b, to, color) {
			out = append(out, to)
		}
	}
	return out
}

// elephantTargets 相象：田字、不過河、塞象眼。
func elephantTargets(b *board.Board, from board.Square, color board.Color) []board.Square {
	var out []board.Square
	f, r := from.File(), from.Rank()
	for _, d := range elephantDirs {
		nr := r + d[1]
		if color == board.Red && nr > 4 {
			continue
		}
		if color == board.Black && nr < 5 {
			continue
		}
		eye := board.MakeSquare(f+d[0]/2, r+d[1]/2)
		if eye == board.InvalidSquare || !b.Get(eye).IsEmpty() {
			continue // 塞象眼
		}
		if to := board.MakeSquare(f+d[0], nr); canLand(b, to, color) {
			out = append(out, to)
		}
	}
	return out
}

// advisorTargets 仕士：斜一格、不出宮。
func advisorTargets(b *board.Board, from board.Square, color board.Color) []board.Square {
	var out []board.Square
	f, r := from.File(), from.Rank()
	for _, d := range diagDirs {
		nf, nr := f+d[0], r+d[1]
		if inPalace(color, nf, nr) && canLand(b, board.MakeSquare(nf, nr), color) {
			out = append(out, board.MakeSquare(nf, nr))
		}
	}
	return out
}

// kingTargets 帥將：直一格、不出宮（飛將另由 kingsFacing 處理）。
func kingTargets(b *board.Board, from board.Square, color board.Color) []board.Square {
	var out []board.Square
	f, r := from.File(), from.Rank()
	for _, d := range orthoDirs {
		nf, nr := f+d[0], r+d[1]
		if inPalace(color, nf, nr) && canLand(b, board.MakeSquare(nf, nr), color) {
			out = append(out, board.MakeSquare(nf, nr))
		}
	}
	return out
}

// pawnTargets 兵卒：未過河僅前進；過河後可左右。
func pawnTargets(b *board.Board, from board.Square, color board.Color) []board.Square {
	var out []board.Square
	f, r := from.File(), from.Rank()
	forward := 1
	if color == board.Black {
		forward = -1
	}
	if fwd := board.MakeSquare(f, r+forward); canLand(b, fwd, color) {
		out = append(out, fwd)
	}
	crossed := (color == board.Red && r >= 5) || (color == board.Black && r <= 4)
	if crossed {
		for _, df := range []int{1, -1} {
			if s := board.MakeSquare(f+df, r); canLand(b, s, color) {
				out = append(out, s)
			}
		}
	}
	return out
}

// inPalace 回報座標是否在該方九宮內。
func inPalace(c board.Color, f, r int) bool {
	if f < 3 || f > 5 {
		return false
	}
	if c == board.Red {
		return r >= 0 && r <= 2
	}
	return r >= 7 && r <= 9
}

// findKing 回傳指定方主將的位置。
func findKing(b *board.Board, c board.Color) board.Square {
	want := board.Piece('K')
	if c == board.Black {
		want = board.Piece('k')
	}
	for s := board.Square(0); s < board.NumSquares; s++ {
		if b.Get(s) == want {
			return s
		}
	}
	return board.InvalidSquare
}

// isAttacked 回報 target 是否被 by 方任一棋子攻擊。
func isAttacked(b *board.Board, target board.Square, by board.Color) bool {
	for s := board.Square(0); s < board.NumSquares; s++ {
		p := b.Get(s)
		if p.IsEmpty() || p.Color() != by {
			continue
		}
		if slices.Contains(pseudoTargets(b, s), target) {
			return true
		}
	}
	return false
}

// kingsFacing 回報兩主將是否在同一直線上直接照面（飛將）。
func kingsFacing(b *board.Board) bool {
	rk := findKing(b, board.Red)
	bk := findKing(b, board.Black)
	if rk == board.InvalidSquare || bk == board.InvalidSquare {
		return false
	}
	if rk.File() != bk.File() {
		return false
	}
	lo, hi := rk.Rank(), bk.Rank()
	if lo > hi {
		lo, hi = hi, lo
	}
	for r := lo + 1; r < hi; r++ {
		if !b.Get(board.MakeSquare(rk.File(), r)).IsEmpty() {
			return false
		}
	}
	return true
}

// inCheck 回報指定方是否被將軍（含飛將）。
func inCheck(b *board.Board, c board.Color) bool {
	k := findKing(b, c)
	if k == board.InvalidSquare {
		return false
	}
	return isAttacked(b, k, c.Opposite()) || kingsFacing(b)
}
