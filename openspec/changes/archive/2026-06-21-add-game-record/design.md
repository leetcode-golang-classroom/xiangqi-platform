## Context

`core/record` 已有 `Record`/`Marshal`/`Unmarshal`/`Replay`。本變更新增上層能力：漸進記錄、復盤導覽、中文記譜清單。建立於 `rule-engine`。

## Goals / Non-Goals

**Goals**：邊下邊記錄（含合法性驗證）、復盤逐手導覽、中文記譜清單。
**Non-Goals**：檔案持久化（屬 `client`/Storage）、UI 渲染、棋譜檔交換格式擴充。

## Decisions

- **`Recorder` 內含一個 `*rules.Game` 當前狀態**：`Add` 以 `ApplyMove` 驗證合法性後才記錄 UCCI，確保棋譜永遠合法、可重放。
- **`Timeline` 以 `Replay` 為基礎**：回傳 `[]*rules.Game`（含起始），索引即手序，前進/後退由呼叫端控制索引。
- **`MovesInChinese` 在重放過程逐步呼叫 `Game.ToChinese`**：中文需「走子前的盤面」上下文，故邊走邊轉。
- **不引入時鐘**：`Record.Date` 由呼叫端設定（核心不依賴系統時間，利於測試與移植）。

## Risks / Trade-offs

- [中文記譜的多子消歧（前/後）在連續棋局可能出現] → `ToChinese` 已支援同縱線兩子前/後；更複雜情形以增量 fixture 補強。
- [Timeline 一次重放全部] → 對局長度有限，記憶體可忽略；若需大量棋譜可改惰性迭代。

## Open Questions

- 無。
