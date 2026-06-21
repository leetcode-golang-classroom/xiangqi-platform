## Why

線上對戰（階段 3）的整體伺服器與正式傳輸尚未實作，但目前缺一個「不需架站、不碰網路」的方式來在本機驗證遠端對局流程：`player.RemotePlayer` 只依賴 `MoveTransport` 介面，現有測試也僅有一次性的 `fakeTransport` 樁，無法把「紅、黑兩個遠端端點互相收發」串成一局來測。

本變更先把**傳輸接縫正式納入規格**，並補上一個**本機回路傳輸 `LoopbackTransport`**：在單一行程內以通道把兩個端點對接，讓兩個 `RemotePlayer` 不經伺服器/網路即可完成整局，作為本機測試與本機對打的主力。正式線上仍走 WebSocket（見下），但屬日後變更，本次不實作。

## What Changes

- 將既有的傳輸接縫 `player.MoveTransport`（`Incoming()` / `Send()`）與 `player.RemotePlayer` 正式寫入 `online-play` 能力規格（目前僅存在於程式碼，無規格涵蓋）。
- 新增 `player.LoopbackTransport`：本機回路（in-process）傳輸實作——一方 `Send` 的走法成為對向端點的 `Incoming`，零網路、決定性、可單元測試。提供成對建立（紅/黑兩端點）的建構式。
- 確立「傳輸實作可替換」契約：核心（`rules`/`core/play`/`Controller`）只依賴 `MoveTransport`，更換傳輸實作不需改動核心。
- 記錄架構決策：**正式線上傳輸採 WebSocket**（理由：須同時支援 Android、WASM 與 LINE LIFF 等目標平台，WebSocket 為三者皆可用的雙向即時通道），但 WebSocket 實作與伺服器 Hub 仍為日後變更，本次僅交付介面接縫與本機回路實作。

## Capabilities

### New Capabilities
- `online-play`: 傳輸中立的遠端走法收發接縫（`MoveTransport`/`RemotePlayer`）與本機回路傳輸（`LoopbackTransport`）。

### Modified Capabilities
<!-- 無 -->

## Impact

- 程式碼：`player` 套件——`remote.go`（既有介面/`RemotePlayer`，僅補規格不改行為）；新增 `loopback.go`（`LoopbackTransport`）。核心 `rules`/`core/play` 不受影響。
- 測試：`player` 新增 `loopback_test.go`（成對收發、兩 `RemotePlayer` 經回路對局、關閉行為）。屬 Go 單元測試，傳輸協定/連線屬實作層，不納入語言中立 `conformance/`。
- 設計文件：`docs/design/online-play.md` 補「傳輸分層（回路/WebSocket）」與 WebSocket 選型理由；`docs/DESIGN.md` 元件圖已含 `Transport(WebSocket)`，補回路節點。
- 不在範圍：WebSocket 傳輸實作、伺服器 Hub/Room、配對、斷線重連（維持階段 3 暫緩）。
