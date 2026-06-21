## Why

`online-play` 能力目前只到 L0 介面（`MoveTransport`）與 L1 本機回路（`LoopbackTransport`）；
正式線上對戰（L2）尚未實作。既有規格已確立「正式線上傳輸採 WebSocket」的決策，但實作、
伺服器權威驗證、配對與斷線重連仍標為日後變更。本變更交付 L2：讓兩名真實玩家能跨網路對局。

象棋為回合制，採 **WebSocket + 權威伺服器**：伺服器是唯一事實來源，重用既有 `RuleEngine`
驗證每一步，杜絕作弊與不同步。客戶端的 WebSocket 傳輸只是 `MoveTransport` 的另一個實作，
與 L1 回路可互換，對局核心（`rules`/`core/play`/`Controller`/`RemotePlayer`）完全不需改動。

## What Changes

- **客戶端 `WSTransport`**（實作 `MoveTransport`）：建立 WebSocket 連線、編解碼 JSON envelope
  `{ type, gameId, payload }`，將伺服器確認的對手走法送入 `Incoming`、本地走法經 `Send` 上送。
- **伺服器 `Hub` / `Room`**（`server` 套件）：
  - **配對**：`join` → 將兩名等待中的客戶端配成一房，建立對局狀態、指派紅/黑、回 `matched`。
  - **權威驗證**：重用 `RuleEngine` 驗證每一步；僅當前回合方可走；合法 → 套用並向雙方廣播
    `move_applied`，對局結束再廣播 `game_over`；非法 → 僅回送出方 `error`、不廣播、狀態不變。
  - **斷線重連**：客戶端帶 `gameId` 重連 → 回 `state_sync`（目前 FEN ＋ 走法序列）還原續局。
- **協定 envelope** 正式納入規格，對齊 `contracts.md §4` 與 `online-play.md` 的訊息型別。
- 更新既有「正式線上傳輸採 WebSocket」需求：由「日後變更」轉為「已實作」。

## Capabilities

### New Capabilities
<!-- 無（皆併入既有 online-play 能力）-->

### Modified Capabilities
- `online-play`: 新增 WebSocket 傳輸、權威伺服器（配對/驗證/廣播/重連）與協定 envelope；
  既有「正式線上傳輸採 WebSocket」需求由日後轉為已實作。

## Impact

- 程式碼：新增 `server` 套件（`Hub`/`Room`/envelope 編解碼、權威驗證重用 `core/rules`）；
  `player` 或新 `client/transport` 新增 `WSTransport`（實作 `MoveTransport`）。核心 `rules`/`core/play` 不受影響。
- 相依：新增 WebSocket 函式庫 `github.com/coder/websocket`（前 `nhooyr.io/websocket`）——**可編入 WASM**，
  以支援 Ebiten web export → LINE LIFF；`gorilla/websocket` 無法編入 WASM 故不採用。
- 測試：Go 單元/整合測試——envelope 編解碼往返、權威驗證（合法廣播/非法僅回送方/非當前回合方拒絕）、
  配對、`state_sync` 重連；以 `httptest` 在 localhost 起真 server、行程內兩客戶端對接。連線/協定屬實作層，
  權威性正確由重用 `RuleEngine`（已受 `conformance/` 覆蓋）保證，不另立語言中立 fixture。
- 文件：`online-play.md` 既有協定/循序圖已涵蓋，補實作對應；`contracts.md §4` 補 envelope 欄位細節。
- 不在範圍：和棋協商 UI、悔棋協商、聊天、觀戰、配對排隊策略、TLS/部署細節（後續另開）。
