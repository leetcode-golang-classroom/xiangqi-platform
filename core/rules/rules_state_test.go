package rules_test

// rules_state_test.go: 深度驗證 inCheck、將死偵測與「合法走法不可吃將」三項性質。
//
// 策略：以 AI 對 AI（總是取第一個合法走法）產生大量局面，
// 並在每一步用暴力法（枚舉對手所有擬合法目的格）與引擎的 InCheck() 互相核對。

import (
	"fmt"
	"testing"

	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/core/rules"
)

// ─────────────────────────────────────────────
// 暴力法：枚舉對手所有棋子的擬合法目的格，
// 若任何格等於 c 方主將的位置 → 被將軍。
// 只能呼叫公開 API：g.PieceAt / g.Turn 等。
// ─────────────────────────────────────────────

// bruteForceInCheck 以暴力法驗證 c 方是否被將軍。
// 透過 g.LegalMoves() 的反向推論：若 c 方在走完任何一步之後，
// 對手有合法走法可以回頭吃 c 方的王，則 c 方已被將軍。
//
// 注意：我們無法直接呼叫 pseudoTargets（未匯出），
// 因此改用「套用空走法後，對方走法能否吃到王」的等效方式。
// 但這樣就是遞迴 LegalMoves，複雜度太高。
//
// 改用更直接的方式：
// 「c 方被將軍」等價於「目前輪到 c 方，但 c 方的走法裡有零個…」
// 不對，這也需要 LegalMoves。
//
// 正確方式：構造一個輪到「對手」的假 Game，
// 對手的 LegalMoves 若能吃掉位於 kingSq 的棋子，就是被將軍。
// 但我們沒有 SetTurn/clone API。
//
// 最直接可行的方式：
// 在 c 方的每一格盤面上找 c 方主將位置 kingSq，
// 然後在「以對手為輪走方」的局面上執行 LegalMoves，
// 看有沒有走法 To == kingSq。
// 由於沒有 SetTurn，我們用 FromFEN 重建局面並強制換手。
//
// 等效觀察（最直接，無需重建局面）：
// c 方被將軍 ⟺ 對手在上一步棋走完後，c 方的所有走法（若有）中，
// 此局面 InCheck() == true。
// 但 InCheck() 本身就是要驗的目標，不能自我驗證。
//
// ── 最終採用方案 ─────────────────────────────────
// 利用 FromFEN 把目前盤面轉成「輪到對手走」的局面，
// 再取 LegalMoves，若有任何走法的 To == kingSq 且目的格棋子為 c 方主將 → 被將軍。
// 這完全只靠公開 API。

// kingSquareOf 掃描盤面找出 c 方主將的位置。
func kingSquareOf(g *rules.Game, c board.Color) board.Square {
	wantKind := byte('k')
	for s := board.Square(0); s < board.NumSquares; s++ {
		p := g.PieceAt(s)
		if !p.IsEmpty() && p.Kind() == wantKind && p.Color() == c {
			return s
		}
	}
	return board.InvalidSquare
}

// bruteForceInCheck 用暴力法判斷 c 方是否被將軍。
// 做法：枚舉對手每個棋子的擬合法目的格（PseudoTargets，不過濾自將），
// 若任何目的格等於 c 方主將的位置 → 被將軍。
//
// 使用 PseudoTargets 而非 LegalMoves，是因為 LegalMoves 會過濾
// 「走後使己方被將軍」的走法；若攻擊王的走法同時暴露攻擊方自己的王，
// LegalMoves 就會排除它，導致誤判「未被將軍」。
func bruteForceInCheck(g *rules.Game, c board.Color) bool {
	kingSq := kingSquareOf(g, c)
	if kingSq == board.InvalidSquare {
		return false
	}
	atk := c.Opposite()
	for s := board.Square(0); s < board.NumSquares; s++ {
		p := g.PieceAt(s)
		if p.IsEmpty() || p.Color() != atk {
			continue
		}
		for _, to := range g.PseudoTargets(s) {
			if to == kingSq {
				return true
			}
		}
	}
	return false
}

