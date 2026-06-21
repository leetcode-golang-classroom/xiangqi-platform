// Package client 提供連向權威伺服器的 WebSocket 傳輸（L2 客戶端）。
//
// WSTransport 實作 player.MoveTransport，與本機回路 LoopbackTransport 在同一接縫上互換，
// 故核心（RemotePlayer／對局迴圈）無須改動。採 github.com/coder/websocket（可編入 WASM，
// 支援 Ebiten web export → LINE LIFF 的瀏覽器客戶端）。
package client

import (
	"context"
	"errors"
	"sync"

	"github.com/coder/websocket"
	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/player"
	"github.com/yuanyu90221/xiangqi-platform/server"
)

// ErrNotMatched 表示尚未完成配對（未知 gameId），無法送出走法。
var ErrNotMatched = errors.New("client: 尚未配對，無法送出走法")

const incomingBuffer = 64

// WSTransport 為客戶端 WebSocket 傳輸。本地走法經 Send 編成 move envelope 上送；
// 伺服器確認的走法（move_applied）經讀取迴圈解碼後送入 Incoming。
type WSTransport struct {
	c       *websocket.Conn
	in      chan board.Move
	matched chan string // 配對完成時送出己方執色（buffered 1）
	ctx     context.Context

	mu     sync.Mutex
	gameID string
	closed bool
}

var _ player.MoveTransport = (*WSTransport)(nil)

// Dial 連向 WS server、送出 join、啟動讀取迴圈。
func Dial(ctx context.Context, url string) (*WSTransport, error) {
	c, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		return nil, err
	}
	t := &WSTransport{
		c:       c,
		in:      make(chan board.Move, incomingBuffer),
		matched: make(chan string, 1),
		ctx:     context.Background(),
	}
	join, err := server.Encode(server.TypeJoin, "", nil)
	if err != nil {
		_ = c.CloseNow()
		return nil, err
	}
	if err := c.Write(ctx, websocket.MessageText, join); err != nil {
		_ = c.CloseNow()
		return nil, err
	}
	go t.readLoop()
	return t, nil
}

// readLoop 持續讀取伺服器訊息並分派；連線結束時關閉 Incoming。
func (t *WSTransport) readLoop() {
	defer close(t.in)
	for {
		_, data, err := t.c.Read(t.ctx)
		if err != nil {
			return
		}
		env, err := server.Decode(data)
		if err != nil {
			continue // 忽略無法解碼的訊息，不中斷連線
		}
		switch env.Type {
		case server.TypeMatched:
			var p server.MatchedPayload
			if server.DecodePayload(env, &p) == nil {
				t.setGameID(env.GameID)
				select {
				case t.matched <- p.Color:
				default:
				}
			}
		case server.TypeStateSync:
			// 重連還原：更新 gameId（盤面還原由上層依 state_sync 處理）。
			t.setGameID(env.GameID)
		case server.TypeMoveApplied:
			var p server.MoveAppliedPayload
			if server.DecodePayload(env, &p) == nil {
				if m, err := board.ParseUCCI(p.Move); err == nil {
					t.in <- m
				}
			}
		case server.TypeGameOver, server.TypeError:
			// 結束/錯誤通知：不影響走法串流，保留 Incoming 由連線關閉時收尾。
		}
	}
}

func (t *WSTransport) setGameID(id string) {
	t.mu.Lock()
	t.gameID = id
	t.mu.Unlock()
}

// WaitMatched 阻塞至配對完成，回傳己方執色（"red"/"black"）。
func (t *WSTransport) WaitMatched(ctx context.Context) (string, error) {
	select {
	case color := <-t.matched:
		return color, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// Incoming 回傳伺服器確認的對手走法串流。
func (t *WSTransport) Incoming() <-chan board.Move { return t.in }

// Send 將本地走法編成 move envelope 上送伺服器。配對前送出回報 ErrNotMatched。
func (t *WSTransport) Send(m board.Move) error {
	t.mu.Lock()
	gid := t.gameID
	t.mu.Unlock()
	if gid == "" {
		return ErrNotMatched
	}
	raw, err := server.Encode(server.TypeMove, gid, server.MovePayload{Move: m.String()})
	if err != nil {
		return err
	}
	return t.c.Write(t.ctx, websocket.MessageText, raw)
}

// Close 關閉連線。
func (t *WSTransport) Close() error {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return nil
	}
	t.closed = true
	t.mu.Unlock()
	return t.c.Close(websocket.StatusNormalClosure, "")
}
