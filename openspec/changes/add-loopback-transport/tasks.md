## 1. 測試先行（player 套件）

- [x] 1.1 `loopback_test.go`：成對端點，紅端 `Send` 之走法於黑端 `Incoming` 送出（反向亦然）
- [x] 1.2 兩個 `RemotePlayer` 經成對 `LoopbackTransport` 交替走子，雙方依序收到對手走法（串成一局）
- [x] 1.3 關閉/收尾行為：端點關閉後 `Incoming` 通道關閉、`Send` 回報錯誤（不 panic）

## 2. 實作（player 套件）

- [x] 2.1 `loopback.go`：`LoopbackTransport` 實作 `MoveTransport`；成對建構式（回傳紅/黑兩端點，互相對接）
- [x] 2.2 一端 `Send` → 對向端 `Incoming`；決定性（同序輸入得同序輸出）

## 3. 設計文件

- [x] 3.1 `docs/design/online-play.md`：補「傳輸分層（本機回路／WebSocket）」與 WebSocket 選型理由（Android／WASM／LIFF）
- [x] 3.2 `docs/DESIGN.md`：元件圖補本機回路傳輸節點（與既有 `Transport(WebSocket)` 並列）

## 4. 驗收

- [x] 4.1 `go test ./...` 全綠（含 `-race`）、`pants test player::` 全綠
- [x] 4.2 `gofmt -l .` 乾淨、`go vet ./...` 無誤
- [x] 4.3 `openspec validate add-loopback-transport` 通過
