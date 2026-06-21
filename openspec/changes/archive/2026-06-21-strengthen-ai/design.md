## Context

AI 在 `player/ai.go`：negamax + alpha-beta，純子力評估，深度 1/2/3。引擎 `rules.Game` 不可變，`ApplyMove`/`LegalMoves` 每節點 clone 棋盤並做將軍偵測。`board.Move` 暴露 `From`/`To` 欄位，`g.PieceAt(sq)` 可讀任一格。搜尋於背景 goroutine，可容忍 ~0.5–2s。

## 設計決策

### 位置評估（PST）
- 新增 `pst map[byte][board.NumSquares]int`，以**紅方視角**定義各 kind（r/n/c/b/a/k/p）的格位分。
- 黑方查表時用 rank 鏡射：對黑子在 `sq` 取 `pst[kind][mirror(sq)]`，`mirror = MakeSquare(file, Ranks-1-rank)`。
- 數值規模：約 ±2～30（單位與子力同為「分」，兵=100）。確保位置因素遠小於最小子力差（兵 100），不致使「得子」反轉；`mateScore=1_000_000` 仍壓過一切。
- `evaluate` 既有子力迴圈中，於每子加位置分；維持「行棋方視角」正負號。

### 王安全（簡易、保守）
- 僅加入極小幅度項，且 **King PST 全表為 0**，確保 `TestAIVariesOnRepeatedPosition` 之裸王盤面兩個王步仍同分（落在 ε 內）。
- Tier 1 先以「將帥照面風險」等既有規則內隱含的安全為主，不引入會破壞既有測試平手的項；如需，僅加對稱、不影響該測試的小項。

### 靜默搜尋（quiescence）
- `negamax` 中 `depth==0` 改呼叫 `quiesce(g, alpha, beta, ply)`。
- `quiesce`：
  1. 若 `g.Result().Over` → 回終局分（將死/困斃 `-mateScore+ply`，和棋 0）。
  2. `stand := evaluate(g)`（stand-pat 下限）；`if stand>=beta return stand`；`if stand>alpha alpha=stand`。
  3. 僅對**吃子著手**（`!g.PieceAt(m.To).IsEmpty()`）遞迴 `-quiesce(child,-beta,-alpha,ply+1)`，做 alpha-beta。
  4. 終止性：吃子使盤面棋子數嚴格遞減 → 必收斂；另設 ply 上限（如 ply>64）保險直接回 stand。
- 吃子著手仍須為合法（取自 `g.LegalMoves()` 過濾吃子），避免送將自殺步。

### MVV-LVA 走法排序
- 工具函式 `orderMoves(g, moves)`：吃子優先，鍵 = `victimValue*16 - attackerValue`（`victim=PieceAt(To)`、`attacker=PieceAt(From)` 之 `pieceValue[kind]`）；非吃子鍵為 0；同鍵以 UCCI 字串穩定排序。
- 用於 `negamax` 與 `quiesce` 的著手遍歷；不改變回傳分值，只改順序（alpha-beta 對任何順序回相同 minimax 值）。

### near-best 變招與 SelectMove 重構
- 一次計算每個 root 著手分數（取代目前算兩遍）：`scored := []{move, score}`。
- `best := max(score)`；`bestMoves := {m: score>=best-ε}`，依（分數降、UCCI 升）排序。
- ε = 16（< 兵值 100，故「得一車/將死」必為唯一近佳手；開局多步同分仍同組）。
- 未 Seed：`visits[posKey] % len(bestMoves)` 輪替（fresh→index 0 可重現；重訪→變招）。
- 已 Seed：`pick(len(bestMoves))` 隨機。

### 難度深度
- `Easy=1, Medium=2, Hard=3`；`difficultyLabel`：`<=Easy 簡單`、`>=Hard 困難`、其餘普通。維持 Easy≤Medium≤Hard、Easy≥1。
- **不加深**：實測（現行每節點重算合法走法的引擎）開局單步 depth2≈1.2s、depth3≈15s、depth4≈95s。固定深度無法支撐 3 以上作預設。棋力提升改由評估（PST）與靜默搜尋達成——正對應使用者抱怨（無章法、送子），深度非主因。

## 風險與相容性

- 既有 9 測試：位置權重 < 子力、ε 小、King PST=0 → mate-in-1 / capture-free-rook / unique-best / fresh-reproducible / varies-on-repeat / seeded-varies 皆維持。
- 效能：Medium（預設）depth2 ≈ 1.2s/步，反應良好且明顯強於舊版同深度純子力 AI；Hard depth3 開局較慢（~15s，中殘局快）。
- 後續（Tier 2，另案）：迭代加深＋時間預算與 TT-move/killer 排序，可在固定時間內達更深、並使 Hard 反應化。
- 無 API 變更，`Player`/`SelectMove` 簽章不變。
