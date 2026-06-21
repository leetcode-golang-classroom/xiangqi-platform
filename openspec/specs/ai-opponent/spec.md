# ai-opponent Specification

## Purpose
TBD - created by archiving change add-ai-opponent. Update Purpose after archive.
## Requirements
### Requirement: 對手抽象介面
系統 SHALL 提供 `Player` 介面，以 `RequestMove(game)` **非同步**取步：回傳一個通道，當該對手決定一步時送出之，統一本地人類、AI 與遠端對手。對局迴圈 MUST 僅依賴此介面而不在意對手種類；`RequestMove` 僅於對局未結束且輪到該方時呼叫。AI 之同步核心 `SelectMove` 於對局已結束時 SHALL 回報錯誤。

#### Scenario: 由介面非同步取得走法
- **WHEN** 對未結束盤面呼叫某 `Player` 的 `RequestMove`
- **THEN** 其通道最終送出一步當前盤面的合法走法

#### Scenario: AI 同步核心於對局結束回報錯誤
- **WHEN** 對已結束（將死/和棋等）盤面呼叫 AI 的 `SelectMove`
- **THEN** 回報錯誤且不回傳走法

### Requirement: AI 搜尋與評估
系統 SHALL 提供實作 `Player` 的 `AI`，以 negamax + alpha-beta 剪枝搜尋至指定深度選步。評估函數 SHALL 以子力價值為主；終局將死/困斃 SHALL 視為對行棋方的必負（必勝），和棋為 0。選步 SHALL 可重現：**全新** AI 於同一盤面與深度的首次選步回傳固定走法（同一 AI 於對局中重訪同一盤面時可變招，見「重複局面變招」）。

#### Scenario: 吃掉無保護的子
- **WHEN** 紅方可一步吃掉無保護的黑車且不被反吃
- **THEN** AI 選擇該吃子走法

#### Scenario: 找到一步將死
- **WHEN** 存在一步走法可將死對方
- **THEN** AI 選擇之，套用後盤面為將死（對方負）

#### Scenario: 回傳合法走法
- **WHEN** 對開局盤面呼叫 AI 選步
- **THEN** 回傳的走法屬於當前合法走法集合

#### Scenario: 全新 AI 首手可重現
- **WHEN** 兩個全新 AI 以相同難度對同一盤面各選一次步
- **THEN** 兩者回傳相同走法

### Requirement: 難度分級
系統 SHALL 以搜尋深度提供難度分級（如 Easy/Medium/Hard），深度越高棋力越強；建立 AI 時 SHALL 可指定難度。

#### Scenario: 難度對應搜尋深度
- **WHEN** 以較高難度建立 AI
- **THEN** 其搜尋深度不小於較低難度者

### Requirement: 重複局面變招
為破解迴圈並增加回應多樣性，AI SHALL 於根節點蒐集所有**等值最佳手**，並於同一對局中**重訪同一盤面**時，在等值最佳手間依造訪次數輪替選擇。當最佳手唯一（如吃子、將死）時 SHALL 仍選該手（不犧牲評估）。

#### Scenario: 重複局面給出不同應對
- **WHEN** 同一 AI 對同一盤面連續選步兩次，且該盤面存在多個等值最佳手
- **THEN** 第二次回傳與第一次不同的（等值）走法

#### Scenario: 唯一最佳手不受影響
- **WHEN** 盤面存在唯一最佳手（如可白吃一子）
- **THEN** 無論造訪幾次，AI 皆選該唯一最佳手

