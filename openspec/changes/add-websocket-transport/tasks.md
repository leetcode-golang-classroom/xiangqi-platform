## 1. 協定 envelope（測試先行）

- [x] 1.1 測試：`move` envelope 編碼→解碼往返等值（型別/gameId/UCCI 走法）
- [x] 1.2 測試：未知 `type` 解碼回報錯誤（不 panic）
- [x] 1.3 實作：envelope 型別與編解碼（client→server：`join`/`move`/`resign`；server→client：`matched`/`move_applied`/`game_over`/`error`/`state_sync`）

## 2. 權威伺服器 Hub / Room（測試先行）

- [x] 2.1 測試：合法走子 → 套用並向雙方廣播 `move_applied`（盤面/輪走方更新）
- [x] 2.2 測試：非法走子 → 僅回送出方 `error`、不廣播、狀態不變
- [x] 2.3 測試：非當前回合方走子 → 回 `error` 並忽略
- [x] 2.4 測試：將死/困斃/和棋 → `move_applied` 後廣播 `game_over`，其後拒絕走子
- [x] 2.5 實作：`Room` 持有對局狀態、重用 `core/rules` 權威驗證、回合檢查、廣播
- [x] 2.6 實作：`Hub` 管理房間與訊息路由（傳輸無關，可不經 WebSocket 直接測 room 邏輯）

## 3. 配對與重連（測試先行）

- [x] 3.1 測試：兩 `join` → 配成一房、指派紅/黑、各回 `matched`（含初始 FEN）
- [x] 3.2 測試：帶 `gameId` 重連 → 回 `state_sync`（目前 FEN ＋ 走法序列），狀態保留
- [x] 3.3 實作：配對佇列、`Room` 建立、`state_sync` 還原

## 4. 客戶端 WSTransport（測試先行）

- [x] 4.1 測試（`httptest` 起 localhost server，行程內客戶端）：`Send` 上送 `move`、收到 `move_applied` 於 `Incoming` 送出
- [x] 4.2 測試：`WSTransport` 滿足 `MoveTransport`，可包成 `RemotePlayer` 與回路互換（核心不改）
- [x] 4.3 實作：`WSTransport`（`github.com/coder/websocket`，可編入 WASM）連線、編解碼、收發

## 5. 文件

- [x] 5.1 `docs/design/online-play.md`：L2 段落補實作對應（套件路徑、函式庫、與循序圖對照）
- [x] 5.2 `docs/design/contracts.md §4`：補 envelope 欄位細節（型別清單、payload 形狀）

## 6. 驗收

- [x] 6.1 `go test ./...` 全綠（含 `-race`，伺服器併發）、`pants test server:: client::` 全綠
- [x] 6.2 `gofmt -l .` 乾淨、`go vet ./...` 無誤
- [x] 6.3 `openspec validate add-websocket-transport` 通過
- [~] 6.4 端對端：自動化 `httptest` 行程內真 WS 對接（`client` 套件，紅走一步→黑 `Incoming` 收到）已覆蓋對局收發；Hub 層重連 `state_sync` 已測。經真 WS 客戶端的「斷線重連續局」完整手測待後續（功能已具備）