// ─────────────────────────────────────────────
// 測試 1：InCheck 快速實作 vs 暴力法一致性
// ─────────────────────────────────────────────

// TestInCheckConsistencyVsBruteForce 跑 100 局 AI 對 AI（每局最多 150 手），
// 在每一局面比對 g.InCheck() 與 bruteForceInCheck 的結果是否一致。
func TestInCheckConsistencyVsBruteForce(t *testing.T) {
	const (
		numGames = 100
		maxPlies = 150
	)

	type mismatch struct {
		gameIdx int
		ply     int
		fen     string
		fast    bool
		brute   bool
	}
	var mismatches []mismatch
	totalPositions := 0

	for gameIdx := 0; gameIdx < numGames; gameIdx++ {
		g := rules.NewGame()
		for ply := 0; ply < maxPlies; ply++ {
			if g.Result().Over {
				break
			}
			totalPositions++
			c := g.Turn()

			fast := g.InCheck()
			brute := bruteForceInCheck(g, c)

			if fast != brute {
				mismatches = append(mismatches, mismatch{
					gameIdx: gameIdx,
					ply:     ply,
					fen:     g.ToFEN(),
					fast:    fast,
					brute:   brute,
				})
				// 只收集前 10 個差異避免輸出爆炸
				if len(mismatches) >= 10 {
					goto done
				}
			}

			moves := g.LegalMoves()
			if len(moves) == 0 {
				break
			}
			var err error
			g, err = g.ApplyMove(moves[0])
			if err != nil {
				t.Fatalf("game %d ply %d: ApplyMove error: %v", gameIdx, ply, err)
			}
		}
	}

done:
	t.Logf("Checked %d positions across %d AI-vs-AI games", totalPositions, numGames)
	if len(mismatches) > 0 {
		t.Errorf("Found %d InCheck mismatches (fast vs brute-force):", len(mismatches))
		for _, mm := range mismatches {
			t.Errorf("  game=%d ply=%d fast=%v brute=%v\n  FEN: %s",
				mm.gameIdx, mm.ply, mm.fast, mm.brute, mm.fen)
		}
	}
}

// ─────────────────────────────────────────────
// 測試 2：將死偵測正確性（游戲確實結束）
// ─────────────────────────────────────────────

// TestCheckmateDetection 驗證對局結束時 Result() 的原因正確。
//
// 注意：過河兵可無限橫移（每步重置 halfmove），導致部分終局永遠觸發不了自然限著，
// 遊戲不一定會在有限步內結束。本測試只驗證「當遊戲結束時，原因必須正確」；
// 未在 maxPlies 內結束的局面只記錄、不視為 bug。
func TestCheckmateDetection(t *testing.T) {
	const (
		numGames = 100
		maxPlies = 500
	)

	validReasons := map[string]bool{
		"checkmate":     true,
		"stalemate":     true,
		"natural_limit": true,
	}
	ended, ongoing := 0, 0

	for gameIdx := 0; gameIdx < numGames; gameIdx++ {
		g := rules.NewGame()
		gameEnded := false
		for ply := 0; ply < maxPlies; ply++ {
			r := g.Result()
			if r.Over {
				// 驗證原因合理
				if !validReasons[r.Reason] {
					t.Errorf("game %d ended with unknown reason %q at ply %d, FEN: %s",
						gameIdx, r.Reason, ply, g.ToFEN())
				}
				// 若是將死，驗證輪走方確實被將且無合法走法
				if r.Reason == "checkmate" {
					if !g.InCheck() {
						t.Errorf("game %d: checkmate but InCheck()=false, FEN: %s",
							gameIdx, g.ToFEN())
					}
					if len(g.LegalMoves()) != 0 {
						t.Errorf("game %d: checkmate but LegalMoves non-empty (len=%d), FEN: %s",
							gameIdx, len(g.LegalMoves()), g.ToFEN())
					}
				}
				// 若是困斃，驗證輪走方未被將但無合法走法
				if r.Reason == "stalemate" {
					if g.InCheck() {
						t.Errorf("game %d: stalemate but InCheck()=true, FEN: %s",
							gameIdx, g.ToFEN())
					}
					if len(g.LegalMoves()) != 0 {
						t.Errorf("game %d: stalemate but LegalMoves non-empty (len=%d), FEN: %s",
							gameIdx, len(g.LegalMoves()), g.ToFEN())
					}
				}
				gameEnded = true
				ended++
				break
			}
			moves := g.LegalMoves()
			if len(moves) == 0 {
				t.Errorf("game %d ply %d: LegalMoves empty but Result().Over=false, FEN: %s",
					gameIdx, ply, g.ToFEN())
				gameEnded = true
				break
			}
			idx := (ply*31 + gameIdx*17) % len(moves)
			var err error
			g, err = g.ApplyMove(moves[idx])
			if err != nil {
				t.Fatalf("game %d ply %d: ApplyMove error: %v", gameIdx, ply, err)
			}
		}
		if !gameEnded {
			ongoing++ // 過河兵橫移等情況可能導致無限迴圈，不視為 bug
		}
	}
	t.Logf("CheckmateDetection: %d ended, %d still ongoing after %d plies", ended, ongoing, maxPlies)
}

