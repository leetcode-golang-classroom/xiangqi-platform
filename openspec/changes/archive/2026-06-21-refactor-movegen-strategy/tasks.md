## 1. TDD：補齊各子測試（重構前）

- [x] 1.1 新增 仕/相/帥/兵 的 conformance 走法案例（advisor/elephant/king/pawn）
- [x] 1.2 確認新案例對現有 switch 實作全綠（含修正手寫 FEN 錯誤）

## 2. 重構為函式值註冊表

- [x] 2.1 定義 `targetFn` 型別與 `pieceTargets` 註冊表
- [x] 2.2 抽出共用 helper（`canLand` 落點過濾、方向常數）
- [x] 2.3 逐子抽出 `rookTargets`/`cannonTargets`/`horseTargets`/`elephantTargets`/`advisorTargets`/`kingTargets`/`pawnTargets`
- [x] 2.4 `pseudoTargets` 改為查表分派

## 3. 驗收（行為不變）

- [x] 3.1 `go test ./...` 全綠
- [x] 3.2 `pants test ::` 全綠
- [x] 3.3 `go vet ./...` 無誤
