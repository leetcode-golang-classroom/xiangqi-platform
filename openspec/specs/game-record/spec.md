# game-record Specification

## Purpose
TBD - created by archiving change add-game-record. Update Purpose after archive.
## Requirements
### Requirement: 對局中漸進記錄
系統 SHALL 提供 `Recorder`，能開新局（對局者、起始 FEN）、逐手附加走法、標記結果，並輸出 `xiangqi-record-v1` 棋譜。附加的走法 MUST 為當前盤面的合法走法，否則回報錯誤且不記錄。

#### Scenario: 記錄合法走法
- **WHEN** 由開局建立 Recorder 並附加合法走法 `h2e2`
- **THEN** 輸出棋譜的 `moves` 包含 `h2e2`，且當前盤面為該步之後

#### Scenario: 拒絕非法走法
- **WHEN** 對開局 Recorder 附加非法走法（如 `e0e5`）
- **THEN** 回報錯誤，且棋譜不記錄該步

### Requirement: 復盤導覽
系統 SHALL 提供 `Timeline`，由棋譜重現每一手後的盤面（含起始局面），長度為走法數加一，並可依手序索引任一盤面。

#### Scenario: 逐手導覽
- **WHEN** 由含 N 步的棋譜建立 Timeline
- **THEN** 其長度為 N+1，索引 0 為起始局面、索引 N 為終局盤面

### Requirement: 中文記譜清單
系統 SHALL 能將棋譜的 UCCI 走法序列轉成中文縱線記譜清單供顯示，順序與走法一致。

#### Scenario: 走法序列轉中文（對應 movelist_cases.json: open-cannon-then-horse）
- **WHEN** 由開局重放走法 `["h2e2", "h9g7"]`
- **THEN** 得到中文清單 `["炮二平五", "馬8進7"]`

### Requirement: 復盤游標
系統 SHALL 提供 `Replayer`，以游標包裝棋譜的 `Timeline`，支援逐手前進/後退與跳轉，並回傳當前盤面與位置。前進/後退/跳轉 MUST 夾制於 `[0, Len-1]` 範圍內（不越界）。

#### Scenario: 邊界夾制的前進後退
- **WHEN** 由含 N 步的棋譜建立 `Replayer`（長度 N+1，游標起始為 0）
- **THEN** 連續 `Next` 至末位後再 `Next` 仍停在末位（回報未移動）；連續 `Prev` 至 0 後再 `Prev` 仍停在 0

#### Scenario: 游標對應盤面
- **WHEN** 將游標跳轉至索引 i
- **THEN** `Current` 回傳第 i 手之後的盤面（與 `Timeline.At(i)` 一致），索引 0 為起始局面

