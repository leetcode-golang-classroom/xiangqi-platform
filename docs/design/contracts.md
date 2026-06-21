# 設計：跨語言移植契約

> 參考文件。任何語言重寫核心，只要符合下列契約即與既有資料/協定相容，並通過 `conformance/*.json`。

## 1. 座標與走法

- 縱線 file `a`–`i`（紅方視角左→右）、橫線 rank `0`–`9`（紅底→黑頂）。
- 一格 = file+rank，如 `e0`。走法 = UCCI `<from><to>`，如 `h2e2`。

## 2. FEN（Xiangqi Forsyth–Edwards Notation）

- 大寫=紅、小寫=黑。棋子字母：
  `K/k` 帥將・`A/a` 仕士・`B/b` 相象・`N/n` 馬・`R/r` 車・`C/c` 炮・`P/p` 兵卒。
- 盤面 10 行以 `/` 分隔，行序 rank9→rank0；數字表連續空格。
- 其後欄位：side-to-move（`w`/`b`）、保留欄、半步/全步計數。
- 開局：`rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1`

## 3. 棋譜容器 `xiangqi-record-v1`（JSON）

```json
{
  "format": "xiangqi-record-v1",
  "red": "玩家A", "black": "玩家B",
  "date": "2026-06-21", "result": "red_win",
  "initialFen": "rnbakabnr/9/1c5c1/...",
  "moves": ["b2e2", "h9g7"]
}
```
`moves` 一律存 UCCI；中文記譜由 `Notation` 即時轉換顯示。

## 4. WebSocket 協定

JSON envelope `{ type, gameId, payload }`（實作：`server.Envelope`）。走法一律以 UCCI 表示。
編解碼為可逆往返；未知 `type` 或損壞 JSON 一律回報錯誤（不靜默忽略、不 panic）。

**型別與 payload 形狀**

| 方向 | type | payload 欄位 |
|---|---|---|
| C→S | `join` | （無；重連時於頂層帶 `gameId`） |
| C→S | `move` | `move`（UCCI，如 `"h2e2"`） |
| C→S | `resign` | （無） |
| S→C | `matched` | `color`（`"red"`/`"black"`）、`initialFen` |
| S→C | `move_applied` | `move`（UCCI）、`fen`、`turn`（`"red"`/`"black"`） |
| S→C | `game_over` | `result`（`"red"`/`"black"`/`"draw"`）、`reason` |
| S→C | `error` | `reason`（僅回送出方，不廣播） |
| S→C | `state_sync` | `fen`、`moves`（UCCI 字串陣列，重連還原用） |

頂層 `gameId`：配對後由 `matched` 帶回，客戶端後續 `move`／重連 `join` 須帶同一值。
權威性：伺服器以 `RuleEngine` 驗證每一步，僅當前回合方合法走法才套用並向雙方廣播。

## 5. RuleEngine 介面

語言中立簽章見 [rule-engine.md](rule-engine.md)。

## 6. Storage 介面

本機棋譜持久化的語言中立契約；各平台自行實作（桌面檔案、行動端平台儲存）。

```
Save(id, record) -> Error          // id 須安全：禁路徑分隔與 ..
Load(id)         -> Record | Error  // 未知 id 回報不存在
List()           -> [Entry]         // 依 id 升冪排序；Entry{ id, red, black, date, result }
Delete(id)       -> Error
```
持久化的資料即 §3 的 `xiangqi-record-v1`，故跨語言一致性由棋譜格式保證（無需額外 fixture）。實作見 [notation-and-record.md](notation-and-record.md)。

## 7. 一致性測試

`conformance/*.json` 黃金案例（FEN / 合法走法 / 勝負 / 復盤 / 中文記譜清單）——每種語言實作以薄 harness 載入並斷言。
