# online-play Specification

## Purpose
TBD - created by archiving change add-loopback-transport. Update Purpose after archive.
## Requirements
### Requirement: 傳輸中立的走法收發介面
系統 SHALL 提供 `MoveTransport` 介面作為遠端走法的收發接縫：`Incoming()` 回傳對手走法的串流通道、`Send(move)` 將本地走法送往遠端並回報傳輸層錯誤。系統 SHALL 提供 `RemotePlayer`（實作 `Player`），以 `MoveTransport` 為走法來源，將遠端對手納入統一對局迴圈。核心邏輯（`rule-engine`／對局迴圈）MUST NOT 依賴任何具體傳輸；更換傳輸實作 MUST 不需改動核心。

#### Scenario: RemotePlayer 經傳輸層轉送對手走法
- **WHEN** 以某 `MoveTransport` 建立 `RemotePlayer`，傳輸層送入一步對手走法
- **THEN** `RemotePlayer.RequestMove` 的通道收到同一步走法，供對局迴圈套用

#### Scenario: 核心不依賴具體傳輸
- **WHEN** 將 `RemotePlayer` 的傳輸層由一種實作替換為另一種（皆實作 `MoveTransport`）
- **THEN** `Player`／對局迴圈程式碼不需改動即可運作

### Requirement: 本機回路傳輸
系統 SHALL 提供 `LoopbackTransport`：在單一行程內成對運作的傳輸實作，不經網路或伺服器。成對建立兩個端點後，一端 `Send` 的走法 SHALL 成為對向端點 `Incoming` 串流送出的走法（反之亦然）。`LoopbackTransport` 用於本機測試與本機對打，行為 MUST 決定性（同序輸入得同序輸出）。

#### Scenario: 一端送出成為對向端收到
- **WHEN** 成對建立紅、黑兩端點，紅端 `Send` 一步走法
- **THEN** 黑端 `Incoming` 串流送出同一步走法；黑端 `Send` 的走法亦於紅端 `Incoming` 送出

#### Scenario: 兩個 RemotePlayer 經回路對局
- **WHEN** 以成對的 `LoopbackTransport` 各自包成 `RemotePlayer`（紅/黑），交替由各方 `Send` 走法
- **THEN** 雙方皆於對向通道依序收到對手走法，可串成一局而無需伺服器或網路

### Requirement: 正式線上傳輸採 WebSocket
為同時支援 Android、WASM 與 LINE LIFF 等目標平台，系統的正式線上傳輸 SHALL 以 WebSocket 實作 `MoveTransport`（雙向即時通道，配合權威伺服器）。WebSocket 傳輸 MUST 滿足同一 `MoveTransport` 接縫，使其與本機回路傳輸可互換、核心無須改動。

> 註：WebSocket 傳輸實作與伺服器 Hub／配對／斷線重連屬日後變更（階段 3），本能力此次僅交付介面接縫與本機回路實作。

#### Scenario: WebSocket 傳輸與回路可互換
- **WHEN** 日後以 WebSocket 實作 `MoveTransport` 取代本機回路
- **THEN** `RemotePlayer`／對局迴圈沿用相同介面，核心程式碼不需改動

