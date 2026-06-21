package player

import (
	"slices"

	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/core/rules"
)

// Human 為人類玩家：實作 Player（與 Interactive），以點選狀態機
// （選子／高亮落點／改選／取消）漸進產出一步。RequestMove 武裝一回合並回傳通道；
// 當點到合法落點時將該步送入通道，交由對局迴圈套用。
type Human struct {
	name string

	game     *rules.Game // 當前回合的盤面（由 RequestMove 武裝）
	armed    bool        // 是否處於等待人類輸入的回合
	ch       chan board.Move
	selected board.Square
	hasSel   bool
	targets  []board.Square
}

// NewHuman 建立人類玩家。
func NewHuman(name string) *Human { return &Human{name: name} }

// Name 回傳玩家名稱。
func (h *Human) Name() string { return h.name }

// RequestMove 武裝一回合：記住當前盤面、清除選取，回傳本回合的走法通道。
func (h *Human) RequestMove(g *rules.Game) <-chan board.Move {
	h.game = g
	h.armed = true
	h.ch = make(chan board.Move, 1)
	h.clearSelection()
	return h.ch
}

// Selected 回傳目前選取的棋格（供 UI 高亮）。
func (h *Human) Selected() (board.Square, bool) { return h.selected, h.hasSel }

// Targets 回傳目前選取子的合法落點（供 UI 高亮）。
func (h *Human) Targets() []board.Square { return h.targets }

// Tap 以一次棋格點擊驅動選子狀態機。未武裝（非本方回合）時忽略。
func (h *Human) Tap(at board.Square) TapResult {
	if !h.armed || h.game == nil {
		return TapIgnored
	}

	if h.hasSel {
		if at == h.selected {
			h.clearSelection()
			return TapCleared
		}
		if slices.Contains(h.targets, at) { // 點擊合法落點 → 產出走法
			h.emit(board.Move{From: h.selected, To: at})
			return TapMoved
		}
		if ts := h.movesFrom(at); len(ts) > 0 { // 改選另一己方可動子
			h.selectAt(at, ts)
			return TapSelected
		}
		h.clearSelection() // 其餘 → 取消
		return TapCleared
	}

	if ts := h.movesFrom(at); len(ts) > 0 { // 未選子：選取己方可動子
		h.selectAt(at, ts)
		return TapSelected
	}
	return TapIgnored
}

// emit 送出產出的走法並解除武裝（待迴圈套用後，下一回合再 RequestMove 重新武裝）。
func (h *Human) emit(m board.Move) {
	h.ch <- m
	h.armed = false
	h.clearSelection()
}

// movesFrom 回傳當前盤面上由 from 出發的合法走法目標；空表示該格非己方可動子。
func (h *Human) movesFrom(from board.Square) []board.Square {
	var ts []board.Square
	for _, m := range h.game.LegalMoves() {
		if m.From == from {
			ts = append(ts, m.To)
		}
	}
	return ts
}

func (h *Human) clearSelection() {
	h.hasSel = false
	h.targets = nil
}

func (h *Human) selectAt(from board.Square, targets []board.Square) {
	h.selected = from
	h.hasSel = true
	h.targets = targets
}
