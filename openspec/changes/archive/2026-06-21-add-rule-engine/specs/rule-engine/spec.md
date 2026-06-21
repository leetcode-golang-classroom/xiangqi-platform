# rule-engine

象棋規則引擎的權威行為規格。每條 Requirement 對應 `conformance/*.json` 中的黃金案例，可直接轉成測試。

## ADDED Requirements

### Requirement: 座標與走法表示
系統 SHALL 以 file `a`–`i`、rank `0`–`9` 表示座標（紅方視角左→右、紅底→黑頂），走法以 UCCI `<from><to>` 表示（如 `h2e2`）。

#### Scenario: 解析 UCCI 走法
- **WHEN** 解析字串 `h2e2`
- **THEN** 得到 from=h2、to=e2 的走法

### Requirement: FEN 與盤面互轉
系統 SHALL 能由 Xiangqi FEN 還原盤面與輪走方，並能將盤面輸出為相同的 FEN（round-trip 一致）。大寫=紅、小寫=黑，行序為 rank9→rank0。

#### Scenario: FEN round-trip（對應 fen_cases.json）
- **WHEN** 以開局 FEN `rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1` 載入後再輸出
- **THEN** 輸出字串與輸入完全相同

### Requirement: 車走法
車 SHALL 沿同一直線（橫或縱）移動任意格，路徑中不得有其他棋子，終點為空或敵子。

#### Scenario: 空盤車的走法（對應 legalmoves_cases.json: rook-on-open-board）
- **WHEN** 紅車位於 c4、盤面其餘相關直線為空
- **THEN** 合法走法為該縱線與該橫線上所有可達格（17 步）

### Requirement: 馬走法與蹩馬腿
馬 SHALL 走「日」字（先一直再一斜）；若移動方向的鄰格（馬腿）有子，該方向走法非法（蹩馬腿）。

#### Scenario: 無阻擋的馬（對應 legalmoves_cases.json: horse-no-block）
- **WHEN** 紅馬位於 e4 且四周馬腿皆空
- **THEN** 八個方向皆為合法走法

#### Scenario: 馬腿被蹩（對應 legalmoves_cases.json: horse-leg-blocked）
- **WHEN** 紅馬位於 e4 且 e5 有子
- **THEN** 需經過 e5 的兩個方向（d6、f6）非法，其餘六方合法

### Requirement: 炮走法與吃子
炮 SHALL 不吃子時如車直走（路徑須全空）；吃子時 SHALL 在炮與目標之間恰有一個棋子（炮架），且目標為敵子。

#### Scenario: 炮的走與吃（對應 legalmoves_cases.json: cannon-move-and-capture）
- **WHEN** 紅炮位於 e4、e6 有一棋子、e8 為敵子
- **THEN** 沿空線各格為合法非吃走法，且可隔 e6 吃 e8

### Requirement: 相象、仕士、帥將、兵卒走法限制
- 相/象 SHALL 走「田」字、不可過河、塞象眼則非法。
- 仕/士 SHALL 走斜一格且不可出九宮。
- 帥/將 SHALL 走直一格且不可出九宮；兩方主將 SHALL NOT 在同一直線上直接照面（飛將）。
- 兵/卒 SHALL 未過河僅可前進一格，過河後可前進或左右一格，不可後退。

#### Scenario: 將帥不可照面
- **WHEN** 一步走法會使紅帥與黑將之間同一縱線無任何棋子
- **THEN** 該走法非法

### Requirement: 走子後不得使己方被將軍
任何走法 SHALL 在套用後不使己方主將處於被將軍狀態（含暴露飛將），否則非法。

#### Scenario: 過濾自將走法
- **WHEN** 某候選走法套用後己方主將被將軍
- **THEN** 該走法不在合法走法集合中

### Requirement: 對局結果判定
系統 SHALL 判定對局結果：被將軍且無任何合法走法為**將死**（checkmate，對方勝）；未被將軍但無任何合法走法為**困斃**（stalemate，該方判負）；尚有合法走法則對局未結束。

#### Scenario: 進行中（對應 result_cases.json: startpos-ongoing）
- **WHEN** 盤面為開局且輪紅走
- **THEN** 結果為未結束（over=false）

#### Scenario: 將死（對應 result_cases.json: rook-checkmate）
- **WHEN** 盤面為 `R3k4/9/9/9/9/9/9/9/9/4K4`、輪黑走，黑將被 a9 車將軍且無合法應將
- **THEN** 結果為 over=true、winner=red、reason=checkmate

### Requirement: 長將判負與和棋
系統 SHALL 在一方以連續將軍循環逼和時判該方負（長將判負），並 SHALL 在自然限著（長時間無進展）達上限時判和。

#### Scenario: 長將判負
- **WHEN** 一方對同一局面以連續將軍重複達規定次數
- **THEN** 判長將方負（reason=perpetual_check）

### Requirement: 棋譜記錄與復盤
系統 SHALL 以 `xiangqi-record-v1` JSON 記錄一局（走法存 UCCI），並 SHALL 能由起始 FEN 依序套用走法重現每一手後的盤面。

#### Scenario: 重放棋譜（對應 record_cases.json: open-with-central-cannon）
- **WHEN** 由開局 FEN 套用走法序列 `["h2e2"]`
- **THEN** 最終盤面等於 `rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C2C4/9/RNBAKABNR b - - 1 1`

### Requirement: 中文記譜顯示
系統 SHALL 能將一步 UCCI 走法轉成中文縱線記譜（如「炮二平五」）供顯示，儲存仍以 UCCI 為準。

#### Scenario: 座標轉中文
- **WHEN** 由開局盤面轉換走法 `h2e2`
- **THEN** 顯示為「炮二平五」
