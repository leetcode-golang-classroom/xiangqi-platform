package record_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuanyu90221/xiangqi-platform/core/record"
)

func TestRecorderAddAndBuild(t *testing.T) {
	r := record.NewStartRecorder("紅", "黑")
	require.NoError(t, r.Add("h2e2"))
	require.NoError(t, r.Add("h9g7"))
	r.SetResult("red_win")

	rec := r.Record()
	assert.Equal(t, record.Format, rec.Format)
	assert.Equal(t, []string{"h2e2", "h9g7"}, rec.Moves)
	assert.Equal(t, "red_win", rec.Result)
	assert.Equal(t, "紅", rec.Red)
}

func TestRecorderRejectsIllegalMove(t *testing.T) {
	r := record.NewStartRecorder("紅", "黑")
	err := r.Add("e0e5") // 帥不能直接跳到 e5
	require.Error(t, err)
	assert.Empty(t, r.Record().Moves, "非法走法不應記錄")
}

func TestRecorderMarshalRoundTrip(t *testing.T) {
	r := record.NewStartRecorder("A", "B")
	require.NoError(t, r.Add("h2e2"))

	data, err := record.Marshal(r.Record())
	require.NoError(t, err)
	got, err := record.Unmarshal(data)
	require.NoError(t, err)
	assert.Equal(t, r.Record(), got)
}

func TestTimelineNavigation(t *testing.T) {
	r := record.NewStartRecorder("A", "B")
	require.NoError(t, r.Add("h2e2"))
	require.NoError(t, r.Add("h9g7"))

	tl, err := record.NewTimeline(r.Record())
	require.NoError(t, err)
	assert.Equal(t, 3, tl.Len(), "2 步 → 3 個盤面（含起始）")
	assert.Equal(t, record.NewStartRecorder("", "").Current().ToFEN(), tl.At(0).ToFEN(), "索引 0 為開局")
	assert.Equal(t, r.Current().ToFEN(), tl.At(tl.Len()-1).ToFEN(), "末索引為終局盤面")
}
