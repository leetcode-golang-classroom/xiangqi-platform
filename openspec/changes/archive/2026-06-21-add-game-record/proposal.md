## Why

棋譜記錄與復盤是平台的招牌功能。`core/record` 目前已有 `Record` 容器、`Marshal`/`Unmarshal` 與 `Replay`，但缺少「對局進行中漸進記錄」與「復盤導覽、中文記譜清單」等上層能力，無法直接支援單機對弈邊下邊存、以及復盤逐手檢視。

## What Changes

- 新增 `Recorder`：對局中漸進記錄——開新局（對局者、起始 FEN）、逐手附加**合法**走法（以規則引擎驗證）、標記結果、輸出 `Record`。
- 新增 `Timeline`：由 `Record` 取得每一手後的盤面（含起始），支援復盤前進/後退導覽。
- 新增 `MovesInChinese`：將棋譜的 UCCI 走法序列轉成中文記譜清單供顯示。
- 行為建立於既有 `rule-engine`（走法合法性、`ToChinese`、`Replay`）之上。

## Capabilities

### New Capabilities
- `game-record`: 棋譜的漸進記錄、復盤導覽與中文記譜清單。

### Modified Capabilities
<!-- 無 -->

## Impact

- 程式碼：`core/record`（新增 `recorder.go`）。
- 測試：`conformance/movelist_cases.json`（中文記譜清單，跨語言）+ `core/record` 單元測試（Recorder/Timeline，Go API）。
- 依賴：`core/rules`、`core/board`；無資料格式變更（沿用 `xiangqi-record-v1`）。
