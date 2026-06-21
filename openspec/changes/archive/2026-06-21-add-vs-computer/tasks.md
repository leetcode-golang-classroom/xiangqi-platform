## 1. 測試先行（純邏輯控制器）

- [x] 1.1 `Controller` 回合分派（人類/AI/結束）與 `IsAITurn`/`IsHumanTurn`
- [x] 1.2 人類回合接受點擊、AI 回合忽略點擊
- [x] 1.3 `StepAI` 在 AI 回合選步並套用、人類回合為無操作、結束後停止、AI 對 AI 可走子

## 2. 實作

- [x] 2.1 `Session.Play(move)`：直接套用合法走法（供 AI）
- [x] 2.2 `play.Opponent` 介面（解耦，不 import player）
- [x] 2.3 `Controller`：`NewController`/`VsComputer`、`IsAITurn`/`IsHumanTurn`、`HumanTap`、`StepAI`、`CurrentOpponent`/`ApplyAIMove`、存取器

## 3. GUI 串接（建構標籤 ebiten）

- [x] 3.1 人類執紅、電腦執黑；AI 搜尋於背景 goroutine、主迴圈套用、顯示「電腦思考中」
- [x] 3.2 棋譜輸出鍵：存檔（透過 `core/storage`）
- [x] 3.3 `go build -tags ebiten` 編譯通過

## 4. 驗收

- [x] 4.1 `go test ./...` 全綠
- [x] 4.2 `pants test ::` 全綠
- [x] 4.3 `openspec validate add-vs-computer` 通過
