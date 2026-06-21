package storage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuanyu90221/xiangqi-platform/core/record"
	"github.com/yuanyu90221/xiangqi-platform/core/storage"
)

func sample(red, black, date, result string) record.Record {
	return record.Record{
		Format:     record.Format,
		Red:        red,
		Black:      black,
		Date:       date,
		Result:     result,
		InitialFEN: "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1",
		Moves:      []string{"h2e2", "h9g7"},
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	s, err := storage.NewFileStore(t.TempDir())
	require.NoError(t, err)

	rec := sample("甲", "乙", "2026-06-21", "red_win")
	require.NoError(t, s.Save("game-001", rec))

	got, err := s.Load("game-001")
	require.NoError(t, err)
	assert.Equal(t, rec, got, "載回應與存入一致")
}

func TestLoadUnknown(t *testing.T) {
	s, err := storage.NewFileStore(t.TempDir())
	require.NoError(t, err)

	_, err = s.Load("nope")
	require.Error(t, err, "未知 ID 應回報錯誤")
}

func TestListSortedWithMeta(t *testing.T) {
	s, err := storage.NewFileStore(t.TempDir())
	require.NoError(t, err)

	require.NoError(t, s.Save("b-game", sample("B紅", "B黑", "2026-06-20", "draw")))
	require.NoError(t, s.Save("a-game", sample("A紅", "A黑", "2026-06-21", "black_win")))

	entries, err := s.List()
	require.NoError(t, err)
	require.Len(t, entries, 2)

	assert.Equal(t, []string{"a-game", "b-game"}, []string{entries[0].ID, entries[1].ID}, "依 id 升冪排序")
	assert.Equal(t, storage.Entry{
		ID: "a-game", Red: "A紅", Black: "A黑", Date: "2026-06-21", Result: "black_win",
	}, entries[0], "Entry 含中介資料")
}

func TestDelete(t *testing.T) {
	s, err := storage.NewFileStore(t.TempDir())
	require.NoError(t, err)

	require.NoError(t, s.Save("g1", sample("甲", "乙", "2026-06-21", "red_win")))
	require.NoError(t, s.Delete("g1"))

	entries, err := s.List()
	require.NoError(t, err)
	assert.Empty(t, entries, "刪除後不再列出")

	_, err = s.Load("g1")
	assert.Error(t, err, "刪除後載入應回報錯誤")
}

func TestDeleteUnknown(t *testing.T) {
	s, err := storage.NewFileStore(t.TempDir())
	require.NoError(t, err)
	assert.Error(t, s.Delete("ghost"), "刪除未知 ID 應回報錯誤")
}

func TestRejectsUnsafeID(t *testing.T) {
	s, err := storage.NewFileStore(t.TempDir())
	require.NoError(t, err)

	rec := sample("甲", "乙", "2026-06-21", "red_win")
	for _, id := range []string{"../escape", "sub/dir", "..", "a/../b"} {
		t.Run(id, func(t *testing.T) {
			assert.Error(t, s.Save(id, rec), "不安全 ID 應被拒絕")
			_, err := s.Load(id)
			assert.Error(t, err)
			assert.Error(t, s.Delete(id))
		})
	}
}
