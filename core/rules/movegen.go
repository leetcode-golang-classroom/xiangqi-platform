package rules

import (
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

// inCheck 回報指定方是否被將軍（含飛將）。
// 從將帥位置反向射線掃描，只查 ~20 個特定格子，比正向 isAttacked 快 10–20 倍。
func inCheck(b *board.Board, c board.Color) bool {
	k := findKing(b, c)
	if k == board.InvalidSquare {
		return false
	}
	atk := c.Opposite()
	kf, kr := k.File(), k.Rank()

	// 車、炮、飛將（同縱線/橫線）：沿四正方向掃射線。
	for _, d := range orthoDirs {
		nf, nr := kf, kr
		hasPlatform := false
		for {
			nf += d[0]
			nr += d[1]
			sq := board.MakeSquare(nf, nr)
			if sq == board.InvalidSquare {
				break
			}
			p := b.Get(sq)
			if p.IsEmpty() {
				continue
			}
			if !hasPlatform {
				// 第一個子：車威脅或飛將（對方將帥）
				if p.Color() == atk && (p.Kind() == 'r' || p.Kind() == 'k') {
					return true
				}
				hasPlatform = true
			} else {
				// 第二個子：炮以第一個子為砲架
				if p.Color() == atk && p.Kind() == 'c' {
					return true
				}
				break
			}
		}
	}

	// 馬：反向跳法（從將帥位置反推馬可能的位置）。
	// 蹩馬腿位置是「馬往將帥方向主分量走一步」後的格子，必須從馬的位置計算，
	// 不可從將帥位置加上順向腿偏移（那樣會少算 file/rank 分量）。
	horseCands := [8][2]int{
		{1, 2}, {-1, 2}, {1, -2}, {-1, -2},
		{2, 1}, {2, -1}, {-2, 1}, {-2, -1},
	}
	for _, hc := range horseCands {
		hs := board.MakeSquare(kf+hc[0], kr+hc[1])
		if !hs.Valid() {
			continue
		}
		// 蹩馬腿：馬往將帥方向的主分量走一步
		var legF, legR int
		if hc[0] == 2 || hc[0] == -2 { // 橫向（file）為主
			legF = kf + hc[0]/2
			legR = kr + hc[1]
		} else { // 縱向（rank）為主
			legF = kf + hc[0]
			legR = kr + hc[1]/2
		}
		leg := board.MakeSquare(legF, legR)
		if leg == board.InvalidSquare || !b.Get(leg).IsEmpty() {
			continue // 蹩馬腿擋住
		}
		if hp := b.Get(hs); hp.Kind() == 'n' && hp.Color() == atk {
			return true
		}
	}

	// 兵/卒：反推攻擊來源（兵只能前進或橫走）。
	if atk == board.Red {
		// 紅兵向上（rank+1），能攻擊到 k 的紅兵在 kr-1 正下方。
		if ps := board.MakeSquare(kf, kr-1); ps.Valid() {
			if p := b.Get(ps); p.Kind() == 'p' && p.Color() == board.Red {
				return true
			}
		}
		// 過河紅兵（rank≥5）還可橫向攻擊。
		for _, df := range [2]int{-1, 1} {
			ps := board.MakeSquare(kf+df, kr)
			if ps.Valid() && ps.Rank() >= 5 {
				if p := b.Get(ps); p.Kind() == 'p' && p.Color() == board.Red {
					return true
				}
			}
		}
	} else {
		// 黑卒向下（rank-1），能攻擊到 k 的黑卒在 kr+1 正上方。
		if ps := board.MakeSquare(kf, kr+1); ps.Valid() {
			if p := b.Get(ps); p.Kind() == 'p' && p.Color() == board.Black {
				return true
			}
		}
		// 過河黑卒（rank≤4）還可橫向攻擊。
		for _, df := range [2]int{-1, 1} {
			ps := board.MakeSquare(kf+df, kr)
			if ps.Valid() && ps.Rank() <= 4 {
				if p := b.Get(ps); p.Kind() == 'p' && p.Color() == board.Black {
					return true
				}
			}
		}
	}

	return false
}
