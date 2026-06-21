# game-record

新增 `Replayer` 復盤游標，建立於 `Timeline` 之上。

## ADDED Requirements

### Requirement: 復盤游標
系統 SHALL 提供 `Replayer`，以游標包裝棋譜的 `Timeline`，支援逐手前進/後退與跳轉，並回傳當前盤面與位置。前進/後退/跳轉 MUST 夾制於 `[0, Len-1]` 範圍內（不越界）。

#### Scenario: 邊界夾制的前進後退
- **WHEN** 由含 N 步的棋譜建立 `Replayer`（長度 N+1，游標起始為 0）
- **THEN** 連續 `Next` 至末位後再 `Next` 仍停在末位（回報未移動）；連續 `Prev` 至 0 後再 `Prev` 仍停在 0

#### Scenario: 游標對應盤面
- **WHEN** 將游標跳轉至索引 i
- **THEN** `Current` 回傳第 i 手之後的盤面（與 `Timeline.At(i)` 一致），索引 0 為起始局面
