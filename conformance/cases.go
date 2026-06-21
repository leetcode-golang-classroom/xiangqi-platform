// Package conformance 載入語言中立的黃金測試案例（fixtures）。
//
// fixtures 為 conformance/*.json，描述「輸入盤面 → 期望輸出」。
// 各語言實作只需一支薄 harness 載入這些案例並斷言，確保跨語言一致。
package conformance

import (
	"embed"
	"encoding/json"
)

//go:embed *.json
var fixtures embed.FS

// FENCase：FEN round-trip 案例。
type FENCase struct {
	Name string `json:"name"`
	FEN  string `json:"fen"`
}

// LegalMovesCase：從 From 格出發的合法走法集合。
type LegalMovesCase struct {
	Name     string   `json:"name"`
	FEN      string   `json:"fen"`
	From     string   `json:"from"`
	Expected []string `json:"expected"`
}

// ResultExpect：期望的對局結果。
type ResultExpect struct {
	Over   bool   `json:"over"`
	Winner string `json:"winner"`
	Reason string `json:"reason"`
}

// ResultCase：勝負判定案例。
type ResultCase struct {
	Name   string       `json:"name"`
	FEN    string       `json:"fen"`
	Result ResultExpect `json:"result"`
}

// RecordCase：棋譜重放案例。
type RecordCase struct {
	Name       string   `json:"name"`
	InitialFEN string   `json:"initialFen"`
	Moves      []string `json:"moves"`
	FinalFEN   string   `json:"finalFen"`
}

// ChineseCase：中文記譜案例。
type ChineseCase struct {
	Name    string `json:"name"`
	FEN     string `json:"fen"`
	Move    string `json:"move"`
	Chinese string `json:"chinese"`
}

// AdjudicateCase：特殊裁定（長將判負等）案例。
type AdjudicateCase struct {
	Name       string       `json:"name"`
	InitialFEN string       `json:"initialFen"`
	Moves      []string     `json:"moves"`
	Result     ResultExpect `json:"result"`
}

// MovelistCase：走法序列 → 中文記譜清單案例。
type MovelistCase struct {
	Name       string   `json:"name"`
	InitialFEN string   `json:"initialFen"`
	Moves      []string `json:"moves"`
	Chinese    []string `json:"chinese"`
}

func load(name string, out any) error {
	data, err := fixtures.ReadFile(name)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

// FENCases 載入 FEN round-trip 案例。
func FENCases() ([]FENCase, error) {
	var f struct {
		Cases []FENCase `json:"cases"`
	}
	err := load("fen_cases.json", &f)
	return f.Cases, err
}

// LegalMovesCases 載入合法走法案例。
func LegalMovesCases() ([]LegalMovesCase, error) {
	var f struct {
		Cases []LegalMovesCase `json:"cases"`
	}
	err := load("legalmoves_cases.json", &f)
	return f.Cases, err
}

// ResultCases 載入勝負判定案例。
func ResultCases() ([]ResultCase, error) {
	var f struct {
		Cases []ResultCase `json:"cases"`
	}
	err := load("result_cases.json", &f)
	return f.Cases, err
}

// RecordCases 載入棋譜重放案例。
func RecordCases() ([]RecordCase, error) {
	var f struct {
		Cases []RecordCase `json:"cases"`
	}
	err := load("record_cases.json", &f)
	return f.Cases, err
}

// ChineseCases 載入中文記譜案例。
func ChineseCases() ([]ChineseCase, error) {
	var f struct {
		Cases []ChineseCase `json:"cases"`
	}
	err := load("chinese_cases.json", &f)
	return f.Cases, err
}

// AdjudicateCases 載入特殊裁定案例。
func AdjudicateCases() ([]AdjudicateCase, error) {
	var f struct {
		Cases []AdjudicateCase `json:"cases"`
	}
	err := load("adjudicate_cases.json", &f)
	return f.Cases, err
}

// MovelistCases 載入中文記譜清單案例。
func MovelistCases() ([]MovelistCase, error) {
	var f struct {
		Cases []MovelistCase `json:"cases"`
	}
	err := load("movelist_cases.json", &f)
	return f.Cases, err
}
