# rule-engine

本變更為內部重構（走法產生改 Strategy），行為不變。同時補齊 相/仕/帥/兵 的可驗證場景。

## MODIFIED Requirements

### Requirement: 相象、仕士、帥將、兵卒走法限制
- 相/象 SHALL 走「田」字、不可過河、塞象眼則非法。
- 仕/士 SHALL 走斜一格且不可出九宮。
- 帥/將 SHALL 走直一格且不可出九宮；兩方主將 SHALL NOT 在同一直線上直接照面（飛將）。
- 兵/卒 SHALL 未過河僅可前進一格，過河後可前進或左右一格，不可後退。

#### Scenario: 將帥不可照面
- **WHEN** 一步走法會使紅帥與黑將之間同一縱線無任何棋子
- **THEN** 該走法非法

#### Scenario: 仕在九宮內斜走（對應 legalmoves_cases.json: advisor-in-palace）
- **WHEN** 紅仕位於 e1
- **THEN** 合法走法為四個對角 d0、f0、d2、f2（皆在九宮內）

#### Scenario: 相走田字、不過河（對應 legalmoves_cases.json: elephant-no-block）
- **WHEN** 紅相位於 e2 且象眼皆空
- **THEN** 合法走法為 c0、g0、c4、g4（不越過河界）

#### Scenario: 塞象眼則該方向非法（對應 legalmoves_cases.json: elephant-eye-blocked）
- **WHEN** 紅相位於 e2 且 d3 有子（c4 方向的象眼）
- **THEN** 往 c4 的走法非法，其餘三方合法

#### Scenario: 帥在九宮內直走（對應 legalmoves_cases.json: king-in-palace）
- **WHEN** 紅帥位於 e1
- **THEN** 合法走法為 d1、f1、e0、e2（皆在九宮內）

#### Scenario: 兵未過河僅前進（對應 legalmoves_cases.json: pawn-before-river）
- **WHEN** 紅兵位於 e3（未過河）
- **THEN** 唯一合法走法為前進至 e4

#### Scenario: 兵過河後可左右（對應 legalmoves_cases.json: pawn-after-river）
- **WHEN** 紅兵位於 e5（已過河）
- **THEN** 合法走法為 e6、d5、f5
