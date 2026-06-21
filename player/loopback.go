package player

import (
	"errors"
	"sync"

	"github.com/yuanyu90221/xiangqi-platform/core/board"
)

// ErrTransportClosed 表示傳輸端點已關閉，無法再送出走法。
var ErrTransportClosed = errors.New("player: 傳輸端點已關閉")

// loopbackBuffer 為各端接收通道的緩衝量，使本機回路在測試/本機對打中
// 不需嚴格交替即可連續送出數步而不阻塞。
const loopbackBuffer = 64

// loopbackPipe 為一對端點共用的交換核心：以共用鎖協調收發與關閉，
// 確保關閉後不會對已關閉通道送值（不 panic）——Send 與 Close 互斥，
// 關閉後的 Send 在取得鎖時先看到已關閉、直接回報錯誤而不觸碰通道。
type loopbackPipe struct {
	mu     sync.Mutex
	closed bool
	chA    chan board.Move // 端點 A 的接收通道（B.Send 寫入此）
	chB    chan board.Move // 端點 B 的接收通道（A.Send 寫入此）
}

// LoopbackTransport 為本機回路（in-process）的 MoveTransport 實作：成對運作，
// 一端 Send 的走法成為對向端 Incoming 送出的走法，零網路、決定性。
// 適用於本機測試與本機對打；正式線上請改用 WebSocket 傳輸（見 docs/design/online-play.md）。
type LoopbackTransport struct {
	pipe *loopbackPipe
	in   chan board.Move // 本端接收通道（對向端 Send 寫入此）
	out  chan board.Move // 對向端接收通道（本端 Send 寫入此）
}

// NewLoopbackPair 建立一對互相對接的本機回路端點（如紅、黑兩方）。
// a.Send 的走法會於 b.Incoming 送出，反之亦然。
func NewLoopbackPair() (a, b *LoopbackTransport) {
	p := &loopbackPipe{
		chA: make(chan board.Move, loopbackBuffer),
		chB: make(chan board.Move, loopbackBuffer),
	}
	a = &LoopbackTransport{pipe: p, in: p.chA, out: p.chB}
	b = &LoopbackTransport{pipe: p, in: p.chB, out: p.chA}
	return a, b
}

// Incoming 回傳本端接收對手走法的串流；端點關閉後此通道關閉。
func (t *LoopbackTransport) Incoming() <-chan board.Move { return t.in }

// Send 將本地走法送往對向端（成為其 Incoming）。端點關閉後回報 ErrTransportClosed。
func (t *LoopbackTransport) Send(m board.Move) error {
	t.pipe.mu.Lock()
	defer t.pipe.mu.Unlock()
	if t.pipe.closed {
		return ErrTransportClosed
	}
	t.out <- m
	return nil
}

// Close 關閉整條回路：兩端 Incoming 通道關閉、後續 Send 回報 ErrTransportClosed。
// 重複關閉為無操作。
func (t *LoopbackTransport) Close() error {
	t.pipe.mu.Lock()
	defer t.pipe.mu.Unlock()
	if t.pipe.closed {
		return nil
	}
	t.pipe.closed = true
	close(t.pipe.chA)
	close(t.pipe.chB)
	return nil
}
