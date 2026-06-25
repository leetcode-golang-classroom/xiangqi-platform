package rules_test

import (
	"fmt"
	"testing"
	"github.com/yuanyu90221/xiangqi-platform/core/rules"
)

// TestNoKingCapturePossibleInLegalMoves 用多條路徑搜尋是否存在「合法走法可吃將帥」的局面
func TestNoKingCapturePossibleInLegalMoves(t *testing.T) {
	// BFS 展開，最多搜 4000 個局面
	type node struct {
		g    *rules.Game
		path []string
	}
	start := node{g: rules.NewGame()}
	queue := []node{start}
	seen := map[string]bool{}
	count := 0
	maxNodes := 4000

	for len(queue) > 0 && count < maxNodes {
		cur := queue[0]
		queue = queue[1:]
		count++

		fen := cur.g.ToFEN()
		if seen[fen] {
			continue
		}
		seen[fen] = true

		if cur.g.Result().Over {
			continue
		}

		for _, m := range cur.g.LegalMoves() {
			victim := cur.g.PieceAt(m.To)
			if !victim.IsEmpty() && victim.Kind() == 'k' {
				t.Errorf("BUG: legal move %s captures king at step %d\nPath: %v\nFEN: %s",
					m, count, append(cur.path, m.String()), fen)
				return
			}
			ng, err := cur.g.ApplyMove(m)
			if err != nil {
				continue
			}
			newPath := make([]string, len(cur.path)+1)
			copy(newPath, cur.path)
			newPath[len(cur.path)] = m.String()
			queue = append(queue, node{g: ng, path: newPath})
		}
	}
	fmt.Printf("Explored %d unique positions\n", len(seen))
}
