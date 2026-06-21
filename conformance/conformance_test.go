// 跨語言一致性測試 harness（Go 實作）。
//
// 設計：table-driven + testify。
// 測試表（table）來自語言中立的 conformance/*.json 黃金案例，
// 每個案例以 t.Run 子測試執行，確保跨語言實作一致。
// 規則引擎尚未實作前，這些測試預期為「紅燈」；實作後應全綠。
package conformance_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuanyu90221/xiangqi-platform/conformance"
	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/core/record"
	"github.com/yuanyu90221/xiangqi-platform/core/rules"
)

func TestFENRoundTrip(t *testing.T) {
	table, err := conformance.FENCases()
	require.NoError(t, err, "載入測試表")

	for _, tc := range table {
		t.Run(tc.Name, func(t *testing.T) {
			g, err := rules.FromFEN(tc.FEN)
			require.NoError(t, err, "FromFEN(%q)", tc.FEN)
			assert.Equal(t, tc.FEN, g.ToFEN(), "FEN round-trip")
		})
	}
}

func TestLegalMoves(t *testing.T) {
	table, err := conformance.LegalMovesCases()
	require.NoError(t, err, "載入測試表")

	for _, tc := range table {
		t.Run(tc.Name, func(t *testing.T) {
			g, err := rules.FromFEN(tc.FEN)
			require.NoError(t, err, "FromFEN(%q)", tc.FEN)
			from, err := board.ParseSquare(tc.From)
			require.NoError(t, err, "ParseSquare(%q)", tc.From)

			var got []string
			for _, m := range g.LegalMoves() {
				if m.From == from {
					got = append(got, m.String())
				}
			}
			assert.ElementsMatch(t, tc.Expected, got, "從 %s 出發的合法走法", tc.From)
		})
	}
}

func TestResult(t *testing.T) {
	table, err := conformance.ResultCases()
	require.NoError(t, err, "載入測試表")

	for _, tc := range table {
		t.Run(tc.Name, func(t *testing.T) {
			g, err := rules.FromFEN(tc.FEN)
			require.NoError(t, err, "FromFEN(%q)", tc.FEN)
			assert.Equal(t, tc.Result, conformance.ResultExpect{
				Over:   g.Result().Over,
				Winner: g.Result().Winner,
				Reason: g.Result().Reason,
			}, "對局結果")
		})
	}
}

func TestChineseNotation(t *testing.T) {
	table, err := conformance.ChineseCases()
	require.NoError(t, err, "載入測試表")

	for _, tc := range table {
		t.Run(tc.Name, func(t *testing.T) {
			g, err := rules.FromFEN(tc.FEN)
			require.NoError(t, err, "FromFEN(%q)", tc.FEN)
			m, err := board.ParseUCCI(tc.Move)
			require.NoError(t, err, "ParseUCCI(%q)", tc.Move)
			assert.Equal(t, tc.Chinese, g.ToChinese(m), "中文記譜")
		})
	}
}

func TestAdjudicate(t *testing.T) {
	table, err := conformance.AdjudicateCases()
	require.NoError(t, err, "載入測試表")

	for _, tc := range table {
		t.Run(tc.Name, func(t *testing.T) {
			r, err := rules.Adjudicate(tc.InitialFEN, tc.Moves)
			require.NoError(t, err, "Adjudicate")
			assert.Equal(t, tc.Result, conformance.ResultExpect{
				Over:   r.Over,
				Winner: r.Winner,
				Reason: r.Reason,
			}, "裁定結果")
		})
	}
}

func TestMoveList(t *testing.T) {
	table, err := conformance.MovelistCases()
	require.NoError(t, err, "載入測試表")

	for _, tc := range table {
		t.Run(tc.Name, func(t *testing.T) {
			rec := record.Record{InitialFEN: tc.InitialFEN, Moves: tc.Moves}
			got, err := record.MovesInChinese(rec)
			require.NoError(t, err, "MovesInChinese")
			assert.Equal(t, tc.Chinese, got, "中文記譜清單")
		})
	}
}

func TestRecordReplay(t *testing.T) {
	table, err := conformance.RecordCases()
	require.NoError(t, err, "載入測試表")

	for _, tc := range table {
		t.Run(tc.Name, func(t *testing.T) {
			g, err := rules.FromFEN(tc.InitialFEN)
			require.NoError(t, err, "FromFEN(%q)", tc.InitialFEN)
			for _, mv := range tc.Moves {
				m, err := board.ParseUCCI(mv)
				require.NoError(t, err, "ParseUCCI(%q)", mv)
				g, err = g.ApplyMove(m)
				require.NoError(t, err, "ApplyMove(%q)", mv)
			}
			assert.Equal(t, tc.FinalFEN, g.ToFEN(), "重放後盤面")
		})
	}
}
