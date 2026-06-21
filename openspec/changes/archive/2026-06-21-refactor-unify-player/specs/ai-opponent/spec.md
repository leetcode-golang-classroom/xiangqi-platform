# ai-opponent

`Player` 介面改為非同步取步；AI 於背景搜尋後送出走法。

## MODIFIED Requirements

### Requirement: 對手抽象介面
系統 SHALL 提供 `Player` 介面，以 `RequestMove(game)` **非同步**取步：回傳一個通道，當該對手決定一步時送出之，統一本地人類、AI 與遠端對手。對局迴圈 MUST 僅依賴此介面而不在意對手種類；`RequestMove` 僅於對局未結束且輪到該方時呼叫。AI 之同步核心 `SelectMove` 於對局已結束時 SHALL 回報錯誤。

#### Scenario: 由介面非同步取得走法
- **WHEN** 對未結束盤面呼叫某 `Player` 的 `RequestMove`
- **THEN** 其通道最終送出一步當前盤面的合法走法

#### Scenario: AI 同步核心於對局結束回報錯誤
- **WHEN** 對已結束（將死/和棋等）盤面呼叫 AI 的 `SelectMove`
- **THEN** 回報錯誤且不回傳走法
