# ai-opponent

對手抽象介面與純 Go AI 引擎（搜尋、評估、難度）。建立於 `rule-engine` 之上。

## ADDED Requirements

### Requirement: 對手抽象介面
系統 SHALL 提供 `Player` 介面，以 `SelectMove(game)` 由當前盤面取得一步走法，統一本地人類、AI 與遠端對手。對局迴圈 MUST 僅依賴此介面而不在意對手種類。AI 等實作於對局已結束時 SHALL 回報錯誤而非走子。

#### Scenario: 由介面取得走法
- **WHEN** 對未結束盤面呼叫某 `Player` 的 `SelectMove`
- **THEN** 回傳一步當前盤面的合法走法

#### Scenario: 對局結束回報錯誤
- **WHEN** 對已結束（將死/和棋等）盤面呼叫 `SelectMove`
- **THEN** 回報錯誤且不回傳走法

### Requirement: AI 搜尋與評估
系統 SHALL 提供實作 `Player` 的 `AI`，以 negamax + alpha-beta 剪枝搜尋至指定深度選步。評估函數 SHALL 以子力價值為主；終局將死/困斃 SHALL 視為對行棋方的必負（必勝），和棋為 0。搜尋 SHALL 具決定性：相同盤面與深度回傳固定走法。

#### Scenario: 吃掉無保護的子
- **WHEN** 紅方可一步吃掉無保護的黑車且不被反吃
- **THEN** AI 選擇該吃子走法

#### Scenario: 找到一步將死
- **WHEN** 存在一步走法可將死對方
- **THEN** AI 選擇之，套用後盤面為將死（對方負）

#### Scenario: 回傳合法走法
- **WHEN** 對開局盤面呼叫 AI 選步
- **THEN** 回傳的走法屬於當前合法走法集合

### Requirement: 難度分級
系統 SHALL 以搜尋深度提供難度分級（如 Easy/Medium/Hard），深度越高棋力越強；建立 AI 時 SHALL 可指定難度。

#### Scenario: 難度對應搜尋深度
- **WHEN** 以較高難度建立 AI
- **THEN** 其搜尋深度不小於較低難度者
