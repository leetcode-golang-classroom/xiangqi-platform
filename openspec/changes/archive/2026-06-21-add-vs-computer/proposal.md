## Why

`ai-opponent`（AI 引擎）與 `local-play`（互動控制器 + GUI）已分別完成，但兩者尚未串接：玩家仍只能人對人。要能「人機對戰」，需要一個對局協調層，依輪走方決定該由人類點擊還是 AI 選步並套用，並在 GUI 中讓 AI 那一方自動走子而不凍結畫面。

## What Changes

- 新增 `match` 能力（`core/play`，純邏輯、可單元測試）：
  - `Controller` 持有一個 `Session` 與各方對手指派（人類 = nil，或 AI）。依當前輪走方判斷由誰走子。
  - 人類回合：只接受點擊（`HumanTap`），AI 回合的點擊忽略。
  - AI 回合：`StepAI` 由對手引擎選步並套用；對局結束後停止。
  - 以結構化 `Opponent` 介面（`Name`/`SelectMove`）解耦——`player.AI` 自然滿足，`core/play` 不反向依賴 `player`，避免 import cycle。
- `Session` 新增 `Play(move)`：直接套用一步合法走法（供 AI 程式化走子）。
- GUI（`cmd/xiangqi`）串接：人類執紅、電腦執黑（可調），AI 搜尋於背景 goroutine 進行（讀取不可變盤面快照），完成後於主迴圈套用，避免畫面凍結；顯示「電腦思考中」。

## Capabilities

### New Capabilities
- `match`: 對局協調——依輪走方分派人類或 AI 走子。

### Modified Capabilities
<!-- 無（Session.Play 為新增方法，不變更既有 requirement 行為）-->

## Impact

- 程式碼：`core/play` 新增 `controller.go`、`Session.Play`；`cmd/xiangqi` 串接 AI（建構標籤 `ebiten`）。
- 測試：`core/play` 新增 `Controller` 單元測試（人類回合點擊、AI 回合選步並套用、回合分派、對局結束停止、AI 對 AI 可走子）。
- 相依：`core/play` 以結構化 `Opponent` 介面解耦，不 import `player`；GUI 以 `player.AI` 注入。
