## Why

規則引擎的 11 條 requirement 已實作且通過 `conformance/`，但部分已**明文規定**的邊界行為缺少語言中立金樣案例：困斃判負（stalemate）、被將軍但未將死（進行中）、飛將使走法非法、受牽制子的自將過濾、重複局面和棋（repetition_draw）、以及同線同子的前/後記譜辨異。補上這些案例可加厚跨語言一致性保證，並讓「requirement ↔ scenario ↔ fixture ↔ 測試」可追溯。

## What Changes

- 為既有 requirement 補上對應 fixture 的新場景（不改變行為，純驗證強化）：
  - `對局結果判定`：新增 **困斃判負** 與 **被將軍但未將死＝進行中** 兩場景。
  - `相象、仕士、帥將、兵卒走法限制`：新增 **飛將使將帥走法非法** 場景。
  - `走子後不得使己方被將軍`：新增 **受牽制子橫移自將則非法** 場景。
  - `長將判負與和棋`：新增 **重複局面（無連將）判和** 場景。
  - `中文記譜顯示`：新增 **同線同子前後辨異** 場景。
- 對應新增 `conformance/*.json` 金樣案例，沿用既有 loader 與 table-driven 測試（無需新增 harness）。

## Capabilities

### New Capabilities
<!-- 無 -->

### Modified Capabilities
- `rule-engine`: 為既有 requirement 補上邊界場景與金樣案例（行為不變）。

## Impact

- 測試：`conformance/{result,legalmoves,adjudicate,chinese}_cases.json` 新增案例；既有測試函式自動涵蓋。
- 程式碼：預期無變更；若新案例揭露缺陷則就地修正（TDD）。
- 規格：`rule-engine` 五條 requirement 以 MODIFIED 補場景。
