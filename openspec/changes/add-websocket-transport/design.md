# 設計：L2 WebSocket 傳輸與權威伺服器

> 延續 `online-play.md` 的傳輸分層：本變更實作 L2（WebSocket 傳輸 + Server Hub），
> L0 介面與 L1 回路已具備。核心不受影響——WSTransport 只是 `MoveTransport` 的另一實作。

## 架構

```
ClientA ──WS──┐                      ┌──WS── ClientB
  RemotePlayer │     Server (Hub)     │ RemotePlayer
  WSTransport ─┤   ┌──────────────┐   ├─ WSTransport
               └──▶│ Room          │◀──┘
                   │  GameState    │
                   │  RuleEngine ──┼─ 權威驗證（重用 core/rules）
                   └──────────────┘
```

- 客戶端：`WSTransport` 實作 `MoveTransport`，包成既有 `RemotePlayer`，接入既有 `Controller` 迴圈。
- 伺服器：`Hub` 管房與路由；`Room` 持對局狀態，每步以 `RuleEngine` 權威驗證後廣播。

## 為何權威伺服器（非 P2P）

伺服器為唯一事實來源 → 防作弊、防不同步、可仲裁逾時/斷線。象棋回合制、訊息稀疏，
單一權威伺服器負載極低。客戶端送「意圖走法」，**以伺服器廣播的 `move_applied` 為準**套用，
而非本地先行套用（避免分歧）。

## 協定 envelope

JSON `{ type, gameId, payload }`（對齊 `contracts.md §4`、`online-play.md` 訊息型別）：

| 方向 | type | payload（重點欄位） |
|---|---|---|
| C→S | `join` | （重連時帶 `gameId`） |
| C→S | `move` | UCCI 走法 |
| C→S | `resign` | — |
| S→C | `matched` | `color`(red/black)、`initialFen` |
| S→C | `move_applied` | `move`(UCCI)、`fen`、`turn` |
| S→C | `game_over` | `result`、`reason` |
| S→C | `error` | `reason`（僅回送出方，不廣播） |
| S→C | `state_sync` | `fen`、`moves[]`（UCCI 序列） |

走法一律 UCCI；中文記譜由客戶端 `Notation` 即時轉換顯示。

## 權威驗證流程（Room）

1. 收到 `move`：檢查送出方是否為**當前回合方**；否 → 回 `error`、忽略。
2. 以 `RuleEngine.ApplyMove`（不可變）驗證並套用；非法 → 僅回送出方 `error`、不廣播、狀態不變。
3. 合法 → 更新對局狀態，向**雙方**廣播 `move_applied`（新 FEN ＋ 輪走方）。
4. 若該步造成將死/困斃/和棋 → 續而向雙方廣播 `game_over`，其後拒絕進一步走子。

> 權威性正確性由重用 `core/rules`（已受 `conformance/` 覆蓋）保證；伺服器只負責「誰能走、何時廣播」，
> 不重寫規則。

## 配對與重連

- **配對**：`Hub` 維護等待佇列；湊滿兩人 → 建 `Room`＋全新 `GameState`、指派紅/黑、各回 `matched`。
- **重連**：`Room` 由伺服器保留（不因單方斷線即銷毀）；客戶端帶 `gameId` 重新 `join` → 回 `state_sync`
  （目前 FEN ＋ 走法序列）還原續局。

## 函式庫選型

採 **`github.com/coder/websocket`**（前 `nhooyr.io/websocket`）：context 友善、API 精簡，且
**可編入 WASM**——支援 Ebiten web export → LINE LIFF 的瀏覽器客戶端。`gorilla/websocket` 無法編入 WASM，故不採用。

## 可測試性

- **Room/Hub 邏輯**傳輸無關：可不經 WebSocket，直接餵解碼後的 envelope 測權威驗證/配對/重連（快、決定性）。
- **WSTransport ＋ 端對端**：以 `httptest.NewServer` 在 localhost 起真 server，行程內兩客戶端對接，
  測編解碼與連線。
- 連線/協定屬實作層，不另立語言中立 `conformance/` fixture（envelope 走 Go 測試；規則權威性已由 conformance 覆蓋）。

## 不在範圍

和棋/悔棋協商、聊天、觀戰、配對排隊策略、逾時時鐘、TLS/部署與水平擴展——後續另開變更。
