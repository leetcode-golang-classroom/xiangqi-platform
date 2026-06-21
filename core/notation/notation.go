// Package notation 負責記譜轉換：FEN ↔ 盤面、UCCI ↔ Move、以及中文記譜顯示。
//
// 儲存一律使用 UCCI（機器標準）；中文記譜僅供顯示，由座標即時轉換。
package notation

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuanyu90221/xiangqi-platform/core/board"
)

// ParseFEN 解析 Xiangqi FEN，回傳盤面、輪走方與計步欄位（halfmove, fullmove）。
func ParseFEN(fen string) (*board.Board, board.Color, int, int, error) {
	parts := strings.Fields(fen)
	if len(parts) < 2 {
		return nil, board.Red, 0, 0, fmt.Errorf("notation: FEN 欄位不足: %q", fen)
	}

	b := &board.Board{}
	rows := strings.Split(parts[0], "/")
	if len(rows) != board.Ranks {
		return nil, board.Red, 0, 0, fmt.Errorf("notation: FEN 應有 %d 行，實得 %d", board.Ranks, len(rows))
	}
	for i, row := range rows {
		rank := board.Ranks - 1 - i // 第 0 行為 rank9
		file := 0
		for _, ch := range row {
			switch {
			case ch >= '1' && ch <= '9':
				file += int(ch - '0')
			default:
				sq := board.MakeSquare(file, rank)
				if sq == board.InvalidSquare {
					return nil, board.Red, 0, 0, fmt.Errorf("notation: FEN 座標超出範圍 (file=%d, rank=%d)", file, rank)
				}
				b.Set(sq, board.Piece(ch))
				file++
			}
		}
		if file != board.Files {
			return nil, board.Red, 0, 0, fmt.Errorf("notation: FEN 第 %d 行寬度不符（得 %d）", i, file)
		}
	}

	turn := board.Red
	if parts[1] == "b" {
		turn = board.Black
	}
	half, full := 0, 1
	if len(parts) >= 5 {
		half, _ = strconv.Atoi(parts[4])
	}
	if len(parts) >= 6 {
		full, _ = strconv.Atoi(parts[5])
	}
	return b, turn, half, full, nil
}

// EncodeFEN 將盤面與狀態輸出為 Xiangqi FEN。
func EncodeFEN(b *board.Board, turn board.Color, halfmove, fullmove int) string {
	var sb strings.Builder
	for rank := board.Ranks - 1; rank >= 0; rank-- {
		empty := 0
		for file := 0; file < board.Files; file++ {
			p := b.Get(board.MakeSquare(file, rank))
			if p.IsEmpty() {
				empty++
				continue
			}
			if empty > 0 {
				sb.WriteString(strconv.Itoa(empty))
				empty = 0
			}
			sb.WriteByte(byte(p))
		}
		if empty > 0 {
			sb.WriteString(strconv.Itoa(empty))
		}
		if rank > 0 {
			sb.WriteByte('/')
		}
	}
	side := "w"
	if turn == board.Black {
		side = "b"
	}
	return fmt.Sprintf("%s %s - - %d %d", sb.String(), side, halfmove, fullmove)
}

// redNum 為紅方使用的中文數字（縱線/步數），索引 1..9。
var redNum = []string{"", "一", "二", "三", "四", "五", "六", "七", "八", "九"}

// pieceNameZh 回傳棋子的中文名稱。
func pieceNameZh(p board.Piece) string {
	red := p.Color() == board.Red
	switch p.Kind() {
	case 'k':
		if red {
			return "帥"
		}
		return "將"
	case 'a':
		if red {
			return "仕"
		}
		return "士"
	case 'b':
		if red {
			return "相"
		}
		return "象"
	case 'n':
		return "馬"
	case 'r':
		return "車"
	case 'c':
		if red {
			return "炮"
		}
		return "砲"
	case 'p':
		if red {
			return "兵"
		}
		return "卒"
	}
	return ""
}

// fileNumZh 回傳縱線編號的中文/數字表示：紅方由右至左以漢字一~九，黑方由右至左以 1~9。
func fileNumZh(fileIdx int, c board.Color) string {
	if c == board.Red {
		return redNum[board.Files-fileIdx] // 紅：9 - fileIdx
	}
	return strconv.Itoa(fileIdx + 1) // 黑：fileIdx + 1
}

// stepNumZh 回傳步數（進/退格數）的表示。
func stepNumZh(n int, c board.Color) string {
	if c == board.Red {
		return redNum[n]
	}
	return strconv.Itoa(n)
}

// ToChinese 將一步走法轉成中文縱線記譜（如「炮二平五」）。
//
// 規則：車/炮/兵(卒)/帥(將) 進退接「步數」、平接「目標縱線」；
// 馬/相(象)/仕(士) 一律接「目標縱線」。同縱線有兩個同種子時以前/後消歧。
func ToChinese(b *board.Board, m board.Move, turn board.Color) string {
	p := b.Get(m.From)
	if p.IsEmpty() {
		return ""
	}
	color := p.Color()
	ff, fr := m.From.File(), m.From.Rank()
	tf, tr := m.To.File(), m.To.Rank()

	// 動作與目標
	var action, target string
	diagonal := p.Kind() == 'n' || p.Kind() == 'a' || p.Kind() == 'b'
	switch {
	case fr == tr:
		action, target = "平", fileNumZh(tf, color)
	default:
		forward := (color == board.Red && tr > fr) || (color == board.Black && tr < fr)
		if forward {
			action = "進"
		} else {
			action = "退"
		}
		if diagonal {
			target = fileNumZh(tf, color)
		} else {
			d := tr - fr
			if d < 0 {
				d = -d
			}
			target = stepNumZh(d, color)
		}
	}

	// 同縱線同種子消歧（前/後）
	var sameFile []int
	for r := 0; r < board.Ranks; r++ {
		q := b.Get(board.MakeSquare(ff, r))
		if !q.IsEmpty() && q.Color() == color && q.Kind() == p.Kind() {
			sameFile = append(sameFile, r)
		}
	}
	if len(sameFile) >= 2 {
		front := sameFile[0]
		for _, r := range sameFile {
			if (color == board.Red && r > front) || (color == board.Black && r < front) {
				front = r
			}
		}
		token := "後"
		if fr == front {
			token = "前"
		}
		return token + pieceNameZh(p) + action + target
	}

	return pieceNameZh(p) + fileNumZh(ff, color) + action + target
}