// TestHalfmoveResetInfiniteLoop 記錄「恆走 moves[0]」策略的已知局限：
// 當黑方兵每回合移動（重置 halfmove 計數器），遊戲可能永不觸發自然限著。
// 此測試驗證該場景確實存在，且規則引擎在此情境下行為正確（不是 bug）：
// 每當兵移動，halfmove 確實被重置為 0（這是正確的規則實作）。
func TestHalfmoveResetInfiniteLoop(t *testing.T) {
	// 已知會導致無限迴圈的 FEN（恆走 moves[0] 策略）
	stuckFEN := "rnbakabnr/9/1c5c1/8p/9/9/8P/1C5CB/4A4/5K1p1 w - - 0 151"
	g, err := rules.FromFEN(stuckFEN)
	if err != nil {
		t.Fatalf("FromFEN: %v", err)
	}

	// 驗證此局面本身是合法的（非遊戲結束）
	r := g.Result()
	if r.Over {
		t.Fatalf("Expected game not over, got %+v", r)
	}

	// 走 4 步，觀察 halfmove 計數器在 FEN 中的變化
	// 預期：紅王走（非兵非吃子）→ halfmove 累加；黑兵移動 → halfmove 歸零
	type step struct {
		move        string
		fenContains string // 期望 FEN 中含有此 halfmove 值
		isPawn      bool
	}
	g1, _ := g.ApplyMove(g.LegalMoves()[0])   // 紅王走：f0e0，halfmove 應從 0→1
	g2, _ := g1.ApplyMove(g1.LegalMoves()[0]) // 黑兵走：h0i0，halfmove 應從 1→0（兵走重置）
	g3, _ := g2.ApplyMove(g2.LegalMoves()[0]) // 紅王走：e0f0，halfmove 應從 0→1
	g4, _ := g3.ApplyMove(g3.LegalMoves()[0]) // 黑兵走：i0h0，halfmove 應從 1→0

	// 驗證 FEN 中的 halfmove 欄位（FEN 格式：pieces turn - - halfmove fullmove）
	checkHalfmove := func(label string, game *rules.Game, wantHalfmove string) {
		fen := game.ToFEN()
		// 找到 "- - " 後的數字
		found := false
		i := 0
		dashCount := 0
		for i < len(fen) {
			if fen[i] == '-' {
				dashCount++
				if dashCount == 2 {
					i += 2 // 跳過 "- "
					end := i
					for end < len(fen) && fen[end] != ' ' {
						end++
					}
					got := fen[i:end]
					if got != wantHalfmove {
						t.Errorf("%s: halfmove in FEN=%q, want %q (full FEN: %s)",
							label, got, wantHalfmove, fen)
					}
					found = true
					break
				}
			}
			i++
		}
		if !found {
			t.Errorf("%s: could not parse halfmove from FEN: %s", label, fen)
		}
	}

	checkHalfmove("after red king move", g1, "1")
	checkHalfmove("after black pawn move (should reset)", g2, "0")
	checkHalfmove("after red king move again", g3, "1")
	checkHalfmove("after black pawn move again (should reset)", g4, "0")

	t.Logf("Confirmed: black pawn moves perpetually reset halfmove counter.")
	t.Logf("This is correct rule engine behavior. 'moves[0]' AI strategy creates infinite games.")
	t.Logf("Stuck FEN: %s", stuckFEN)
}

