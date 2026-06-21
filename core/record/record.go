// Package record 為棋譜記錄與復盤。
//
// 棋譜以語言中立容器 xiangqi-record-v1（JSON）儲存：走法一律用 UCCI 座標。
package record

import (
	"encoding/json"

	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/core/rules"
)

// Format 為棋譜容器格式識別字串。
const Format = "xiangqi-record-v1"

// Record 為一局棋譜。Moves 為 UCCI 走法序列。
type Record struct {
	Format     string   `json:"format"`
	Red        string   `json:"red"`
	Black      string   `json:"black"`
	Date       string   `json:"date"`
	Result     string   `json:"result"`
	InitialFEN string   `json:"initialFen"`
	Moves      []string `json:"moves"`
}

// Marshal 將棋譜序列化為 JSON。
func Marshal(r Record) ([]byte, error) {
	if r.Format == "" {
		r.Format = Format
	}
	return json.MarshalIndent(r, "", "  ")
}

// Unmarshal 由 JSON 還原棋譜。
func Unmarshal(data []byte) (Record, error) {
	var r Record
	err := json.Unmarshal(data, &r)
	return r, err
}

// Replay 由起始 FEN 起逐步套用走法，回傳每一手後的對局狀態序列
// （含起始狀態，故長度為 len(Moves)+1）。
func Replay(r Record) ([]*rules.Game, error) {
	g, err := rules.FromFEN(r.InitialFEN)
	if err != nil {
		return nil, err
	}
	games := []*rules.Game{g}
	for _, mv := range r.Moves {
		m, err := board.ParseUCCI(mv)
		if err != nil {
			return nil, err
		}
		g, err = g.ApplyMove(m)
		if err != nil {
			return nil, err
		}
		games = append(games, g)
	}
	return games, nil
}
