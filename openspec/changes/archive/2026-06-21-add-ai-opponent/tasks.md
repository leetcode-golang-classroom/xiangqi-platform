## 1. 測試先行（純邏輯，table-driven + testify）

- [x] 1.1 `Player` 介面與終局回報錯誤
- [x] 1.2 AI：吃無保護子、一步將死、起手回傳合法走法、難度→深度

## 2. 實作

- [x] 2.1 `rule-engine` 新增唯讀 `Game.PieceAt(square)`
- [x] 2.2 `player.Player` 介面 + `AI`（negamax + alpha-beta）
- [x] 2.3 評估函數（子力價值 + 過河兵加成 + 終局必勝負）、難度（Easy/Medium/Hard → 深度）
- [x] 2.4 `player/BUILD`（pants tailor）

## 3. 驗收

- [x] 3.1 `go test ./...` 全綠
- [x] 3.2 `pants test ::` 全綠
- [x] 3.3 `openspec validate add-ai-opponent` 通過