// ─────────────────────────────────────────────
// 測試 3：合法走法不可吃將（深度驗證）
// ─────────────────────────────────────────────

// TestNoLegalMoveCapuresKing 在 AI 對 AI 的每一局面，
// 驗證 LegalMoves 中沒有任何走法可以直接吃掉對手主將。
func TestNoLegalMoveCapuresKing(t *testing.T) {
	const (
		numGames = 100
		maxPlies = 150
	)

	type kingCapture struct {
		gameIdx int
		ply     int
		move    string
		fen     string
	}
	var captures []kingCapture
	totalPositions := 0

	for gameIdx := 0; gameIdx < numGames; gameIdx++ {
		g := rules.NewGame()
		for ply := 0; ply < maxPlies; ply++ {
			if g.Result().Over {
				break
			}
			totalPositions++

			for _, m := range g.LegalMoves() {
				victim := g.PieceAt(m.To)
				if !victim.IsEmpty() && victim.Kind() == 'k' {
					captures = append(captures, kingCapture{
						gameIdx: gameIdx,
						ply:     ply,
						move:    m.String(),
						fen:     g.ToFEN(),
					})
					if len(captures) >= 10 {
						goto doneCapture
					}
				}
			}

			moves := g.LegalMoves()
			if len(moves) == 0 {
				break
			}
			var err error
			g, err = g.ApplyMove(moves[0])
			if err != nil {
				t.Fatalf("game %d ply %d: ApplyMove error: %v", gameIdx, ply, err)
			}
		}
	}

doneCapture:
	t.Logf("Checked %d positions for king-capture legality across %d games", totalPositions, numGames)
	if len(captures) > 0 {
		t.Errorf("Found %d position(s) where a legal move captures the king:", len(captures))
		for _, kc := range captures {
			t.Errorf("  game=%d ply=%d move=%s\n  FEN: %s", kc.gameIdx, kc.ply, kc.move, kc.fen)
		}
	}
}

// ─────────────────────────────────────────────
// 測試 4：特殊局面的 inCheck 精確驗證
// ─────────────────────────────────────────────

