# 設計：本機回路傳輸與傳輸分層

## 背景

`player.RemotePlayer` 只依賴 `MoveTransport`（`Incoming() <-chan board.Move` / `Send(board.Move) error`）。
這層接縫讓「線上對手」與底層傳輸解耦。本變更在此接縫上補一個**本機回路**實作，並確立傳輸分層。

## 傳輸分層

| 層 | 角色 | 用途 | 本次交付 |
|---|---|---|---|
| L0 `MoveTransport` 介面 | 傳輸中立接縫 | 核心唯一依賴 | 既有（補規格） |
| L1 `LoopbackTransport`（記憶體內） | 兩端點以通道對接、零網路 | 本機測試、CI、本機對打 | ✅ 本次 |
| L2 WebSocket 傳輸 + Server Hub | 真正即時通道＋權威伺服器 | 正式線上、端對端測試 | ⏳ 日後（階段 3） |

「local 方便測試」的便利性來自 L0 介面接縫，而非 WebSocket：L1 在單一行程內即可把紅、黑兩端串成一局，
不需架站、不碰網路、決定性，最適合單元/整合測試。WebSocket（L2）屬正式線上傳輸，可日後以同一接縫接入。

## 為何正式線上選 WebSocket（非為了測試，而是目標平台）

目標平台含 **Android、WASM（Ebiten web export → LINE LIFF 網頁）**。WebSocket 是同時能跑在
瀏覽器 WASM、行動裝置與桌面的雙向即時通道；象棋為回合制，WebSocket + 權威伺服器已足夠
（選型 vs WebRTC 見 `docs/design/online-play.md`）。因此 WebSocket 由部署目標決定，與測試便利性無關。

> 函式庫提醒（留待 L2 實作時定案）：若要支援 WASM/LIFF，宜選可編入 WASM 的 WebSocket 函式庫
> （如 `github.com/coder/websocket`，前 `nhooyr.io/websocket`）；`gorilla/websocket` 無法編入 WASM。

## LoopbackTransport 設計要點

- **成對建立**：建構式回傳兩個對接端點（紅/黑）。端點 A 的 `Send` 寫入端點 B 的 `Incoming` 來源通道，反之亦然——即一條「交叉接線」。
- **決定性**：以緩衝通道承載，同序 `Send` 得同序 `Incoming`；不引入時間/亂數，利於可重現測試。
- **收尾**：提供關閉端點的方式；關閉後對向 `Incoming` 通道關閉、本端 `Send` 回報錯誤而非 panic（測試涵蓋）。
- **與權威性的關係**：L1 無權威伺服器，僅轉送走法；走法合法性仍由對局迴圈各自的 `Session`/`RuleEngine` 把關。L2 才引入伺服器端權威驗證。

## 不在範圍

WebSocket 傳輸實作、伺服器 Hub/Room、配對、斷線重連、協定 envelope 編解碼——維持階段 3 暫緩，
待本機回路與接縫穩定後再開獨立變更。
