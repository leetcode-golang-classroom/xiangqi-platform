package record_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuanyu90221/xiangqi-platform/core/record"
)

func sampleRecord() record.Record {
	return record.Record{
		Format:     record.Format,
		Red:        "甲",
		Black:      "乙",
		InitialFEN: "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1",
		Moves:      []string{"h2e2", "h9g7"},
	}
}

func TestReplayerLenAndStart(t *testing.T) {
	r, err := record.NewReplayer(sampleRecord())
	require.NoError(t, err)
	assert.Equal(t, 3, r.Len(), "2 步 → 3 盤面")
	assert.Equal(t, 0, r.Index(), "游標起始為 0")
}

func TestReplayerNextPrevClamp(t *testing.T) {
	r, err := record.NewReplayer(sampleRecord())
	require.NoError(t, err)

	assert.True(t, r.Next())
	assert.True(t, r.Next())
	assert.Equal(t, 2, r.Index())
	assert.False(t, r.Next(), "末位再前進應夾制且回報未移動")
	assert.Equal(t, 2, r.Index())

	assert.True(t, r.Prev())
	assert.True(t, r.Prev())
	assert.Equal(t, 0, r.Index())
	assert.False(t, r.Prev(), "起始再後退應夾制且回報未移動")
	assert.Equal(t, 0, r.Index())
}

func TestReplayerSeekClampAndCurrent(t *testing.T) {
	rec := sampleRecord()
	r, err := record.NewReplayer(rec)
	require.NoError(t, err)

	tl, err := record.NewTimeline(rec)
	require.NoError(t, err)

	r.Seek(99) // 越界 → 夾制至末位
	assert.Equal(t, r.Len()-1, r.Index())
	assert.Equal(t, tl.At(r.Len()-1).ToFEN(), r.Current().ToFEN())

	r.Seek(-5) // 越界 → 夾制至 0
	assert.Equal(t, 0, r.Index())
	assert.Equal(t, tl.At(0).ToFEN(), r.Current().ToFEN(), "索引 0 為起始局面")

	r.Seek(1)
	assert.Equal(t, tl.At(1).ToFEN(), r.Current().ToFEN())
}