// TestInCheckSpecificPositions 用已知局面精確驗證 InCheck / 被將判斷。
func TestInCheckSpecificPositions(t *testing.T) {
	cases := []struct {
		name      string
		fen       string
		wantCheck bool // 是否輪走方被將
	}{
		{
			// 開局盤面，紅先走，不被將
			name:      "initial_position_no_check",
			fen:       "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1",
			wantCheck: false,
		},
		{
			// 黑炮無砲架，不能將紅帥（炮在 e5，紅帥在 e0，中間無遮擋）
			name:      "cannon_check_on_red_king_no_platform",
			fen:       "4k4/9/9/9/4c4/9/9/9/9/4K4 w - - 0 1",
			wantCheck: false,
		},
		{
			// 黑炮有砲架（紅炮架）將紅帥
			// 黑炮在 e7，紅仕在 e2，紅帥在 e0 → 炮架=仕，炮吃帥
			name:      "cannon_check_with_platform",
			fen:       "4k4/9/4c4/9/9/9/9/4A4/9/4K4 w - - 0 1",
			wantCheck: true,
		},
		{
			// 紅車在同橫線將黑將
			// 車在 a9，黑將在 e9
			name:      "rook_check_on_black_king",
			fen:       "R3k4/9/9/9/9/9/9/9/9/4K4 b - - 0 1",
			wantCheck: true,
		},
		{
			// 馬在 f0（file=5, rank=0），紅帥在 d0（file=3, rank=0）
			// 兩王不對面（黑將在 e9，紅帥在 d0，不同縱線）
			// 黑馬在 f0，跳到 d0 需要 df=-2,dr=0（非日字步法）→ 不能將
			// 且腿方向全部越界或被遮，應無將軍
			name: "horse_at_f0_cannot_check_d0",
			fen:  "4k4/9/9/9/9/9/9/9/9/3Kn4 w - - 0 1",
			// 黑馬 f0(5,0) 日字跳目標：(4,2),(6,2),(3,1),(7,1)等
			// 沒有一個是 d0(3,0)；且兩王不同縱線（黑將在 e9，紅帥在 d0）
			wantCheck: false,
		},
		{
			// 馬在 f2，帥在 e0：df=-1 dr=-2，leg=(f+0,2-1)=(f,1) 若空則可到 e0
			name:      "horse_check_f2_to_e0",
			fen:       "4k4/9/9/9/9/9/9/9/5n3/4K4 w - - 0 1",
			wantCheck: true,
		},
		{
			// 炮沒有砲架不能將
			name:      "cannon_no_platform_no_check",
			fen:       "4k4/9/9/9/9/4c4/9/9/9/4K4 w - - 0 1",
			wantCheck: false,
		},
		{
			// 紅兵過河橫向將黑將：黑將在 e9，紅兵在 d9（過河 rank≥5，可橫走）
			name:      "pawn_crossed_lateral_check_on_black_king",
			fen:       "3Pk4/9/9/9/9/9/9/9/9/4K4 b - - 0 1",
			wantCheck: true,
		},
		{
			// 黑卒在 d0 橫向攻擊紅帥在 e0（黑卒已過河 rank≤4，可橫走）
			name:      "pawn_crossed_lateral_d0_to_e0",
			fen:       "4k4/9/9/9/9/9/9/9/9/3pK4 w - - 0 1",
			wantCheck: true, // 黑卒 d0 橫走到 e0 = 紅帥位置
		},
		{
			// 黑卒在 e1 向下（rank-1）將紅帥在 e0
			name:      "black_pawn_forward_check_e1_to_e0",
			fen:       "4k4/9/9/9/9/9/9/9/9/4pK3 w - - 0 1",
			wantCheck: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g, err := rules.FromFEN(tc.fen)
			if err != nil {
				t.Fatalf("FromFEN error: %v", err)
			}
			got := g.InCheck()
			if got != tc.wantCheck {
				t.Errorf("InCheck()=%v, want %v (FEN: %s)", got, tc.wantCheck, tc.fen)
				// 同時跑暴力法提供額外診斷
				brute := bruteForceInCheck(g, g.Turn())
				t.Errorf("  bruteForceInCheck=%v (for cross-reference)", brute)
			}
		})
	}
}

// ─────────────────────────────────────────────
// 測試 4b：馬將軍的 inCheck Bug 回歸測試
// ─────────────────────────────────────────────

