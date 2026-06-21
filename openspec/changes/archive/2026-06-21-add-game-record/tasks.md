## 1. 測試先行

- [x] 1.1 新增 `conformance/movelist_cases.json`（走法序列 → 中文清單）+ loader + 測試
- [x] 1.2 新增 `core/record` 單元測試（Recorder 漸進記錄、拒絕非法、Timeline 長度/索引）

## 2. 實作

- [x] 2.1 `Recorder`：`NewRecorder` / `Add`（合法性驗證）/ `SetResult` / `Current` / `Record`
- [x] 2.2 `Timeline`：`NewTimeline` / `Len` / `At`
- [x] 2.3 `MovesInChinese`：重放並逐步轉中文

## 3. 驗收

- [x] 3.1 `go test ./...` 全綠
- [x] 3.2 `pants test ::` 全綠
- [x] 3.3 `openspec validate add-game-record` 通過
