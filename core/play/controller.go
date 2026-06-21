package play

import (
	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/core/rules"
	"github.com/yuanyu90221/xiangqi-platform/player"
)

// Controller 協調一局：持有 Session 與各方 player.Player（人類/AI/遠端），以統一迴圈推進。
// 不區分對手種類——每次 Step 向當前輪走方的 Player 請求一步，完成即套用。
type Controller struct {
	s       *Session
	players map[board.Color]player.Player

	pending  <-chan board.Move // 當前回合進行中的取步通道
	inFlight bool
}

// NewController 以 Session 與紅、黑兩方 Player 建立協調器。
func NewController(s *Session, red, black player.Player) *Controller {
	return &Controller{
		s:       s,
		players: map[board.Color]player.Player{board.Red: red, board.Black: black},
	}
}

// VsComputer 建立人機對局：humanColor 為人類執方，另一方由 ai 控制。
// 回傳對局協調器與人類玩家（供 UI 餵入點擊、讀取選取高亮）。
func VsComputer(red, black string, humanColor board.Color, ai player.Player) (*Controller, *player.Human) {
	s := NewSession(red, black)
	h := player.NewHuman("玩家")
	if humanColor == board.Red {
		return NewController(s, h, ai), h
	}
	return NewController(s, ai, h), h
}

// Session 回傳底層對局狀態。
func (c *Controller) Session() *Session { return c.s }

// Current 回傳當前盤面。
func (c *Controller) Current() *rules.Game { return c.s.Current() }

// Turn 回傳當前輪走方。
func (c *Controller) Turn() board.Color { return c.s.Turn() }

// Outcome 回傳對局結果。
func (c *Controller) Outcome() rules.Result { return c.s.Outcome() }

// CurrentPlayer 回傳當前輪走方的 Player。
func (c *Controller) CurrentPlayer() player.Player { return c.players[c.s.Turn()] }

// CurrentInteractive 回傳當前輪走方的互動式玩家（人類，若當前為人類且對局未結束）；
// UI 以此餵入點擊並讀取選取高亮，而 Controller 不需依賴具體型別。
func (c *Controller) CurrentInteractive() (player.Interactive, bool) {
	if c.s.Outcome().Over {
		return nil, false
	}
	iv, ok := c.CurrentPlayer().(player.Interactive)
	return iv, ok
}

// Thinking 回報當前是否正等待非互動式玩家（AI/遠端）取步。
func (c *Controller) Thinking() bool {
	if !c.inFlight || c.s.Outcome().Over {
		return false
	}
	_, interactive := c.CurrentPlayer().(player.Interactive)
	return !interactive
}

// Step 推進統一迴圈一格：對局未結束時，向當前 Player 請求一步（若尚未請求），
// 並在其送出後套用。回傳本次是否套用了一步。應每幀呼叫。
func (c *Controller) Step() bool {
	if c.s.Outcome().Over {
		c.inFlight = false
		return false
	}
	if !c.inFlight {
		c.pending = c.CurrentPlayer().RequestMove(c.s.Current())
		c.inFlight = true
		return false
	}
	select {
	case m := <-c.pending:
		c.inFlight = false
		c.pending = nil
		_ = c.s.Play(m)
		return true
	default:
		return false
	}
}

// Undo 回退最後一手，並重置進行中的取步請求。
func (c *Controller) Undo() bool {
	c.cancelPending()
	return c.s.Undo()
}

// Resign 當前輪走方認輸，並重置進行中的取步請求。
func (c *Controller) Resign() {
	c.cancelPending()
	c.s.Resign()
}

// cancelPending 取消進行中的取步：丟棄通道（背景搜尋的結果將被忽略）。
func (c *Controller) cancelPending() {
	c.inFlight = false
	c.pending = nil
}
