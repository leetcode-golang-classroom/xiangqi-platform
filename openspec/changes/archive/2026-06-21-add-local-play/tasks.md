## 1. 測試先行（純邏輯控制器）

- [x] 1.1 `core/play` 單元測試：選子取得合法落點、點落點走子並記譜、改選、取消選取
- [x] 1.2 悔棋回退、認輸判對方勝、結束後拒絕走子、棋譜輸出（table-driven + testify）

## 2. 實作

- [x] 2.1 `Session`：`NewSession` / `Tap`（狀態機）/ `Selected` / `Targets` / `Current` / `Turn` / `Outcome`
- [x] 2.2 `Undo` / `Resign` / `Record`（輸出 xiangqi-record-v1）
- [x] 2.3 `core/play/BUILD`（pants tailor）

## 3. Ebiten 渲染層（建構標籤隔離）

- [x] 3.1 `cmd/xiangqi`（`//go:build ebiten`）：棋盤/棋子繪製、滑鼠點擊→`Session.Tap`
- [x] 3.2 確認預設建構不含 Ebiten（無頭 CI 不受影響）

## 4. 驗收

- [x] 4.1 `go test ./...` 全綠（不含 ebiten 標籤）
- [x] 4.2 `pants test ::` 全綠
- [x] 4.3 `go build -tags ebiten ./cmd/xiangqi` 編譯通過
- [x] 4.4 `openspec validate add-local-play` 通過
