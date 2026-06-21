# 設計：對手抽象與 AI（Player + Human + AI）

> 抽象層（`player` 套件，純 Go）。**術語統一**：所有「取步者」相關者皆歸 `player` 套件
> ——`Player`/`Interactive` 介面與 `Human`/`AI` 實作；對局機制（`Session` 狀態、
> `Controller` 迴圈）在 `core/play`，單向依賴 `player`（`player` 不依賴 `core/play`，無 import cycle）。

## `Player`（對手抽象介面）

**職責**
- 統一三種取步者：**本地人類** `Human`、**AI**、**遠端玩家**（階段 3）。
- 對局迴圈不分對手種類，一律「請求一步」；差別僅在「如何決定」——人類漸進互動、AI 搜尋、遠端等待。
- 取步為**非同步**：回傳通道，決定後送出（適配人類即時輸入、AI 背景搜尋、遠端等待，皆不阻塞迴圈）。

**對外介面**
```
Player interface {
    Name() -> string
    RequestMove(game) -> <-chan Move          // 非同步：決定一步時送入通道
}
// Interactive：需人類漸進輸入者（Human）；Controller 以此分辨人類回合
Interactive interface {
    Player
    Tap(square) -> TapResult                  // 點選狀態機：選子/落點/改選/取消
    Selected() -> (square, bool)              // 供 UI 高亮
    Targets() -> [square]
}
```
- `Human`：點選狀態機；`RequestMove` 武裝一回合，點到合法落點時將該步送入通道。
- `AI`：背景 goroutine alpha-beta 搜尋，完成後送入通道（同步核心 `SelectMove` 對局結束回報 `ErrGameOver`）。
- `RemotePlayer`（階段 3）：等待 WebSocket 傳回後送入通道。

### 循序圖：統一對局迴圈（對手多型）
```mermaid
sequenceDiagram
    participant Loop as Controller.Step
    participant P as Player(介面)
    participant S as Session
    loop 每幀，直到對局結束
        Loop->>P: RequestMove(current)
        Note right of P: Human→等點擊<br/>AI→背景搜尋<br/>Remote→等WebSocket
        P-->>Loop: move（經通道，完成時）
        Loop->>S: Play(move)
        S-->>Loop: 換手 → 下一個 Player
    end
```

## `AI`（階段 2）

**職責**
- 實作 `Player`：以 negamax + alpha-beta 剪枝搜尋至指定深度選出一步。
- 透過 `RuleEngine` 產生候選走法並模擬（`ApplyMove` 不可變，利於搜尋樹展開）；以 `Game.PieceAt` 唯讀讀子供評估。
- 評估函數：子力價值（車 900／炮 450／馬 400／仕相 200／兵 100）+ 過河兵加成；終局將死/困斃視為行棋方必負（含步數修正，偏好較快將死），和棋為 0。
- 難度分級：`Easy`/`Medium`/`Hard` → 搜尋深度（深度越高棋力越強）。
- 決定性：穩定走法順序 + 固定 tie-break，相同盤面與深度回傳固定走法（利於單元測試）。

> AI 的最佳手取決於評估與排序，屬實作層決策，**不**納入語言中立 `conformance/`；以 `player` 套件的 Go 單元測試驗證（吃無保護子、一步將死、起手合法、終局回報錯誤、難度→深度）。

### 循序圖：AI 思考一步（背景搜尋）
```mermaid
sequenceDiagram
    participant Loop as Controller.Step
    participant AI as AI
    participant G as goroutine
    participant Eng as RuleEngine
    Loop->>AI: RequestMove(game)
    AI->>G: go SelectMove(game)
    AI-->>Loop: 通道（立即返回，不阻塞）
    loop negamax + alpha-beta（背景）
        G->>Eng: LegalMoves / ApplyMove 模擬
        Eng-->>G: 子盤面 → 評估（PieceAt 子力差）
    end
    G-->>Loop: bestMove（經通道，後續幀 Step 收到並套用）
```

## 人機對戰選邊（GUI）

`VsComputer(red, black, humanColor, ai)` 可指定人類執方（紅或黑），回傳 `Controller` 與 `Human`。
GUI 以鍵 `1`/`2` 選執紅/執黑；人類執黑時棋盤 180° 翻轉，使己方永遠在畫面下方。
開局輪紅，若紅為 AI 則 `Step` 自動先行。
