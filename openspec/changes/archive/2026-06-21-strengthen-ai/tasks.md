## 1. 測試（先紅）

- [x] 1.1 新增 `TestAIAvoidsLosingCapture`：給「吃有保護子會淨虧」盤面，斷言 AI 不選該送子吃法（quiescence）
- [x] 1.2 新增 `TestPSTPrefersDevelopment`：等子力盤面下，斷言 AI 偏好位置較佳之著手
- [x] 1.3 新增 `TestNearBestStillUniqueOnTactics`：得子/將死仍為唯一選擇（ε 不破壞戰術）
- [x] 1.4 調整 `difficultyLabel` 門檻後補/改標籤對應測試

## 2. 位置評估

- [x] 2.1 新增 `pst` 表（紅方視角，各 kind），數值約 ±2～30；King 表全 0
- [x] 2.2 `evaluate` 對每子加位置分（黑方 rank 鏡射），維持行棋方視角

## 3. 靜默搜尋

- [x] 3.1 實作 `quiesce(g, alpha, beta, ply)`：終局分 / stand-pat / 僅吃子延伸 / ply 上限
- [x] 3.2 `negamax` 於 `depth==0` 改呼叫 `quiesce`

## 4. 走法排序

- [x] 4.1 實作 `orderMoves`（MVV-LVA，UCCI 平手）
- [x] 4.2 於 `negamax` 與 `quiesce` 套用排序

## 5. 選步與難度

- [x] 5.1 重構 `SelectMove`：root 分數算一次、ε 容差近佳手、保留 visits/Seed 兩路徑
- [x] 5.2 深度維持 Easy=1/Medium=2/Hard=3（實測 depth≥3 開局過慢，棋力改由評估+靜默搜尋提升）；`difficultyLabel` 門檻沿用 Easy/Hard 常數

## 6. 驗證與歸檔

- [x] 6.1 `go test ./...` 全綠（含既有 9 測試）、`gofmt -l .` 乾淨、`go vet ./...` 無誤
- [x] 6.2 `go build -tags ebiten -o bin/xiangqi ./cmd/xiangqi` 成功；量測 Hard 每步耗時
- [x] 6.3 `openspec validate strengthen-ai --strict` 通過
- [x] 6.4 `openspec archive strengthen-ai --yes`，確認 `openspec/specs/ai-opponent` 更新
