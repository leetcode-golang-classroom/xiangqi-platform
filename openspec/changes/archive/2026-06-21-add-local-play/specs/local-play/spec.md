# local-play

本機對局的點選互動控制器與棋盤渲染。建立於 `rule-engine` 與 `game-record` 之上。

## ADDED Requirements

### Requirement: 點選互動狀態機
系統 SHALL 提供 `Session`，以「點擊棋格」驅動對局：未選子時點擊輪走方可動的棋子 SHALL 選取該子並提供其所有合法落點；已選子時點擊合法落點 SHALL 執行該步並清除選取；點擊另一己方可動子 SHALL 改選該子；點擊其餘空格或不可選處 SHALL 取消選取。點擊不改變盤面以外狀態時 MUST NOT 產生非法走法。

#### Scenario: 選子並取得合法落點
- **WHEN** 開局輪紅走，點擊紅炮 h2
- **THEN** h2 被選取，且合法落點集合等同該炮在當前盤面的合法走法目標

#### Scenario: 點擊合法落點即走子並記譜
- **WHEN** 已選取 h2 後點擊合法落點 e2
- **THEN** 套用走法 `h2e2`、輪到黑方、棋譜新增 `h2e2`，且選取被清除

#### Scenario: 改選另一己方棋子
- **WHEN** 已選取某子後，點擊另一個輪走方可動的棋子
- **THEN** 改為選取後者並提供其合法落點

#### Scenario: 點擊不可選處取消選取
- **WHEN** 已選取某子後，點擊空格或非合法落點且非己方可動子之處
- **THEN** 清除選取，盤面不變

### Requirement: 悔棋與認輸
系統 SHALL 支援悔棋（回退最後一手，回到前一盤面與輪走方）與認輸（由當前輪走方認輸，判對方勝）。悔棋於無走法時 SHALL 為無操作。

#### Scenario: 悔棋回退一手
- **WHEN** 已走一手後執行悔棋
- **THEN** 盤面與輪走方回到該手之前，且棋譜移除該手

#### Scenario: 認輸判對方勝
- **WHEN** 輪紅走時紅方認輸
- **THEN** 對局結束，結果為黑方勝（reason=resign）

### Requirement: 對局結束判定與棋譜輸出
系統 SHALL 在每步後依 `rule-engine` 判定對局是否結束（將死/困斃/和棋），結束後 SHALL 拒絕進一步走子；並 SHALL 能隨時輸出 `xiangqi-record-v1` 棋譜（走法為 UCCI），供 `storage` 持久化。

#### Scenario: 結束後拒絕走子
- **WHEN** 對局已結束（將死或認輸）後再點擊任何棋格
- **THEN** 不產生走法，盤面與結果不變

#### Scenario: 輸出棋譜
- **WHEN** 對局進行數手後輸出棋譜
- **THEN** 得到 `xiangqi-record-v1` 容器，其 `moves` 為已走的 UCCI 序列、含對局者與結果
