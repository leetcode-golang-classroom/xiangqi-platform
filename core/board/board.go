// Package board 定義象棋盤面的座標系與棋子表示。
//
// 座標系（UCCI 慣例）：縱線 file a–i（紅方視角左→右，0..8），
// 橫線 rank 0–9（紅底→黑頂）。一格以 file+rank 表示，如 "e0"。
//
// 本套件僅含純資料結構，無規則邏輯，完全可移植。
package board

const (
	// Files 縱線數（a..i）。
	Files = 9
	// Ranks 橫線數（0..9）。
	Ranks = 10
	// NumSquares 棋盤總點數。
	NumSquares = Files * Ranks
)

// Color 表示一方顏色。
type Color int

const (
	// Red 紅方（楚，FEN 大寫）。
	Red Color = iota
	// Black 黑方（漢，FEN 小寫）。
	Black
)

func (c Color) String() string {
	if c == Red {
		return "red"
	}
	return "black"
}

// Opposite 回傳對手顏色。
func (c Color) Opposite() Color {
	if c == Red {
		return Black
	}
	return Red
}

// Piece 以 FEN 字母表示棋子：大寫=紅、小寫=黑。
//
//	K/k 帥將  A/a 仕士  B/b 相象  N/n 馬  R/r 車  C/c 炮  P/p 兵卒
//
// Empty 表示空格。
type Piece byte

// Empty 為空格。
const Empty Piece = 0

// Square 為棋盤座標索引：index = rank*Files + file，範圍 0..NumSquares-1。
type Square int

// InvalidSquare 為無效座標。
const InvalidSquare Square = -1

// MakeSquare 由 file(0..8) 與 rank(0..9) 組成 Square。
func MakeSquare(file, rank int) Square {
	if file < 0 || file >= Files || rank < 0 || rank >= Ranks {
		return InvalidSquare
	}
	return Square(rank*Files + file)
}

// File 回傳縱線索引 0..8。
func (s Square) File() int { return int(s) % Files }

// Rank 回傳橫線索引 0..9。
func (s Square) Rank() int { return int(s) / Files }

// Valid 回報座標是否在盤面內。
func (s Square) Valid() bool { return s >= 0 && s < NumSquares }

// Board 為 9×10 盤面，以 Piece 陣列表示。
type Board struct {
	cells [NumSquares]Piece
}

// Get 取得指定格的棋子（空格回傳 Empty）。
func (b *Board) Get(sq Square) Piece { return b.cells[sq] }

// Set 設定指定格的棋子。
func (b *Board) Set(sq Square, p Piece) { b.cells[sq] = p }

// Clone 回傳盤面的深拷貝。
func (b *Board) Clone() *Board {
	nb := *b
	return &nb
}
