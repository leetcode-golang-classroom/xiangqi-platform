package player_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/core/rules"
	"github.com/yuanyu90221/xiangqi-platform/player"
)

// RemotePlayer 須實作 Player 介面。
var _ player.Player = (*player.RemotePlayer)(nil)

// fakeTransport 為測試用的傳輸樁：以預置通道餵入對手走法。
type fakeTransport struct {
	in   chan board.Move
	sent []board.Move
}

func (f *fakeTransport) Incoming() <-chan board.Move { return f.in }
func (f *fakeTransport) Send(m board.Move) error     { f.sent = append(f.sent, m); return nil }

func TestRemotePlayerYieldsMoveFromTransport(t *testing.T) {
	tr := &fakeTransport{in: make(chan board.Move, 1)}
	m, err := board.ParseUCCI("h2e2")
	assert.NoError(t, err)
	tr.in <- m

	p := player.NewRemotePlayer("遠端對手", tr)
	assert.Equal(t, "遠端對手", p.Name())

	got := <-p.RequestMove(rules.NewGame())
	assert.Equal(t, m, got, "RequestMove 應轉送傳輸層收到的對手走法")
}
