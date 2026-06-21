# 設計：線上對戰（Transport + Server Hub）

> 平台層、協定契約。**階段 3，暫緩**。象棋為回合制，採 WebSocket + 權威伺服器。
> 選型理由（vs WebRTC）見專案規劃；協定欄位定義見 [contracts.md](contracts.md)。

## 元件職責

- `Transport`（客戶端）：WebSocket 連線，編解碼 JSON envelope；作為 `RemotePlayer` 的走法來源。
- `Server Hub / Room`：配對、房間管理、**權威驗證**（重用 `RuleEngine` 驗證每一步）、廣播、斷線重連。

## 協定（JSON envelope `{ type, gameId, payload }`）

- client→server：`join` / `move` / `resign` / `draw_offer` / `draw_accept` / `takeback_request` / `chat`
- server→client：`matched` / `move_applied` / `game_over` / `error` / `opponent_left` / `state_sync`

## 循序圖：線上走子（權威伺服器）
```mermaid
sequenceDiagram
    participant A as ClientA
    participant S as Server(Hub/Room)
    participant E as RuleEngine
    participant B as ClientB
    A->>S: {type:"move", from, to}
    S->>S: 檢查是否 A 的回合
    S->>E: ApplyMove(state, move)
    alt 合法
        E-->>S: newState
        S-->>A: {type:"move_applied", move, fen, turn}
        S-->>B: {type:"move_applied", move, fen, turn}
        opt 對局結束
            S-->>A: {type:"game_over", result, reason}
            S-->>B: {type:"game_over", result, reason}
        end
    else 非法
        E-->>S: Error
        S-->>A: {type:"error"}（狀態不變、不廣播）
    end
```

## 循序圖：配對與斷線重連
```mermaid
sequenceDiagram
    participant A as ClientA
    participant S as Server
    participant B as ClientB
    A->>S: {type:"join"}
    B->>S: {type:"join"}
    S->>S: 配對成功，建立 Room + GameState
    S-->>A: {type:"matched", color:"red", initialFen}
    S-->>B: {type:"matched", color:"black", initialFen}
    Note over A: A 斷線…重新連上
    A->>S: {type:"join", gameId}
    S-->>A: {type:"state_sync", fen, moves[]}
    Note over A: 還原盤面，續局
```
