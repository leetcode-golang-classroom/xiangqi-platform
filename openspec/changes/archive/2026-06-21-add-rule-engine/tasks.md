## 1. 測試地基（已完成 / 紅燈）

- [x] 1.1 建立 `conformance/*.json` 黃金案例（fen / legalmoves / result / record）
- [x] 1.2 建立 table-driven + testify harness（`conformance/conformance_test.go`）
- [x] 1.3 確認 `go test` 與 `pants test ::` 可驅動測試並呈現紅燈

## 2. 記譜與盤面轉換

- [x] 2.1 實作 `notation.ParseFEN`（FEN → 盤面/輪走/計步）
- [x] 2.2 實作 `notation.EncodeFEN`（盤面 → FEN），使 fen_cases round-trip 轉綠
- [x] 2.3 串接 `rules.FromFEN` / `Game.ToFEN`

## 3. 走法產生（擬合法）

- [x] 3.1 車：直線走/吃
- [x] 3.2 馬：日字 + 蹩馬腿（horse-no-block / horse-leg-blocked 轉綠）
- [x] 3.3 炮：直走 + 隔炮架吃子（cannon-move-and-capture 轉綠）
- [x] 3.4 相象（田字、不過河、塞象眼）、仕士（斜一格、不出宮）
- [x] 3.5 帥將（直一格、不出宮、將帥不照面）、兵卒（過河變、不後退）
- [x] 3.6 `Game.LegalMoves` 彙整各子擬合法走法（rook-on-open-board 轉綠）

## 4. 合法性與勝負判定

- [x] 4.1 將軍偵測（含飛將）
- [x] 4.2 過濾「走後自將」的走法
- [x] 4.3 將死 / 困斃判定（rook-checkmate、startpos-ongoing 轉綠）
- [x] 4.4 長將判負（`rules.Adjudicate` 重複局面 + 連續將軍偵測，adjudicate_cases 轉綠）
- [x] 4.5 自然限著和棋（`NaturalLimitPlies` 步數上限，natural-limit-draw 轉綠）

## 5. 走子、記錄與復盤

- [x] 5.1 實作 `Game.ApplyMove`（不可變，回傳新狀態），更新計步與輪走
- [x] 5.2 `record.Replay`（由起始 FEN 逐步重放，record_cases 轉綠）
- [x] 5.3 實作 `Game.ToChinese`（座標 → 中文縱線記譜，chinese_cases 轉綠）

## 6. 驗收

- [x] 6.1 `go test ./...` 全綠
- [x] 6.2 `pants test ::` 全綠
- [x] 6.3 `openspec validate add-rule-engine` 通過
