// Package player 為對局的「取步者」：抽象介面與其實作（人類、AI、日後遠端）。
//
// 術語統一：所有與「玩家」相關者皆歸此套件——`Player`/`Interactive` 介面與
// `Human`/`AI` 實作。對局機制（`Session` 狀態、`Controller` 迴圈）在 core/play，
// 單向依賴本套件；本套件不依賴 core/play，避免 import cycle。
package player

import (
	"errors"

	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/core/rules"
)

// ErrGameOver 表示對局已結束，無法再選步。
var ErrGameOver = errors.New("player: 對局已結束")

// Player 為統一的取步者抽象：本地人類、AI、（日後）遠端皆實作。
// RequestMove 為非同步：回傳一個通道，當該方決定一步時送出之。對局迴圈只依賴此介面。
type Player interface {
	Name() string
	// RequestMove 請求一步：回傳的通道將於該方決定時送出一步走法。
	// 僅於對局未結束且輪到該方時呼叫。
	RequestMove(g *rules.Game) <-chan board.Move
}

// Interactive 為需要人類漸進輸入的 Player（如 Human）：以點擊驅動選子，
// 並可供 UI 讀取選取與合法落點高亮。對局迴圈以此介面分辨人類回合，
// 而不需依賴具體型別。
type Interactive interface {
	Player
	Tap(at board.Square) TapResult
	Selected() (board.Square, bool)
	Targets() []board.Square
}

// TapResult 描述一次點擊造成的結果。
type TapResult int

const (
	TapIgnored  TapResult = iota // 無效點擊（未武裝、點對方子等）
	TapSelected                  // 選取或改選了一個己方可動子
	TapMoved                     // 對已選子點到合法落點，產出一步
	TapCleared                   // 取消了選取
)
