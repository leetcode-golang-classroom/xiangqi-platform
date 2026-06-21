# storage Specification

## Purpose
TBD - created by archiving change add-storage. Update Purpose after archive.
## Requirements
### Requirement: 棋譜存檔與載入
系統 SHALL 提供 `Store`，能以識別字 `id` 將 `Record` 存檔並依同一 `id` 載回，載回的棋譜 MUST 與存入者一致（往返不失真）。

#### Scenario: 存讀往返
- **WHEN** 以 `id` 存入一份棋譜，再以同一 `id` 載入
- **THEN** 載回的 `Record` 與存入者相等

#### Scenario: 載入未知識別字
- **WHEN** 以不存在的 `id` 載入
- **THEN** 回報錯誤（不存在）

### Requirement: 棋局清單
系統 SHALL 提供 `List`，回傳已存棋譜的輕量 `Entry`（含 `id`、對局者、日期、結果），順序穩定（依 `id` 排序），且無需載入完整走法。

#### Scenario: 列出已存棋譜
- **WHEN** 存入多份棋譜後呼叫 `List`
- **THEN** 回傳每份棋譜的 `Entry`，依 `id` 升冪排序，且含對局者與結果等中介資料

### Requirement: 刪除棋譜
系統 SHALL 提供 `Delete`，移除指定 `id` 的棋譜；刪除後該 `id` MUST 不再出現於 `List`，且再次載入 MUST 回報不存在。

#### Scenario: 刪除後不再列出
- **WHEN** 刪除一份已存棋譜
- **THEN** 該 `id` 不再出現於 `List`，且以該 `id` 載入回報錯誤

### Requirement: 識別字安全性
系統 SHALL 拒絕不安全的識別字，凡含路徑分隔字元或 `..` 片段者 MUST 於存檔/載入/刪除時回報錯誤，避免目錄穿越。

#### Scenario: 拒絕目錄穿越
- **WHEN** 以含 `/` 或 `..` 的 `id` 存檔
- **THEN** 回報錯誤，且不寫入儲存目錄以外的位置

