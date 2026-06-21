package board

// IsEmpty 回報是否為空格。
func (p Piece) IsEmpty() bool { return p == Empty }

// Color 回傳棋子顏色（大寫=紅、小寫=黑）。空格不應呼叫此方法。
func (p Piece) Color() Color {
	if p >= 'A' && p <= 'Z' {
		return Red
	}
	return Black
}

// Kind 回傳棋子種類（一律小寫字母）：
//
//	r 車 n 馬 c 炮 b 相象 a 仕士 k 帥將 p 兵卒
func (p Piece) Kind() byte {
	if p >= 'A' && p <= 'Z' {
		return byte(p) - 'A' + 'a'
	}
	return byte(p)
}
