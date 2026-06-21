package record

import "github.com/yuanyu90221/xiangqi-platform/core/rules"

// Replayer 為復盤游標：以游標包裝 Timeline，支援逐手前進/後退與跳轉，
// 並回傳當前盤面。所有移動皆夾制於 [0, Len-1]。供 GUI 與其他前端共用。
type Replayer struct {
	tl        *Timeline
	i         int
	notations []string // 每手中文記譜（len = Len-1，與走法數一致）
}

// NewReplayer 由棋譜建立復盤游標，游標起始於 0（起始局面）。
func NewReplayer(rec Record) (*Replayer, error) {
	tl, err := NewTimeline(rec)
	if err != nil {
		return nil, err
	}
	notations, err := MovesInChinese(rec)
	if err != nil {
		return nil, err
	}
	return &Replayer{tl: tl, notations: notations}, nil
}

// Len 回傳盤面數（走法數 + 1）。
func (r *Replayer) Len() int { return r.tl.Len() }

// Notations 回傳每一手的中文記譜（長度為 Len-1，與走法數一致；第 i 項為第 i+1 手）。
// 供前端在跳手清單與狀態列顯示記譜，毋須各自重算。
func (r *Replayer) Notations() []string { return r.notations }

// Index 回傳當前游標位置（0 為起始局面）。
func (r *Replayer) Index() int { return r.i }

// Current 回傳游標當前位置的盤面。
func (r *Replayer) Current() *rules.Game { return r.tl.At(r.i) }

// Next 前進一手；已在末位則不動並回傳 false。
func (r *Replayer) Next() bool {
	if r.i >= r.tl.Len()-1 {
		return false
	}
	r.i++
	return true
}

// Prev 後退一手；已在起始則不動並回傳 false。
func (r *Replayer) Prev() bool {
	if r.i <= 0 {
		return false
	}
	r.i--
	return true
}

// Seek 跳轉至索引 i，越界則夾制於 [0, Len-1]。
func (r *Replayer) Seek(i int) {
	switch {
	case i < 0:
		i = 0
	case i > r.tl.Len()-1:
		i = r.tl.Len() - 1
	}
	r.i = i
}
