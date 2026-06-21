package server

import (
	"sync"

	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/core/rules"
)

// Conn 為單一客戶端連線的傳輸抽象：伺服器送一則 envelope 給此客戶端。
// WebSocket 連線與測試替身皆可實作，使 Room/Hub 邏輯傳輸無關、可決定性測試。
type Conn interface {
	Send(env Envelope) error
}

// Room 持有一局的權威狀態：對局、紅/黑座位、走法序列。
// 每步以 core/rules 權威驗證後向雙方廣播（規則正確性由 rules 既有測試保證，
// Room 只負責「誰能走、何時廣播」）。所有狀態變更經 mu 序列化（伺服器併發）。
type Room struct {
	mu    sync.Mutex
	id    string
	game  *rules.Game
	red   Conn
	black Conn
	moves []string // 已走 UCCI 序列（供重連 state_sync）
	over  bool
}

// NewRoom 以給定對局狀態與紅/黑座位建立房間。
func NewRoom(id string, game *rules.Game, red, black Conn) *Room {
	return &Room{id: id, game: game, red: red, black: black}
}

// colorOf 回報該連線於本局的執色；非本局玩家回 ok=false。
func (r *Room) colorOf(c Conn) (board.Color, bool) {
	switch c {
	case r.red:
		return board.Red, true
	case r.black:
		return board.Black, true
	default:
		return board.Red, false
	}
}

func (r *Room) sendErr(c Conn, reason string) {
	if c != nil {
		_ = c.Send(makeEnvelope(TypeError, r.id, ErrorPayload{Reason: reason}))
	}
}

// broadcast 向雙方（仍在線者）送出同一 envelope。
func (r *Room) broadcast(env Envelope) {
	if r.red != nil {
		_ = r.red.Send(env)
	}
	if r.black != nil {
		_ = r.black.Send(env)
	}
}

// HandleMove 權威處理一步走法：回合檢查 → 規則驗證 → 套用並廣播。
// 非當前回合方、非法走法、局終後走子皆僅回送出方 error、不廣播、狀態不變。
func (r *Room) HandleMove(from Conn, ucci string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.over {
		r.sendErr(from, "對局已結束")
		return
	}
	color, ok := r.colorOf(from)
	if !ok {
		r.sendErr(from, "非本局玩家")
		return
	}
	if color != r.game.Turn() {
		r.sendErr(from, "尚未輪到你走")
		return
	}
	m, err := board.ParseUCCI(ucci)
	if err != nil {
		r.sendErr(from, "走法格式錯誤")
		return
	}
	ng, err := r.game.ApplyMove(m)
	if err != nil {
		r.sendErr(from, "非法走法")
		return
	}

	r.game = ng
	r.moves = append(r.moves, ucci)
	r.broadcast(makeEnvelope(TypeMoveApplied, r.id, MoveAppliedPayload{
		Move: ucci,
		Fen:  ng.ToFEN(),
		Turn: ng.Turn().String(),
	}))

	if res := ng.Result(); res.Over {
		r.over = true
		result := res.Winner
		if result == "" {
			result = "draw"
		}
		r.broadcast(makeEnvelope(TypeGameOver, r.id, GameOverPayload{
			Result: result,
			Reason: res.Reason,
		}))
	}
}

// disconnect 將該連線的座位標為空（不銷毀房間，保留狀態以待重連）。
func (r *Room) disconnect(c Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.red == c {
		r.red = nil
	}
	if r.black == c {
		r.black = nil
	}
}

// reconnect 將新連線填入空座位，並回 state_sync 還原盤面與走法序列。
func (r *Room) reconnect(c Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch {
	case r.red == nil:
		r.red = c
	case r.black == nil:
		r.black = c
	}
	_ = c.Send(makeEnvelope(TypeStateSync, r.id, StateSyncPayload{
		Fen:   r.game.ToFEN(),
		Moves: append([]string(nil), r.moves...),
	}))
}
