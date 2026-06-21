## 1. 補金樣案例（沿用既有 loader / table-driven 測試）

- [x] 1.1 `result_cases.json`：新增 `stalemate-loss`、`black-in-check-ongoing`
- [x] 1.2 `legalmoves_cases.json`：新增 `flying-general-forbids-king-move`、`pin-rook-no-horizontal`
- [x] 1.3 `adjudicate_cases.json`：新增 `repetition-draw-no-checks`
- [x] 1.4 `chinese_cases.json`：新增 `red-front-rook-advance`

## 2. 驗證

- [x] 2.1 `go test ./...` 全綠（若揭露缺陷則就地修正）
- [x] 2.2 `pants test ::` 全綠
- [x] 2.3 `openspec validate enrich-rule-conformance` 通過
