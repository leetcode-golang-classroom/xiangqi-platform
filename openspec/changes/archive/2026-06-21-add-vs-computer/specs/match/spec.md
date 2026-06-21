# match

對局協調——依輪走方分派人類或 AI 走子。建立於 `local-play`（Session）與 `ai-opponent`（Opponent）之上。

## ADDED Requirements

### Requirement: 對手指派與回合分派
系統 SHALL 提供 `Controller`，持有一個 `Session` 與各方（紅/黑）對手指派；未指派者視為人類。`Controller` SHALL 依當前輪走方判定該回合由人類或 AI 走子，並於對局結束時兩者皆不走子。

#### Scenario: 判定當前回合的走子者
- **WHEN** 紅方未指派對手、黑方指派 AI，且輪到黑方
- **THEN** 當前為 AI 回合（非人類回合）

#### Scenario: 對局結束不再走子
- **WHEN** 對局已結束
- **THEN** 既非人類回合亦非 AI 回合

### Requirement: 人類回合只接受點擊
系統 SHALL 在人類回合接受點擊互動（沿用 `Session` 點選狀態機），並 SHALL 在 AI 回合忽略人類點擊。

#### Scenario: AI 回合忽略人類點擊
- **WHEN** 當前為 AI 回合時對任意棋格點擊
- **THEN** 不選子、不走子，盤面與選取狀態不變

### Requirement: AI 回合自動選步
系統 SHALL 在 AI 回合由該方對手引擎選出一步並套用，推進對局並記譜；非 AI 回合時 SHALL 為無操作。

#### Scenario: AI 回合走一步
- **WHEN** 當前為 AI 回合並要求 AI 走子
- **THEN** 套用一步該方合法走法，輪走方換手，棋譜新增該步

#### Scenario: 人類回合要求 AI 走子為無操作
- **WHEN** 當前為人類回合卻要求 AI 走子
- **THEN** 不走子，回報未走子
