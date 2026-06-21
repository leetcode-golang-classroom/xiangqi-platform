# online-play

交付 L2 正式線上對戰：WebSocket 傳輸、權威伺服器（配對/驗證/廣播/重連）與協定 envelope。

## MODIFIED Requirements

### Requirement: 正式線上傳輸採 WebSocket
為同時支援 Android、WASM 與 LINE LIFF 等目標平台，系統的正式線上傳輸 SHALL 以 WebSocket 實作 `MoveTransport`（雙向即時通道，配合權威伺服器）。WebSocket 傳輸 MUST 滿足同一 `MoveTransport` 接縫，使其與本機回路傳輸可互換、核心無須改動。`WSTransport` SHALL 建立 WebSocket 連線、編解碼協定 envelope，將伺服器確認的對手走法送入 `Incoming`、本地走法經 `Send` 上送伺服器。

#### Scenario: WebSocket 傳輸與回路可互換
- **WHEN** 以 `WSTransport` 取代 `LoopbackTransport` 包成 `RemotePlayer`
- **THEN** `RemotePlayer`／對局迴圈沿用相同 `MoveTransport` 介面，核心程式碼不需改動

#### Scenario: WSTransport 經 envelope 收發走法
- **WHEN** 伺服器送來一則 `move_applied`（對手走法），且本地呼叫 `Send` 一步
- **THEN** 對手走法於 `WSTransport.Incoming` 送出；本地走法被編成 `move` envelope 上送伺服器

## ADDED Requirements

### Requirement: WebSocket 協定 envelope
系統 SHALL 以 JSON envelope `{ type, gameId, payload }` 作為線上對戰的傳輸格式，欄位對齊 `contracts.md §4`。client→server 型別 SHALL 含 `join` / `move` / `resign`；server→client 型別 SHALL 含 `matched` / `move_applied` / `game_over` / `error` / `state_sync`。走法 payload 一律以 UCCI 表示。編解碼 MUST 為可逆往返（同一 envelope 編碼後解碼得等值結果）。

#### Scenario: envelope 編解碼往返
- **WHEN** 將一則 `move` envelope（含 `gameId` 與 UCCI 走法）編碼後再解碼
- **THEN** 得到型別、`gameId` 與走法皆等值的 envelope

#### Scenario: 未知型別回報錯誤
- **WHEN** 解碼到未知 `type` 的 envelope
- **THEN** 回報錯誤而非靜默忽略或 panic

### Requirement: 權威伺服器驗證走法
線上對局的伺服器 SHALL 為唯一事實來源，重用 `RuleEngine` 對每一步做權威驗證。伺服器 SHALL 僅接受當前回合方送來的走法：走法合法時 SHALL 套用並向**對局雙方**廣播 `move_applied`（含套用後盤面與輪走方），對局因該步結束時 SHALL 續而向雙方廣播 `game_over`（含結果與原因）；走法非法或來自非當前回合方時 SHALL 僅向送出方回 `error`、MUST NOT 廣播、且對局狀態不變。

#### Scenario: 合法走子廣播雙方
- **WHEN** 當前回合方送來一步合法走法
- **THEN** 伺服器套用該步並向紅、黑雙方廣播 `move_applied`（盤面與輪走方更新）

#### Scenario: 非法走子僅回報送出方
- **WHEN** 當前回合方送來一步非法走法
- **THEN** 伺服器僅向送出方回 `error`，不廣播，對局盤面與輪走方不變

#### Scenario: 非當前回合方走子被拒
- **WHEN** 非當前回合方送來任意走法
- **THEN** 伺服器回 `error` 並忽略該步，狀態不變

#### Scenario: 對局結束廣播 game_over
- **WHEN** 一步合法走法造成將死/困斃/和棋
- **THEN** 伺服器於 `move_applied` 後向雙方廣播 `game_over`（含結果與原因），其後拒絕進一步走子

### Requirement: 配對
系統 SHALL 提供 `Hub`，將送出 `join` 的客戶端配對成局：湊滿兩名時 SHALL 建立 `Room` 與全新對局狀態、指派紅/黑、並分別向兩客戶端回 `matched`（含己方執色與初始 FEN）。

#### Scenario: 兩客戶端配對成局
- **WHEN** 兩名客戶端先後送出 `join`
- **THEN** 伺服器建立一房與初始對局狀態，分別回 `matched`（一方紅、一方黑，含初始 FEN），其後即可在該 `gameId` 對局

### Requirement: 斷線重連
系統 SHALL 支援以 `gameId` 重連續局：客戶端帶既有 `gameId` 重新 `join` 時，伺服器 SHALL 回 `state_sync`（含目前 FEN 與已走的 UCCI 走法序列），使客戶端還原盤面並繼續對局；對局狀態 SHALL 由伺服器保留、不因單方斷線而遺失。

#### Scenario: 斷線後以 gameId 重連
- **WHEN** 一方對局中斷線後帶原 `gameId` 重新連上並 `join`
- **THEN** 伺服器回 `state_sync`（目前 FEN 與走法序列），客戶端還原盤面後可繼續對局
