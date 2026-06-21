## 1. 測試先行

- [x] 1.1 `core/record` 新增 `Replayer` 單元測試：長度、Next/Prev 邊界夾制、Seek 夾制、Current 對應盤面

## 2. 實作

- [x] 2.1 `record.Replayer`：`NewReplayer` / `Len` / `Index` / `Current` / `Next` / `Prev` / `Seek`

## 3. GUI（建構標籤 ebiten）

- [x] 3.1 復盤模式：`L` 由 `records/` 載入、`←`/`→` 導覽、`L` 切換下一份、`Esc` 返回
- [x] 3.2 對局結束自動存譜（保留手動 `S`），顯示存檔路徑
- [x] 3.3 對局結束橫幅：中央顯示「棋局結束」與勝方

## 4. 驗收

- [x] 4.1 `go test ./...` 全綠、`pants test ::` 全綠、`go build -tags ebiten` 通過
- [x] 4.2 `openspec validate add-gui-replay` 通過
