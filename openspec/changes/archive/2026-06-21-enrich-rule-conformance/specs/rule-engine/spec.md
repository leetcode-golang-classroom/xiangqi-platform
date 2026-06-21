# rule-engine

為既有 requirement 補上邊界場景與金樣案例（行為不變）。

## MODIFIED Requirements

### Requirement: 相象、仕士、帥將、兵卒走法限制
- 相/象 SHALL 走「田」字、不可過河、塞象眼則非法。
- 仕/士 SHALL 走斜一格且不可出九宮。
- 帥/將 SHALL 走直一格且不可出九宮；兩方主將 SHALL NOT 在同一直線上直接照面（飛將）。
- 兵/卒 SHALL 未過河僅可前進一格，過河後可前進或左右一格，不可後退。

#### Scenario: 將帥不可照面
- **WHEN** 一步走法會使紅帥與黑將之間同一縱線無任何棋子
- **THEN** 該走法非法

#### Scenario: 飛將使帥的走法非法（對應 legalmoves_cases.json: flying-general-forbids-king-move）
- **WHEN** 紅帥位於 d0、黑將位於 e9，且帥移往 e0 會使兩將於 e 線照面
- **THEN** 紅帥唯一合法走法為 d1（往 e0 因飛將非法）

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

### Requirement: 走子後不得使己方被將軍
任何走法 SHALL 在套用後不使己方主將處於被將軍狀態（含暴露飛將），否則非法。

#### Scenario: 過濾自將走法
- **WHEN** 某候選走法套用後己方主將被將軍
- **THEN** 該走法不在合法走法集合中

#### Scenario: 受牽制的車不可橫移（對應 legalmoves_cases.json: pin-rook-no-horizontal）
- **WHEN** 紅車位於 a3，於 a 線上夾在紅帥 a0 與黑車 a9 之間（被牽制）
- **THEN** 合法走法僅沿 a 線（a1、a2、a4–a8、吃 a9），所有橫向走法因暴露飛將／被將而非法

### Requirement: 對局結果判定
系統 SHALL 判定對局結果：被將軍且無任何合法走法為**將死**（checkmate，對方勝）；未被將軍但無任何合法走法為**困斃**（stalemate，該方判負）；尚有合法走法則對局未結束。

#### Scenario: 進行中（對應 result_cases.json: startpos-ongoing）
- **WHEN** 盤面為開局且輪紅走
- **THEN** 結果為未結束（over=false）

#### Scenario: 將死（對應 result_cases.json: rook-checkmate）
- **WHEN** 盤面為 `R3k4/9/9/9/9/9/9/9/9/4K4`、輪黑走，黑將被 a9 車將軍且無合法應將
- **THEN** 結果為 over=true、winner=red、reason=checkmate

#### Scenario: 困斃判負（對應 result_cases.json: stalemate-loss）
- **WHEN** 黑方僅餘將於 d9、未被將軍，但 d8 與 e9 皆被紅子控制而無任何合法走法
- **THEN** 結果為 over=true、winner=red、reason=stalemate

#### Scenario: 被將軍但未將死＝進行中（對應 result_cases.json: black-in-check-ongoing）
- **WHEN** 黑將 e9 被 e 線紅車將軍，但仍可走 d9 或 f9 應將
- **THEN** 結果為未結束（over=false）

### Requirement: 長將判負與和棋
系統 SHALL 在一方以連續將軍循環逼和時判該方負（長將判負），並 SHALL 在自然限著（長時間無進展）達上限時判和。

#### Scenario: 長將判負
- **WHEN** 一方對同一局面以連續將軍重複達規定次數
- **THEN** 判長將方負（reason=perpetual_check）

#### Scenario: 重複局面（無連將）判和（對應 adjudicate_cases.json: repetition-draw-no-checks）
- **WHEN** 雙方各自往返移子使同一局面重複三次，且過程中無任一方連續將軍
- **THEN** 判和（over=true、winner 為空、reason=repetition_draw）

### Requirement: 中文記譜顯示
系統 SHALL 能將一步 UCCI 走法轉成中文縱線記譜（如「炮二平五」）供顯示，儲存仍以 UCCI 為準。

#### Scenario: 座標轉中文
- **WHEN** 由開局盤面轉換走法 `h2e2`
- **THEN** 顯示為「炮二平五」

#### Scenario: 同線同子前後辨異（對應 chinese_cases.json: red-front-rook-advance）
- **WHEN** 紅方兩車同在 a 線（a0、a3），將較前者（a3）進二至 a5
- **THEN** 顯示為「前車進二」（以前/後取代縱線數）