// TestHorseCheckBugRegression 針對已知的馬蹩馬腿方向判斷錯誤進行回歸測試。
//
// BUG 說明（movegen.go inCheck 函數）：
//
//	inCheck 使用反向掃描法判斷馬是否將軍。
//	對於每個可能攻擊主將的馬位置，它檢查「蹩馬腿」是否被封鎖。
//	但 horseCands 中的蹩馬腿偏移量 (lf, lr) 使用的是「從主將位置出發」的方向，
//	與正確的「從馬的位置出發」的腿方向不符。
//
//	正向：馬在 H=(hf,hr)，跳到 T=H+(df,dr)，腿在 H+(lf,lr)。
//	逆向（從主將 K=T 找馬）：馬在 H=K+(-df,-dr)，腿在 H+(lf,lr)=K+(lf-df, lr-dr)。
//	但 inCheck 錯誤地將腿設在 K+(lf, lr)（沒有 -df/-dr 修正）。
//
//	已驗證的錯誤局面：
//	FEN: 4k1b1c/4a4/5a3/2C5r/P1p1P4/9/2Cp1n3/N3B4/4Kc1N1/3A1AB2 w - - 1 64
//	黑馬在 f3，腿在 f2（空），可將紅帥 e1。
//	但 inCheck 錯誤地檢查腿在 e2（有紅象），誤判為不能將。
func TestHorseCheckBugRegression(t *testing.T) {
	cases := []struct {
		name      string
		fen       string
		wantCheck bool
		note      string
	}{
		{
			// 已從 AI 對 AI 遊戲中發現的真實 bug 局面：
			// 黑馬在 f3(5,3)，腿 f2(5,2) 空，可將紅帥 e1(4,1)
			// 但 inCheck 錯誤地檢查腿 e2(4,2)（有紅象 B），誤認為被蹩
			name:      "horse_f3_checks_king_e1_leg_f2_bug",
			fen:       "4k1b1c/4a4/5a3/2C5r/P1p1P4/9/2Cp1n3/N3B4/4Kc1N1/3A1AB2 w - - 1 64",
			wantCheck: true, // 黑馬 f3 透過 leg f2（空）將紅帥 e1
			note:      "inCheck bug: checks leg e2 (has elephant) instead of f2 (empty)",
		},
		{
			// 第二個 bug 局面
			name:      "horse_f3_checks_king_e1_variant2",
			fen:       "4k1b1c/4a4/5a3/2C5r/1Pp1P4/9/2C2n3/N2pB4/4Kc1N1/3A1AB2 w - - 0 65",
			wantCheck: true,
			note:      "inCheck bug: same horse-leg mismatch",
		},
		{
			// 最小復現：
			// 紅帥在 d1(3,1)，紅象在 d2(3,2)（象在 inCheck 錯誤腿位置），
			// 黑馬在 e3(4,3)，腿應在 e2(4,2)（空），可將帥 d1。
			// inCheck 錯誤地檢查腿在 d2（有紅象）→ 誤判為蹩馬腿，漏將。
			// 兩王不同縱線（黑將在 b9，紅帥在 d1）→ 無飛將干擾。
			// FEN rows (rank9..rank0):
			//   rank9: 1k7 → k at b9(1,9)
			//   rank3: 4n4 → n at e3(4,3)
			//   rank2: 3B5 → B at d2(3,2)
			//   rank1: 3K5 → K at d1(3,1)
			name: "minimal_horse_d1_e3_leg_e2_empty",
			fen:  "1k7/9/9/9/9/9/4n4/3B5/3K5/9 w - - 0 1",
			// 黑馬 e3(4,3) → d1(3,1)：df=-1,dr=-2，腿=horse+(0,-1)=(4,2)=e2（空）
			// inCheck 從 d1(3,1) 查 horseCand[3]={-1,-2,0,-1}：horse=d1+(-1,-2)=(2,-1) invalid
			// horseCand[0]={1,2,0,1}：horse=(4,3)=e3 ✓，leg=d1+(0,1)=(3,2)=d2 → 有B → 誤認蹩馬腿
			// 正確腿應在 e2(4,2)，不在 d2(3,2)。
			wantCheck: true,
			note:      "minimal: horse e3 checks king d1 via leg e2; inCheck wrongly checks d2 (has elephant)",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			g, err := rules.FromFEN(tc.fen)
			if err != nil {
				t.Fatalf("FromFEN error (%s): %v", tc.note, err)
			}
			fast := g.InCheck()
			brute := bruteForceInCheck(g, g.Turn())

			// Primary assertion: fast should equal brute
			if fast != brute {
				t.Errorf("BUG CONFIRMED: fast InCheck=%v, brute=%v\n  FEN: %s\n  Note: %s",
					fast, brute, tc.fen, tc.note)
			}
			// Secondary assertion: expected value
			if fast != tc.wantCheck {
				t.Errorf("InCheck()=%v, want=%v\n  FEN: %s\n  Note: %s",
					fast, tc.wantCheck, tc.fen, tc.note)
			}
		})
	}
}

