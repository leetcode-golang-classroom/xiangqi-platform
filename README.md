# 中國象棋平台（Xiangqi Platform）

手機可安裝、可離線對弈的中國象棋遊戲，主打**棋譜記錄與復盤**。
分階段：**① 單機雙人對戰（主目標）→ ② AI 對手 → ③ 線上對戰（暫緩）**。

技術棧：**Go + Ebitengine**；建構以 **Pants**；規格以 **OpenSpec** 管理。

---

## 文件與目錄架構

```
xiangqi-platform/
├── README.md            # 本檔：架構與建構說明
├── pants.toml           # Pants 建構設定（啟用 Go 後端）
├── go.mod               # Go module
├── docs/
│   └── DESIGN.md        # 元件職責、介面契約、互動循序圖（mermaid）
├── openspec/            # ★規格真相來源（spec-driven，由 OpenSpec 管理）
│   ├── specs/           #   現況規格，依領域分類
│   └── changes/         #   進行中的變更提案（proposal/spec/design/tasks）
├── core/                # 純邏輯核心（完全可移植、可單元測試）
│   ├── board/           #   座標系與棋子表示
│   ├── rules/           #   規則引擎：走法、合法性、勝負判定
│   ├── notation/        #   FEN / UCCI / 中文記譜轉換
│   ├── record/          #   棋譜序列化、漸進記錄與復盤
│   ├── storage/         #   本機棋譜持久化（Store / FileStore）
│   └── play/            #   對局互動控制器（點選狀態機，無圖形相依）
├── cmd/xiangqi/         # Ebiten 棋盤渲染層（//go:build ebiten 標籤隔離）
├── player/              # 對手抽象：Player 介面 + AI（alpha-beta，階段 2）
├── server/              # WebSocket 權威伺服器（階段 3）
└── conformance/         # ★跨語言一致性測試 fixtures（語言中立 JSON）
```

### 該讀哪份文件

| 想了解 | 看這裡 |
|---|---|
| 架構與如何建構 | 本 `README.md` |
| 元件職責 / 介面 / 循序圖 | `docs/DESIGN.md` |
| 各功能的「為什麼與要做什麼」規格 | `openspec/specs/` 與 `openspec/changes/` |
| 跨語言移植要符合的契約 | `docs/DESIGN.md` 末節 + `conformance/*.json` |

---

## 環境需求

| 工具 | 版本 | 用途 |
|---|---|---|
| Go | 1.25+ | 核心與伺服器 |
| Pants | 2.23+ | 建構與測試協調 |
| OpenSpec CLI | 1.2+ | 規格驅動開發流程 |

---

## 建構與測試

以 Pants 為主（一致的建構/測試介面）：

```bash
pants tailor ::               # 依原始碼產生/更新 BUILD 目標
pants test ::                 # 跑全 repo 測試（含 conformance）
pants run cmd/xiangqi:build   # 編譯單機 GUI 到 ./bin/xiangqi
pants run cmd/xiangqi:gui     # 直接執行單機雙人對弈 GUI（需圖形環境）
pants lint fmt ::             # 格式化與靜態檢查
```

開發期亦可直接用 Go：

```bash
go test ./...                      # 跑所有測試（不含 ebiten 標籤）
go build -tags ebiten -o bin/xiangqi ./cmd/xiangqi   # 編譯 GUI 到 bin/
go run -tags ebiten ./cmd/xiangqi                    # 啟動 GUI（需圖形環境）
```

> **Ebiten 渲染層以建構標籤隔離**：`cmd/xiangqi` 原始碼帶 `//go:build ebiten`，
> 故預設建構/測試（含無頭 CI、`pants test ::`）不編譯圖形相依；對局邏輯全在純邏輯
> 控制器 `core/play.Session`，可獨立單元測試。GUI 僅在 `-tags ebiten` 時編譯，需顯示
> 環境執行。棋子以嵌入的 CJK 字型（`cmd/xiangqi/assets/`）繪製正確中文字。
> Pants Go 後端不支援自訂 build tag，故 `cmd/xiangqi:build`/`:gui` 以 shell 目標
> 包一層原生 `go` 指令。GUI 為**人機對戰**，可自由選邊（AI 搜尋於背景執行）。
> 互動：滑鼠左鍵選子→點落點走子；鍵盤 `1` 執紅、`2` 執黑（執黑時棋盤翻轉，己方在下方）、
> `N` 同陣營新局、`U` 悔棋、`R` 認輸、`S` 匯出棋譜、`L` 進入復盤、`Q` 結束遊戲（任何模式皆可）。
> **棋譜**：對局結束時**自動存譜**（亦可隨時按 `S`），存成 `records/<id>.json`（`xiangqi-record-v1`）。
> **復盤**：按 `L` 由 `records/` 載入，`←`/`→` 逐手前進後退、`L` 切換下一份、`Esc` 返回。
> 對局結束時畫面中央顯示「棋局結束」與勝方。

> **⚠️ Pants + Go ≥1.24 暫時性 workaround**
> Pants 的 Go 後端（含至 2.28.0）會設定已被 Go 1.24 移除的 `GOEXPERIMENT`，
> 導致 `unknown GOEXPERIMENT coverageredesign`。下載 Pants venv 後請執行一次：
> ```bash
> ./scripts/patch-pants-go.sh   # 冪等；Pants 版本變更後需重跑；上游修復後移除
> ```

> 行動端打包（Android `.aar` / iOS `.framework`）由 `ebitenmobile bind`（gomobile）產生，為獨立步驟，不經 Pants 核心。

---

## 規格驅動開發流程（OpenSpec）

功能開發前先以 OpenSpec 對齊意圖，再實作：

```bash
openspec init . --tools claude                    # 初始化（一次）
openspec new change <name> --schema spec-driven   # 開新變更提案
# 撰寫 proposal / spec delta / design / tasks
openspec validate <name>                          # 驗證規格
openspec list / show <name> / status              # 查看
openspec archive <name>                           # 實作完成後合併進 specs/
```

`openspec/specs/` 是現況真相來源；`openspec/changes/` 是進行中的修改（含 ADDED/MODIFIED/REMOVED delta）。

---

## 跨語言移植

核心邏輯純化、契約語言中立。日後以 Rust/TS 等重寫核心時，只要符合：座標系、Xiangqi FEN、`xiangqi-record-v1` 棋譜格式、WebSocket 協定、`RuleEngine` 介面，並通過 `conformance/*.json` 黃金案例即相容。詳見 `docs/DESIGN.md`。
