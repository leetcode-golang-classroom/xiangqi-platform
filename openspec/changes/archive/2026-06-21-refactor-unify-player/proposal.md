## Why

目前對手抽象不對稱：AI/遠端是「著手來源」（`SelectMove`），人類則是 `Session` 內建的點選互動 + `Controller` 的 nil-vs-Opponent 分派（`IsAITurn`/`IsHumanTurn`）。人類與 AI 在對局迴圈的角色其實相同——都只是「決定下一步的人」，差別僅在**如何決定**（人類漸進互動、AI 搜尋、遠端等待）。統一為單一 `Player` 抽象可讓對局迴圈一致（人機／AI對AI／日後遠端共用），移除分派分支，並符合原始設計願景。

## What Changes

- 統一 `Player` 介面為**非同步取步**：`Name()` + `RequestMove(game) <-chan Move`。AI 於背景 goroutine 搜尋後送出；人類於 UI 完成選子後送出；（日後）遠端於 WebSocket 回傳後送出。
- 將點選互動狀態機（選子／高亮／改選／取消）由 `Session` 抽出為 `Human`（實作 `Player`）；`Session` 收斂為純對局狀態（走子歷史、記譜、悔棋、認輸、`Play`）。
- `Controller` 改為**統一迴圈**：每幀 `Step()` 向當前 `Player` 請求一步、完成即套用；移除 `IsAITurn`/`IsHumanTurn`/`StepAI`/`HumanTap` 分派；保留對局層級的 `Undo`/`Resign`。
- AI 由同步 `SelectMove` 之上補 `RequestMove`（背景搜尋）；GUI 改用統一迴圈，點擊餵給當前 `Human`。
- **術語統一**：`Player`/`Interactive` 介面與 `TapResult` 移入 `player` 套件，與 `Human`/`AI` 實作同處；`core/play`（`Session`+`Controller`）單向依賴 `player`。消除「`play.Player` 介面 vs `player` 套件」的詞義重疊。
- GUI 支援**自由選邊**：人類可執紅或執黑（`VsComputer` 已接受 `humanColor`），棋盤自動翻轉使己方永遠在下方。

## Capabilities

### Modified Capabilities
- `local-play`: 點選互動狀態機由 `Session` 移至 `Human`（行為保留、位置調整）；`Session` 收斂為對局狀態。
- `ai-opponent`: `Player` 介面改為非同步 `RequestMove`；AI 補背景取步。
- `match`: `Controller` 改為統一迴圈，對手不分人類/AI 一致處理。

## Impact

- 程式碼：`core/play` 拆出 `human.go`、`session.go` 收斂、`controller.go` 重寫；`player/ai.go` 補 `RequestMove`；`cmd/xiangqi` 改用統一迴圈。
- 測試：新增 `Human` 單元測試；改寫 `Session`/`Controller` 測試；AI 補 `RequestMove` 測試。行為（合法性、記譜、勝負）不變。
- 介面契約：`docs/design` 對手抽象與循序圖更新。
