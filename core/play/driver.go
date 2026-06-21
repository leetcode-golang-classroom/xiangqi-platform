package play

import (
	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/core/rules"
	"github.com/yuanyu90221/xiangqi-platform/player"
)

// Driver 為渲染層（GUI）推進與繪製一局所需的最小面：離線的 Controller（人機）與
// 線上的 OnlineController（權威伺服器對戰）皆滿足，故 GUI 可以同一套程式驅動兩者。
//
// 離線專屬操作（Undo／Resign）不在此介面，由 GUI 於離線模式下對具體型別呼叫。
type Driver interface {
	// Current 回傳當前盤面。
	Current() *rules.Game
	// Turn 回傳當前輪走方。
	Turn() board.Color
	// Outcome 回傳對局結果。
	Outcome() rules.Result
	// CurrentInteractive 回傳當前可接受輸入的互動式玩家（若有）。
	CurrentInteractive() (player.Interactive, bool)
	// Thinking 回報當前是否正等待非本地輸入（對手／伺服器確認／AI）。
	Thinking() bool
	// Step 推進迴圈一格，回傳本次是否套用了一步。應每幀呼叫。
	Step() bool
	// Session 回傳底層對局狀態（記譜、結果等）。
	Session() *Session
}

// 編譯期保證：離線 Controller 滿足 Driver。
var _ Driver = (*Controller)(nil)
