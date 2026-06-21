## Why

`pseudoTargets` 目前以單一 `switch p.Kind()` 內聯 7 種棋子的走法邏輯，不易對個別棋子做隔離測試，未來要支援規則變體時也須改動同一個大函式。改成 Strategy（函式值註冊表）可讓各子邏輯獨立、可單獨測試、易擴充。

## What Changes

- 將 `core/rules/movegen.go` 的 `pseudoTargets` `switch` 重構為**函式值註冊表**：`map[byte]targetFn`，每種棋子一個產生函式。
- 抽出共用 helper（直線滑動、落點過濾），供各策略重用。
- **行為不變**：純內部重構，對外介面與規則結果完全相同。
- 先補齊 仕/相/帥/兵 的專屬 conformance 測試（TDD），確保重構前每子皆有獨立覆蓋。

## Capabilities

### New Capabilities
<!-- 無：純重構，不新增能力 -->

### Modified Capabilities
<!-- 無：規則行為不變，rule-engine 既有 requirements 不修改 -->

## Impact

- 程式碼：`core/rules/movegen.go`（重構）、`conformance/legalmoves_cases.json`（新增各子案例）。
- 風險：低——以既有 + 新增 conformance 測試全程把關，重構前後皆須全綠。
- 無資料格式、介面或規格變更。
