## Why

招牌功能「棋譜記錄與復盤」要可用，棋譜必須能**留存於本機**並在棋局清單中重新開啟。目前 `core/record` 只能在記憶體中建立與序列化 `Record`，沒有任何持久化能力；單機 App 無法存下對局、列出歷史、刪除或載入復盤。

## What Changes

- 新增 `storage` 能力：以 `Store` 介面定義本機棋譜持久化契約——`Save` / `Load` / `List` / `Delete`。
- 新增 `FileStore`：以單一目錄為後端的檔案系統實作，每局棋譜存為一個 `xiangqi-record-v1` JSON 檔。
- `List` 回傳輕量 `Entry`（對局者、日期、結果），讓棋局清單免於載入完整走法。
- 強化 ID 安全性：拒絕含路徑分隔或 `..` 的識別字，避免目錄穿越。
- 沿用既有資料格式（`xiangqi-record-v1`），不變更棋譜內容契約。

## Capabilities

### New Capabilities
- `storage`: 本機棋譜的存檔、載入、列表與刪除。

### Modified Capabilities
<!-- 無 -->

## Impact

- 程式碼：新增 `core/storage`（`store.go`），依賴 `core/record`；`core/record` 維持純邏輯（不引入 os）。
- 測試：`core/storage` 單元測試（存讀往返、列表排序與中介資料、刪除、未知 ID、ID 安全性），以 `t.TempDir()` 隔離檔案系統。
- 可移植性：`Store` 為語言中立介面契約；`FileStore` 為桌面/單機實作，行動端可另實作平台儲存。跨語言一致性契約仍為既有 `xiangqi-record-v1` JSON（已在 `conformance/` 涵蓋），故本變更不新增語言中立 fixture。