// ─────────────────────────────────────────────
// 測試 5：大量隨機漫步局面的暴力核對
// ─────────────────────────────────────────────

// TestInCheckBruteForceDeepGames 用更多局（200 局）深層核對 InCheck 正確性，
// 以提升信心，特別針對可能在 BFS 廣搜中漏掉的深層局面。
func TestInCheckBruteForceDeepGames(t *testing.T) {
	const (
		numGames = 200
		maxPlies = 200
	)

	mismatchCount := 0
	totalPositions := 0

	for gameIdx := 0; gameIdx < numGames; gameIdx++ {
		g := rules.NewGame()
		// 根據 gameIdx 選不同起手以增加多樣性
		// 前幾步根據 gameIdx 做 mod 選擇
		for ply := 0; ply < maxPlies; ply++ {
			if g.Result().Over {
				break
			}
			totalPositions++
			c := g.Turn()

			fast := g.InCheck()
			brute := bruteForceInCheck(g, c)

			if fast != brute {
				mismatchCount++
				t.Errorf("InCheck mismatch at game=%d ply=%d: fast=%v brute=%v\n  FEN: %s",
					gameIdx, ply, fast, brute, g.ToFEN())
				if mismatchCount >= 5 {
					t.Logf("Stopping after 5 mismatches. Checked %d positions.", totalPositions)
					return
				}
			}

			moves := g.LegalMoves()
			if len(moves) == 0 {
				break
			}
			// 選走法：用 ply 和 gameIdx 做偽隨機，增加多樣性
			idx := (ply*31 + gameIdx*17) % len(moves)
			var err error
			g, err = g.ApplyMove(moves[idx])
			if err != nil {
				t.Fatalf("game %d ply %d: ApplyMove error: %v", gameIdx, ply, err)
			}
		}
	}

	t.Logf("Deep brute-force check: %d positions, %d mismatches", totalPositions, mismatchCount)
}

// ─────────────────────────────────────────────
// 測試 6：飛將（兩將對面）偵測
// ─────────────────────────────────────────────

// TestFlyingKingCheck 驗證飛將（兩將在同縱線無子間隔）的 InCheck 偵測。
func TestFlyingKingCheck(t *testing.T) {
	cases := []struct {
		name      string
		fen       string
		turn      string // "w" 或 "b"，哪方被飛將
		wantCheck bool
	}{
		{
			// 兩將同縱線，中間無子，輪紅走 → 紅帥被黑將飛將
			name:      "flying_king_red_in_check",
			fen:       "4k4/9/9/9/9/9/9/9/9/4K4 w - - 0 1",
			wantCheck: false, // 距離 9 格，中間全空，但此時輪紅走 → 紅被黑將（飛將）？
			// 等等：飛將方向：黑將 e9 朝下掃描到紅帥 e0，中間無子 → 紅帥被將
			// inCheck(board, Red) 應為 true
			// 測試 g.InCheck()，g.Turn()=Red，所以 InCheck() = inCheck(board, Red)
		},
	}

	// 修正 flying_king 的期望值（重新分析）
	// FEN "4k4/9/9/9/9/9/9/9/9/4K4 w - - 0 1"：
	//   rank0: 4K4 → 紅帥在 e0（file=4, rank=0）
	//   rank9: 4k4 → 黑將在 e9（file=4, rank=9）
	//   兩者同縱線（file=4），中間 rank1..rank8 全空
	//   inCheck(board, Red)：從紅帥 e0 向上掃（d[0]=0, d[1]=1），第一個棋子是黑將 → 飛將 → true
	// 所以 wantCheck 應為 true。

	for _, tc := range cases {
		_ = tc // 上面有 wantCheck=false 錯誤，改在下面直接寫精確測試
	}

	t.Run("flying_king_both_kings_face_to_face", func(t *testing.T) {
		// 兩將同縱線，中間無子：飛將規則下紅帥被將
		// 注意：brute-force 方法無法偵測飛將（因為黑將「走」到紅帥位置不是合法走法）
		// 此測試只驗證 fast InCheck() 的行為。
		fen := "4k4/9/9/9/9/9/9/9/9/4K4 w - - 0 1"
		g, err := rules.FromFEN(fen)
		if err != nil {
			t.Fatalf("FromFEN: %v", err)
		}
		got := g.InCheck()
		t.Logf("flying_king face-to-face: fast InCheck=%v (expected: true)", got)
		// InCheck() should return true: two kings face each other on file e, no piece between
		if !got {
			t.Errorf("Expected InCheck()=true for flying king position, got false. FEN: %s", fen)
		}
		// Verify that LegalMoves is empty (king is in flying-king check, all moves must escape)
		// or has moves that fix the flying king.
		t.Logf("LegalMoves count in flying king position: %d", len(g.LegalMoves()))
	})

	t.Run("flying_king_blocked_by_piece", func(t *testing.T) {
		// 兩將同縱線，中間有一仕 → 不是飛將
		fen := "4k4/9/9/9/9/4A4/9/9/9/4K4 w - - 0 1"
		g, err := rules.FromFEN(fen)
		if err != nil {
			t.Fatalf("FromFEN: %v", err)
		}
		got := g.InCheck()
		brute := bruteForceInCheck(g, g.Turn())
		t.Logf("flying_king_blocked: InCheck=%v, bruteForce=%v", got, brute)
		if got != brute {
			t.Errorf("Mismatch: fast InCheck=%v, brute=%v  FEN: %s", got, brute, fen)
		}
		// 此局面輪走方（紅）不應被將
		if got {
			t.Errorf("Expected no check (blocked), but InCheck()=true")
		}
	})
}

