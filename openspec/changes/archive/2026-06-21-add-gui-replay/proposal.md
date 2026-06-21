## Why

招牌功能「棋譜記錄與復盤」目前只到程式庫層：`Record`/`Timeline`/`Storage` 已可存讀，但 GUI 既不能載入復盤、對局結束也只在狀態列小字顯示，且棋譜僅在手動按 `S` 時才產出——玩家容易以為「沒有產出棋譜」。本變更讓復盤體驗在 GUI 可用，並確保每局結束必有棋譜。

## What Changes

- 新增可測的 `Replayer`（`core/record`）：以游標包裝 `Timeline`，提供 `Next`/`Prev`/`Seek`（邊界夾制）/`Current`/`Index`/`Len`，作為 GUI 與日後其他前端共用的復盤導覽核心。
- GUI（`cmd/xiangqi`）：
  - **載入復盤**：鍵 `L` 進入復盤模式，由 `records/` 載入棋譜並以 `←`/`→` 逐手前進/後退、`L` 切換下一份、`Esc` 返回對局。
  - **對局結束自動存譜**：偵測對局結束時自動存成 `xiangqi-record-v1`（仍保留手動 `S`），確保棋譜必有產出並顯示存檔路徑。
  - **結束橫幅**：對局結束時於畫面中央顯示「棋局結束」與勝方（紅勝/黑勝/和局＋原因），不再只有狀態列小字。

## Capabilities

### New Capabilities
<!-- 無（Replayer 歸入既有 game-record）-->

### Modified Capabilities
- `game-record`: 新增 `Replayer` 復盤游標（建立於 `Timeline` 之上）。

## Impact

- 程式碼：`core/record` 新增 `Replayer`；`cmd/xiangqi` 新增復盤模式、自動存譜、結束橫幅（建構標籤 `ebiten`）。
- 測試：`core/record` 新增 `Replayer` 單元測試（長度、邊界夾制、游標盤面）。
- GUI 行為（渲染/輸入）不納入語言中立測試，沿用既有慣例。
