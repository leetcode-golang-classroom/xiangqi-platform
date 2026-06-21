# match

`Controller` 改為統一迴圈：對手不分人類/AI 一致處理。

## MODIFIED Requirements

### Requirement: 對手指派與回合分派
系統 SHALL 提供 `Controller`，持有一個 `Session` 與各方（紅/黑）`Player` 指派。`Controller` SHALL 以統一迴圈推進：每次 `Step` 向當前輪走方的 `Player` 請求一步（`RequestMove`），待其送出後套用並換手；對局結束時不再請求。迴圈 MUST 不區分對手種類（人類/AI/遠端一致）。

#### Scenario: 統一迴圈推進對局
- **WHEN** 紅、黑皆指派 AI，反覆 `Step`
- **THEN** 每步皆為當前方的合法走法，輪走方依序換手，棋譜逐步累積

#### Scenario: 對局結束停止請求
- **WHEN** 對局已結束後再 `Step`
- **THEN** 不向任何 `Player` 請求、不走子，盤面與結果不變

### Requirement: 人類回合只接受點擊
系統 SHALL 在當前輪走方為 `Interactive`（人類）時，將點擊轉交該玩家處理（沿用點選狀態機），其完成的走法經由統一迴圈套用；當前方非互動式（AI/遠端）時點擊 SHALL 被忽略。

#### Scenario: 人類回合點擊產出走法
- **WHEN** 當前為人類回合，依序點擊己方棋子與其合法落點
- **THEN** 該步經由迴圈套用、換手、棋譜新增該步

#### Scenario: 非人類回合忽略點擊
- **WHEN** 當前為 AI 回合時點擊任意棋格
- **THEN** 不選子、不走子，盤面與選取狀態不變

### Requirement: AI 回合自動選步
系統 SHALL 在當前輪走方為 AI 時，由其 `RequestMove` 背景取步，完成後經由統一迴圈套用，無需人為操作。

#### Scenario: AI 回合自動走一步
- **WHEN** 當前為 AI 回合並持續 `Step`
- **THEN** AI 背景選出一步合法走法並套用，輪走方換手、棋譜新增該步
