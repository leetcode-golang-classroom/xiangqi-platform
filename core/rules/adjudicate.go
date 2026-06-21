package rules

import "github.com/yuanyu90221/xiangqi-platform/core/board"

// Adjudicate 重放一段走法並裁定特殊結果：
//   - 三次重複局面時，若某一方在其每一步都將軍對方，判該方「長將判負」。
//   - 兩方皆長將或皆非 → 判和（perpetual 互將）。
//   - 無重複局面 → 回傳最終盤面的一般 Result（含將死/困斃/自然限著）。
//
// 局面重複以 positionKey（盤面 + 輪走方）判定，忽略計步。
func Adjudicate(initialFEN string, moves []string) (Result, error) {
	g, err := FromFEN(initialFEN)
	if err != nil {
		return Result{}, err
	}

	seen := map[string]int{g.positionKey(): 1}
	var redMoves, redChecks, blackMoves, blackChecks int
	threefold := false

	for _, mv := range moves {
		m, err := board.ParseUCCI(mv)
		if err != nil {
			return Result{}, err
		}
		mover := g.turn
		g, err = g.ApplyMove(m)
		if err != nil {
			return Result{}, err
		}
		// mover 是否將軍對方（此時 g.turn 為對方）
		gaveCheck := inCheck(g.board, g.turn)
		if mover == board.Red {
			redMoves++
			if gaveCheck {
				redChecks++
			}
		} else {
			blackMoves++
			if gaveCheck {
				blackChecks++
			}
		}

		seen[g.positionKey()]++
		if seen[g.positionKey()] >= 3 {
			threefold = true
			break
		}
	}

	if !threefold {
		return g.Result(), nil
	}

	redPerp := redMoves > 0 && redChecks == redMoves
	blackPerp := blackMoves > 0 && blackChecks == blackMoves
	switch {
	case redPerp && !blackPerp:
		return Result{Over: true, Winner: "black", Reason: "perpetual_check"}, nil
	case blackPerp && !redPerp:
		return Result{Over: true, Winner: "red", Reason: "perpetual_check"}, nil
	default:
		return Result{Over: true, Reason: "repetition_draw"}, nil
	}
}
