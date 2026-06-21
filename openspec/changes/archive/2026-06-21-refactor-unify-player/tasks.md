## 1. 測試先行

- [x] 1.1 `Human` 單元測試：選子取得落點、點落點產出走法（送通道）、改選、取消、RequestMove 重新武裝
- [x] 1.2 改寫 `Session` 測試（收斂為 Play/Undo/Resign/Record，移除 Tap）
- [x] 1.3 改寫 `Controller` 測試（統一 Step 迴圈：AI對AI、人機、結束停止、非人類忽略點擊）
- [x] 1.4 AI 補 `RequestMove` 測試（通道送出合法走法）

## 2. 實作

- [x] 2.1 `Session` 收斂：移除選取狀態與 Tap，保留 `Play`/`Undo`/`Resign`/`Record`/`Current`/`Turn`/`Outcome`
- [x] 2.2 新增 `play.Human`（選取狀態機 + `RequestMove`）與 `play.Player` 結構化介面、`TapResult`
- [x] 2.3 `Controller` 重寫為統一迴圈：`Step`/`Undo`/`Resign`/`CurrentHuman`/`Thinking`
- [x] 2.4 `player.AI` 補 `RequestMove`（背景搜尋）
- [x] 2.5 `cmd/xiangqi` 改用統一迴圈，點擊餵給當前 `Human`

## 3. 文件與驗收

- [x] 3.1 更新 `docs/design`（對手抽象、循序圖）與 README
- [x] 3.2 `go test ./...` 全綠、`pants test ::` 全綠、`go build -tags ebiten` 通過
- [x] 3.3 `openspec validate refactor-unify-player` 通過