// ─────────────────────────────────────────────
// 測試 7：inCheck 暴力法的自我驗證（sanity check）
// ─────────────────────────────────────────────

// TestBruteForceInCheckSanity 驗證 bruteForceInCheck 本身的正確性：
// 在一個已知被將的局面，它應回傳 true；已知未被將，回傳 false。
func TestBruteForceInCheckSanity(t *testing.T) {
	t.Run("should_detect_cannon_check", func(t *testing.T) {
		// 炮隔一子將：黑炮在 e7，e3 有紅仕，紅帥在 e0
		fen := "4k4/9/9/4c4/9/9/9/4A4/9/4K4 w - - 0 1"
		g, _ := rules.FromFEN(fen)
		if !bruteForceInCheck(g, board.Red) {
			t.Error("bruteForceInCheck should return true for cannon check, got false")
		}
	})

	t.Run("should_not_detect_check_in_start", func(t *testing.T) {
		g := rules.NewGame()
		if bruteForceInCheck(g, board.Red) {
			t.Error("bruteForceInCheck should return false at start position for red, got true")
		}
		if bruteForceInCheck(g, board.Black) {
			t.Error("bruteForceInCheck should return false at start position for black, got true")
		}
	})
}

// ─────────────────────────────────────────────
// 輔助：列印統計
// ─────────────────────────────────────────────

func TestPrintGameStats(t *testing.T) {
	const numGames = 50
	var checkmates, stalemates, naturalLimits, ongoing int

	for i := 0; i < numGames; i++ {
		g := rules.NewGame()
		for ply := 0; ply < 300; ply++ {
			r := g.Result()
			if r.Over {
				switch r.Reason {
				case "checkmate":
					checkmates++
				case "stalemate":
					stalemates++
				case "natural_limit":
					naturalLimits++
				}
				goto nextGame
			}
			moves := g.LegalMoves()
			if len(moves) == 0 {
				break
			}
			idx := (ply*31 + i*17) % len(moves)
			var err error
			g, err = g.ApplyMove(moves[idx])
			if err != nil {
				break
			}
		}
		ongoing++
	nextGame:
	}

	fmt.Printf("Game outcomes over %d games: checkmate=%d stalemate=%d natural_limit=%d ongoing=%d\n",
		numGames, checkmates, stalemates, naturalLimits, ongoing)
}
