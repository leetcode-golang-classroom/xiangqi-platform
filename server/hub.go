package server

import (
	"fmt"
	"sync"

	"github.com/yuanyu90221/xiangqi-platform/core/rules"
)

// Hub 管理配對與訊息路由：將送 join 的客戶端配成房間，並依 gameId 將走法路由至 Room。
// 房間由伺服器保留（不因單方斷線而銷毀），以支援 gameId 重連。
type Hub struct {
	mu      sync.Mutex
	waiting Conn // 等待配對的單一客戶端（nil 表示無人等待）
	rooms   map[string]*Room
	seq     int // 房號流水序（決定性，免用亂數/時間）
}

// NewHub 建立空的 Hub。
func NewHub() *Hub {
	return &Hub{rooms: make(map[string]*Room)}
}

// Handle 處理一則來自客戶端的原始訊息（解碼後依型別分派）。
// 損壞 JSON、未知型別或非預期的 server→client 型別回報錯誤。
func (h *Hub) Handle(conn Conn, raw []byte) error {
	env, err := Decode(raw)
	if err != nil {
		return err
	}
	switch env.Type {
	case TypeJoin:
		h.handleJoin(conn, env)
	case TypeMove:
		mp, err := DecodeMove(env)
		if err != nil {
			return err
		}
		h.routeMove(conn, env.GameID, mp.Move)
	case TypeResign:
		// 認輸協商屬後續變更範圍（design.md「不在範圍」）；暫接受不處理。
	default:
		return fmt.Errorf("server: 非預期的客戶端訊息型別 %q", env.Type)
	}
	return nil
}

// handleJoin 處理配對與重連。帶 gameId 視為重連；否則進配對佇列。
func (h *Hub) handleJoin(conn Conn, env Envelope) {
	if env.GameID != "" { // 重連
		h.mu.Lock()
		r := h.rooms[env.GameID]
		h.mu.Unlock()
		if r == nil {
			_ = conn.Send(makeEnvelope(TypeError, env.GameID, ErrorPayload{Reason: "查無此局"}))
			return
		}
		r.reconnect(conn)
		return
	}

	h.mu.Lock()
	if h.waiting == nil { // 第一人，等待
		h.waiting = conn
		h.mu.Unlock()
		return
	}
	// 第二人，配成一房：先到=紅、後到=黑。
	red := h.waiting
	h.waiting = nil
	h.seq++
	id := fmt.Sprintf("g-%d", h.seq)
	game := rules.NewGame()
	h.rooms[id] = NewRoom(id, game, red, conn)
	h.mu.Unlock()

	fen := game.ToFEN()
	_ = red.Send(makeEnvelope(TypeMatched, id, MatchedPayload{Color: "red", InitialFen: fen}))
	_ = conn.Send(makeEnvelope(TypeMatched, id, MatchedPayload{Color: "black", InitialFen: fen}))
}

// routeMove 依 gameId 找房並交由 Room 權威處理。
func (h *Hub) routeMove(conn Conn, gameID, ucci string) {
	h.mu.Lock()
	r := h.rooms[gameID]
	h.mu.Unlock()
	if r == nil {
		_ = conn.Send(makeEnvelope(TypeError, gameID, ErrorPayload{Reason: "查無此局"}))
		return
	}
	r.HandleMove(conn, ucci)
}

// Disconnect 標記該連線於等待佇列／所有房間離線（房間狀態保留以待重連）。
func (h *Hub) Disconnect(conn Conn) {
	h.mu.Lock()
	if h.waiting == conn {
		h.waiting = nil
	}
	rooms := make([]*Room, 0, len(h.rooms))
	for _, r := range h.rooms {
		rooms = append(rooms, r)
	}
	h.mu.Unlock()

	for _, r := range rooms {
		r.disconnect(conn)
	}
}
