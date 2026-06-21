## Context

`pseudoTargets`（`core/rules/movegen.go`）以 `switch p.Kind()` 內聯 7 子走法。需提升各子測試隔離與變體擴充性，且不改變行為。

## Goals / Non-Goals

**Goals**：各子走法邏輯獨立、可單獨測試、易擴充；保持行為不變、全程綠燈。
**Non-Goals**：不改規則、不改對外介面、不做效能優化（bitboard 等）。

## Decisions

- **採函式值註冊表（風格 A），非介面（風格 B）**：
  ```go
  type targetFn func(b *board.Board, from board.Square, c board.Color) []board.Square
  var pieceTargets = map[byte]targetFn{ 'r': rookTargets, ... }
  ```
  理由：比介面輕、貼合 Go 慣例；策略目前無需攜帶狀態。介面（B）留待未來策略需設定（如變體規則參數）時再升級。
- **抽出共用 helper**：`slideTargets`（車/炮共用的直線掃描）、落點允許判斷，避免重複。
- **逐子抽出、全程綠燈**：屬安全重構。

## 取捨（與維持 switch 比較）

| 面向 | Strategy（函式註冊表） | switch |
|---|---|---|
| 測試隔離 | ✅ 各子獨立函式 | 整段測 |
| 擴充（變體棋子） | ✅ 加一筆註冊 | 改 switch |
| 可讀性 | 分散多函式 | ✅ 一處覽全 |
| 效能 | map 查找（AI 熱路徑，微小） | ✅ 直接跳轉 |

## Risks / Trade-offs

- [過度設計：象棋棋子固定 7 種] → 以「測試隔離 + 未來變體」為採用理由；若僅求可讀，switch 亦足夠。本次依需求採 Strategy。
- [行為回歸] → 既有 + 新增 conformance 全程把關。

## Open Questions

- 無。
