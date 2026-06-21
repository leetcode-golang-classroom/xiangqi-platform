## 1. 測試先行

- [x] 1.1 新增 `core/storage` 單元測試（存讀往返、未知 ID、List 排序與中介資料、刪除、ID 安全性），以 `t.TempDir()` 隔離

## 2. 實作

- [x] 2.1 `Store` 介面 + `Entry` 中介資料型別
- [x] 2.2 `FileStore`：`NewFileStore` / `Save`（含 ID 安全性檢查）/ `Load` / `List`（排序）/ `Delete`
- [x] 2.3 `core/storage/BUILD`（pants tailor）

## 3. 驗收

- [x] 3.1 `go test ./...` 全綠
- [x] 3.2 `pants test ::` 全綠
- [x] 3.3 `openspec validate add-storage` 通過
