package record

import (
	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/core/rules"
)

// Recorder 於對局進行中漸進記錄棋譜：逐手附加合法走法並維護當前盤面。
type Recorder struct {
	red, black string
	initialFEN string
	game       *rules.Game
	moves      []string
	result     string
}

// NewRecorder 由指定起始 FEN 開新局記錄。
func NewRecorder(initialFEN, red, black string) (*Recorder, error) {
	g, err := rules.FromFEN(initialFEN)
	if err != nil {
		return nil, err
	}
	return &Recorder{red: red, black: black, initialFEN: initialFEN, game: g}, nil
}

// NewStartRecorder 由標準開局開新局記錄。
func NewStartRecorder(red, black string) *Recorder {
	g := rules.NewGame()
	return &Recorder{red: red, black: black, initialFEN: g.ToFEN(), game: g}
}

// Add 附加一步 UCCI 走法；走法須合法，否則回報錯誤且不記錄。
func (r *Recorder) Add(uci string) error {
	m, err := board.ParseUCCI(uci)
	if err != nil {
		return err
	}
	ng, err := r.game.ApplyMove(m) // 驗證合法性
	if err != nil {
		return err
	}
	r.game = ng
	r.moves = append(r.moves, uci)
	return nil
}

// SetResult 標記對局結果（red_win/black_win/draw 等）。
func (r *Recorder) SetResult(result string) { r.result = result }

// Current 回傳當前（最後一手之後）的對局狀態。
func (r *Recorder) Current() *rules.Game { return r.game }

// Record 輸出目前累積的棋譜。
func (r *Recorder) Record() Record {
	return Record{
		Format:     Format,
		Red:        r.red,
		Black:      r.black,
		Result:     r.result,
		InitialFEN: r.initialFEN,
		Moves:      append([]string(nil), r.moves...),
	}
}

// Timeline 為復盤導覽：每一手後的盤面序列（含起始局面）。
type Timeline struct {
	games []*rules.Game
}

// NewTimeline 由棋譜建立導覽時間軸。
func NewTimeline(rec Record) (*Timeline, error) {
	games, err := Replay(rec)
	if err != nil {
		return nil, err
	}
	return &Timeline{games: games}, nil
}

// Len 回傳盤面數（走法數 + 1）。
func (t *Timeline) Len() int { return len(t.games) }

// At 回傳第 ply 手之後的盤面（ply=0 為起始局面）。
func (t *Timeline) At(ply int) *rules.Game { return t.games[ply] }

// MovesInChinese 將棋譜的 UCCI 走法序列轉成中文記譜清單（順序一致）。
func MovesInChinese(rec Record) ([]string, error) {
	g, err := rules.FromFEN(rec.InitialFEN)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rec.Moves))
	for _, uci := range rec.Moves {
		m, err := board.ParseUCCI(uci)
		if err != nil {
			return nil, err
		}
		out = append(out, g.ToChinese(m)) // 須以走子前盤面轉換
		g, err = g.ApplyMove(m)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
