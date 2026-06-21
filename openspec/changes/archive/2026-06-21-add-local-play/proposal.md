## Why

階段 1 的純邏輯核心（rule-engine、game-record、storage）已齊備，但尚無「對局互動」層把它們串成可玩流程。本機雙人對弈需要一個控制器，將「玩家點擊棋格」轉成選子、合法點高亮、走子、悔棋、認輸與結束判定，並邊下邊記譜。此控制器是單機 App 與（日後）線上前端共用的互動核心。

## What Changes

- 新增 `local-play` 能力，以**純邏輯互動控制器** `Session`（`core/play`，不依賴任何圖形庫，可單元測試）實作：
  - 點選互動狀態機：未選子時點己方可動子→選取並高亮其合法落點；已選子時點合法落點→走子並記譜、點另一己方可動子→改選、點他處→取消。
  - 悔棋（回退上一手）、認輸（當前方認輸、對方勝）、結束判定（將死/困斃/和棋沿用 rule-engine）。
  - 隨時輸出 `xiangqi-record-v1` 棋譜（可交由 `storage` 持久化）。
- 新增 **Ebiten 棋盤渲染層**（`cmd/xiangqi`，置於 `//go:build ebiten` 標籤後）：繪製棋盤/棋子（CJK 字型）、滑鼠點擊映射到 `Session.Tap`。以建構標籤隔離，確保無頭環境的 `go test ./...` 與 `pants test ::` 不受影響。

## Capabilities

### New Capabilities
- `local-play`: 本機對局的點選互動控制器與棋盤渲染。

### Modified Capabilities
<!-- 無 -->

## Impact

- 程式碼：新增 `core/play`（`session.go`，純邏輯）；新增 `cmd/xiangqi`（Ebiten，建構標籤 `ebiten`）。
- 測試：`core/play` 單元測試（選子/高亮/走子/改選/取消/悔棋/認輸/結束/記譜輸出），table-driven + testify。
- 相依：`core/rules`、`core/board`、`core/record`；Ebiten 僅由標籤化的渲染層引入，不進入預設建構/測試。
- 建構：渲染層以 `go run -tags ebiten ./cmd/xiangqi` 或 Pants 對應目標執行（GUI，需圖形環境）。
