## Why

中國象棋平台的所有功能（單機對弈、棋譜復盤、AI 對手、未來線上對戰）都建立在「正確的規則」之上。規則引擎是最確定、純邏輯、無 UI/網路依賴的核心，必須先做穩，後續功能才有可信地基。

## What Changes

- 新增**象棋規則引擎**核心套件（純 Go，零 UI/網路依賴）：
  - 盤面表示與座標系（已有 `core/board`）。
  - **走法產生**：車、馬、炮、相/象、仕/士、帥/將、兵/卒，含蹩馬腿、塞象眼、炮架吃子、過河兵變、仕相不出宮/不過河、將帥不照面。
  - **合法性過濾**：走子後不得使己方被將軍（含飛將）。
  - **勝負判定**：將軍、應將、將死、困斃（無合法走法判負）、長將判負、自然限著和棋。
  - **FEN ↔ 盤面** 與 **UCCI ↔ Move** 轉換、中文記譜顯示。
- 走法與狀態採**不可變**語意（`ApplyMove` 回傳新狀態），利於復盤與 AI 搜尋。
- 以語言中立的 `conformance/*.json` 黃金案例作為跨語言一致性契約。

## Capabilities

### New Capabilities
- `rule-engine`: 象棋規則的權威來源——走法產生、合法性、勝負判定，以及 FEN/UCCI/中文記譜轉換。

### Modified Capabilities
<!-- 無既有 spec 需修改 -->

## Impact

- 程式碼：`core/board`、`core/rules`、`core/notation`、`core/record`、`conformance/`。
- 介面契約：`RuleEngine` 語言中立簽章（見 `docs/DESIGN.md`），供 UI/AI/伺服器與未來他語言實作共用。
- 測試：`conformance/*.json` + table-driven harness（`go test` 與 `pants test` 皆可驅動）。
