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

JSON envelope `{ type, gameId, payload }`。訊息型別見 [online-play.md](online-play.md)。

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
