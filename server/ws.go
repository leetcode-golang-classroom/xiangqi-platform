package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/coder/websocket"
)

// wsConn 將一條 WebSocket 連線包成 server.Conn：Send 將 envelope 序列化後寫出。
// 以指標作為身分（Hub/Room 以 == 比對座位），每條連線一個 *wsConn。
type wsConn struct {
	c   *websocket.Conn
	ctx context.Context
}

func (w *wsConn) Send(env Envelope) error {
	data, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return w.c.Write(w.ctx, websocket.MessageText, data)
}

// Handler 回傳處理 WebSocket 連線的 http.Handler：升級連線、逐則讀取交由 Hub 分派，
// 連線結束時通知 Hub 斷線（房間狀態保留以待重連）。
func (h *Hub) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			return
		}
		// 寫出用獨立 context，避免廣播被單一請求生命週期取消。
		conn := &wsConn{c: c, ctx: context.Background()}
		defer h.Disconnect(conn)
		defer c.CloseNow()

		for {
			_, data, err := c.Read(r.Context())
			if err != nil {
				return
			}
			_ = h.Handle(conn, data)
		}
	})
}
