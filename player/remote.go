package player

import (
	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/core/rules"
)

// MoveTransport 為「傳輸中立」的遠端走法收發抽象：對局核心只依賴此介面，
// 不關心底層是 WebSocket 即時伺服器、LINE bot（回合制非同步）或 LIFF 網頁。
// 任一種線上方式，只要實作 MoveTransport，即可接入既有的 Player/Controller 迴圈，
// 核心邏輯（rules/play）完全不需改動。
//
// 注意：此處僅定義介面（接縫），尚未提供任何傳輸實作——線上對戰方式待定後再實作。
type MoveTransport interface {
	// Incoming 回傳遠端對手走法的串流：每當對手（經權威伺服器確認）走一步即送出。
	Incoming() <-chan board.Move
	// Send 將本地走法送往遠端（權威伺服器/對手）；回報傳輸層錯誤。
	Send(m board.Move) error
}

// RemotePlayer 以 MoveTransport 為走法來源，將遠端對手包裝成統一的 Player。
// 它是「介面骨架」：把 RequestMove 接到傳輸層的對手走法串流即可，
// 真正的連線、配對、權威驗證等由日後的 MoveTransport 實作負責。
type RemotePlayer struct {
	name string
	tr   MoveTransport
}

// NewRemotePlayer 以顯示名稱與傳輸層建立遠端玩家。
func NewRemotePlayer(name string, tr MoveTransport) *RemotePlayer {
	return &RemotePlayer{name: name, tr: tr}
}

// Name 回傳遠端玩家顯示名稱。
func (p *RemotePlayer) Name() string { return p.name }

// RequestMove 回傳傳輸層的對手走法串流：對局迴圈待其送出一步即套用。
// 盤面參數於此實作未使用（遠端對手自有盤面），保留以符合 Player 介面。
func (p *RemotePlayer) RequestMove(_ *rules.Game) <-chan board.Move {
	return p.tr.Incoming()
}
