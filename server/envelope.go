// Package server 實作 L2 線上對戰：WebSocket 傳輸與權威伺服器（配對／驗證／廣播／重連）。
//
// 本檔定義協定 envelope —— 線上對戰的傳輸格式，欄位對齊 docs/design/contracts.md §4。
// 走法一律以 UCCI 表示；中文記譜由客戶端即時轉換顯示。
package server

import (
	"encoding/json"
	"fmt"
)

// MsgType 為協定訊息型別。
type MsgType string

const (
	// client → server
	TypeJoin   MsgType = "join"   // 加入配對；重連時帶既有 gameId
	TypeMove   MsgType = "move"   // 送出一步走法（UCCI）
	TypeResign MsgType = "resign" // 認輸

	// server → client
	TypeMatched     MsgType = "matched"      // 配對成局：己方執色與初始 FEN
	TypeMoveApplied MsgType = "move_applied" // 權威套用後廣播：走法、FEN、輪走方
	TypeGameOver    MsgType = "game_over"    // 對局結束：結果與原因
	TypeError       MsgType = "error"        // 僅回送出方：錯誤原因（不廣播）
	TypeStateSync   MsgType = "state_sync"   // 重連還原：目前 FEN 與走法序列
)

// valid 回報型別是否為已知協定型別。
func (t MsgType) valid() bool {
	switch t {
	case TypeJoin, TypeMove, TypeResign,
		TypeMatched, TypeMoveApplied, TypeGameOver, TypeError, TypeStateSync:
		return true
	default:
		return false
	}
}

// Envelope 為傳輸外層：{ type, gameId, payload }。
// payload 以 json.RawMessage 保留，待依 type 解碼成對應 payload 型別。
type Envelope struct {
	Type    MsgType         `json:"type"`
	GameID  string          `json:"gameId"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// ── payload 型別（對齊 design.md envelope 表）──

// MovePayload — client→server move：UCCI 走法。
type MovePayload struct {
	Move string `json:"move"`
}

// MatchedPayload — server→client matched：己方執色與初始 FEN。
type MatchedPayload struct {
	Color      string `json:"color"` // "red" / "black"
	InitialFen string `json:"initialFen"`
}

// MoveAppliedPayload — server→client move_applied：套用後盤面與輪走方。
type MoveAppliedPayload struct {
	Move string `json:"move"` // UCCI
	Fen  string `json:"fen"`
	Turn string `json:"turn"` // 輪走方 "red" / "black"
}

// GameOverPayload — server→client game_over：結果與原因。
type GameOverPayload struct {
	Result string `json:"result"` // "red" / "black" / "draw"
	Reason string `json:"reason"`
}

// ErrorPayload — server→client error：錯誤原因（僅回送出方）。
type ErrorPayload struct {
	Reason string `json:"reason"`
}

// StateSyncPayload — server→client state_sync：目前 FEN 與已走 UCCI 序列。
type StateSyncPayload struct {
	Fen   string   `json:"fen"`
	Moves []string `json:"moves"`
}

// Encode 將給定型別、gameId 與 payload 編成 envelope JSON。
// payload 為 nil 時（如 resign）省略 payload 欄位。未知型別回報錯誤。
func Encode(typ MsgType, gameID string, payload any) ([]byte, error) {
	if !typ.valid() {
		return nil, fmt.Errorf("server: 未知的訊息型別 %q", typ)
	}
	var raw json.RawMessage
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("server: 編碼 payload 失敗：%w", err)
		}
		raw = b
	}
	return json.Marshal(Envelope{Type: typ, GameID: gameID, Payload: raw})
}

// makeEnvelope 在伺服器內部建構一則 envelope（payload 已序列化）。
// payload 為簡單結構，json.Marshal 不會失敗；為 nil 時省略 payload。
func makeEnvelope(typ MsgType, gameID string, payload any) Envelope {
	var raw json.RawMessage
	if payload != nil {
		b, _ := json.Marshal(payload)
		raw = b
	}
	return Envelope{Type: typ, GameID: gameID, Payload: raw}
}

// Decode 解碼 envelope JSON。損壞 JSON 或未知型別皆回報錯誤（不靜默忽略、不 panic）。
func Decode(data []byte) (Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return Envelope{}, fmt.Errorf("server: 解碼 envelope 失敗：%w", err)
	}
	if !env.Type.valid() {
		return Envelope{}, fmt.Errorf("server: 未知的訊息型別 %q", env.Type)
	}
	return env, nil
}

// DecodePayload 將 envelope 的 payload 解碼到 out（指標）。
func DecodePayload(env Envelope, out any) error {
	if len(env.Payload) == 0 {
		return fmt.Errorf("server: envelope（型別 %q）無 payload 可解碼", env.Type)
	}
	if err := json.Unmarshal(env.Payload, out); err != nil {
		return fmt.Errorf("server: 解碼 payload 失敗：%w", err)
	}
	return nil
}

// DecodeMove 解碼 move payload。
func DecodeMove(env Envelope) (MovePayload, error) {
	var mp MovePayload
	err := DecodePayload(env, &mp)
	return mp, err
}
