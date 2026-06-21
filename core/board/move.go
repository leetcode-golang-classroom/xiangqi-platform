package board

import "fmt"

// Move 表示一步走法：起點到終點。對應 UCCI 座標表示法，如 "h2e2"。
type Move struct {
	From Square
	To   Square
}

// fileRune 將 file 索引 0..8 轉成 'a'..'i'。
func fileRune(file int) byte { return byte('a' + file) }

// ParseSquare 解析 "e0" 形式的座標字串。
func ParseSquare(s string) (Square, error) {
	if len(s) != 2 {
		return InvalidSquare, fmt.Errorf("board: 無效座標 %q", s)
	}
	file := int(s[0] - 'a')
	rank := int(s[1] - '0')
	sq := MakeSquare(file, rank)
	if sq == InvalidSquare {
		return InvalidSquare, fmt.Errorf("board: 座標超出範圍 %q", s)
	}
	return sq, nil
}

// String 以 "e0" 形式輸出座標。
func (s Square) String() string {
	if !s.Valid() {
		return "--"
	}
	return fmt.Sprintf("%c%d", fileRune(s.File()), s.Rank())
}

// ParseUCCI 解析 "h2e2" 形式的走法字串。
func ParseUCCI(s string) (Move, error) {
	if len(s) != 4 {
		return Move{}, fmt.Errorf("board: 無效 UCCI 走法 %q", s)
	}
	from, err := ParseSquare(s[:2])
	if err != nil {
		return Move{}, err
	}
	to, err := ParseSquare(s[2:])
	if err != nil {
		return Move{}, err
	}
	return Move{From: from, To: to}, nil
}

// String 以 UCCI 形式輸出走法，如 "h2e2"。
func (m Move) String() string { return m.From.String() + m.To.String() }
