## Why

階段 1 的單機雙人對弈已可玩。階段 2 要讓玩家能與電腦對戰：需要一個對手抽象，使對局流程不在意對手是人類、AI 或（日後）遠端；並提供純 Go 的 AI 引擎（搜尋 + 評估）與難度分級。AI 重用既有 `rule-engine`（走法產生、合法性、終局判定），無外部引擎相依，便於跨平台與離線運作。

## What Changes

- 新增 `ai-opponent` 能力（`player` 套件，純 Go）：
  - `Player` 抽象介面：`SelectMove(game) -> move`，統一「本地人類／AI／遠端」三種對手；對局迴圈只依賴此介面。
  - `AI` 實作 `Player`：以 negamax + alpha-beta 剪枝搜尋至指定深度，評估函數以子力價值為主（車/馬/炮/仕/相/兵）加上過河兵加成；終局（將死/困斃）回傳必勝/必負分值（含步數，偏好較快將死）、和棋為 0。
  - 難度分級：以搜尋深度表示（Easy/Medium/Hard）。
  - 搜尋具決定性（穩定走法順序、固定 tie-break），相同盤面與深度輸出固定走法，便於單元測試。
- 為評估提供唯讀盤面存取：`rule-engine` 新增 `Game.PieceAt(square)`（唯讀，不改變不可變性與既有行為）。

## Capabilities

### New Capabilities
- `ai-opponent`: 對手抽象介面與純 Go AI 引擎（搜尋、評估、難度）。

### Modified Capabilities
<!-- 無（PieceAt 為唯讀存取器，不變更既有 requirement 行為）-->

## Impact

- 程式碼：新增 `player`（`player.go`、`ai.go`）；`core/rules` 新增唯讀 `PieceAt`。
- 測試：`player` 單元測試（吃子取優、一步將死、起手回合回傳合法走法、終局回報錯誤、難度→深度）。
- 一致性：AI 走法取決於評估與排序，屬實作層決策，**不**納入語言中立 `conformance/`（不同實作的最佳手可不同）；以 Go 單元測試驗證。
- 相依：`core/rules`、`core/board`；無外部引擎、無圖形相依，進入預設建構/測試。
