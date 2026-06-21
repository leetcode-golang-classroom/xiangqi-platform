// Package play 為本機對局的狀態與互動。
//
// 分層：
//   - Session：純粹的對局權威狀態（走子歷史、記譜、悔棋、認輸、Play），不含 UI 選取。
//   - Player（結構化介面）+ Human/AI：統一的「取步者」，非同步產出一步。
//   - Controller：統一對局迴圈，向當前 Player 請求一步並套用，不分對手種類。
package play

import (
	"errors"

	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/core/record"
	"github.com/yuanyu90221/xiangqi-platform/core/rules"
)

// ErrGameOver 表示對局已結束，無法再走子。
var ErrGameOver = errors.New("play: 對局已結束")

// Session 持有一局的權威狀態（不含 UI 選取狀態）。
type Session struct {
	red, black string
	initialFEN string
	games      []*rules.Game // games[0]=起始；len = 走法數+1
	moves      []board.Move

	manualWinner string // 認輸等人工結果的勝方（"red"/"black"）；空表示無
}

// NewSession 由標準開局建立對局。
func NewSession(red, black string) *Session {
	g := rules.NewGame()
	return &Session{
		red:        red,
		black:      black,
		initialFEN: g.ToFEN(),
		games:      []*rules.Game{g},
	}
}

// Current 回傳當前（最後一手之後）盤面。
func (s *Session) Current() *rules.Game { return s.games[len(s.games)-1] }

// Turn 回傳當前輪走方。
func (s *Session) Turn() board.Color { return s.Current().Turn() }

// Outcome 回傳對局結果：人工結果（認輸）優先，否則沿用 rule-engine 判定。
func (s *Session) Outcome() rules.Result {
	if s.manualWinner != "" {
		return rules.Result{Over: true, Winner: s.manualWinner, Reason: "resign"}
	}
	return s.Current().Result()
}

// Play 套用一步走法。走法須合法且對局未結束，否則回報錯誤。成功後推進盤面並記錄。
func (s *Session) Play(m board.Move) error {
	if s.Outcome().Over {
		return ErrGameOver
	}
	ng, err := s.Current().ApplyMove(m)
	if err != nil {
		return err
	}
	s.games = append(s.games, ng)
	s.moves = append(s.moves, m)
	return nil
}

// Undo 回退最後一手；無走法時回傳 false（無操作）。
func (s *Session) Undo() bool {
	if len(s.moves) == 0 {
		return false
	}
	s.games = s.games[:len(s.games)-1]
	s.moves = s.moves[:len(s.moves)-1]
	s.manualWinner = ""
	return true
}

// Resign 由當前輪走方認輸，判對方勝。
func (s *Session) Resign() {
	if s.Turn() == board.Red {
		s.manualWinner = "black"
	} else {
		s.manualWinner = "red"
	}
}

// Record 輸出當前棋譜（xiangqi-record-v1，走法為 UCCI）。
func (s *Session) Record() record.Record {
	moves := make([]string, len(s.moves))
	for i, m := range s.moves {
		moves[i] = m.String()
	}
	return record.Record{
		Format:     record.Format,
		Red:        s.red,
		Black:      s.black,
		Result:     s.resultString(),
		InitialFEN: s.initialFEN,
		Moves:      moves,
	}
}

// resultString 將結果映射為棋譜的結果字串。
func (s *Session) resultString() string {
	o := s.Outcome()
	if !o.Over {
		return ""
	}
	switch o.Winner {
	case "red":
		return "red_win"
	case "black":
		return "black_win"
	default:
		return "draw"
	}
}
