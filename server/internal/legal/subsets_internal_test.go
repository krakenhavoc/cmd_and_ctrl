package legal

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestAllowedSubsetsDoesNotWalkDeadPrefixes — #624 review. choose_cards
// now runs every set size from Min to Max through allowedSubsets, and
// `scanned` counts only finished subsets, so at the sizes where fewer
// than `limit` subsets exist (k = n, k = n-1) the budget never trips.
// A walk that extends prefixes which can no longer reach k cards then
// visits every one of them — about 2^n nodes, under the read lock. With
// 64 candidates that never finishes; with pruning it is instant, and
// the answers are the same.
func TestAllowedSubsetsDoesNotWalkDeadPrefixes(t *testing.T) {
	const n, limit = 64, 12
	pool := make([]uuid.UUID, n)
	for i := range pool {
		pool[i] = uuid.New()
	}

	type result struct{ all, allButOne [][]uuid.UUID }
	done := make(chan result, 1)
	go func() {
		done <- result{
			all:       allowedSubsets(pool, n, limit, nil),
			allButOne: allowedSubsets(pool, n-1, limit, nil),
		}
	}()

	var got result
	select {
	case got = <-done:
	case <-time.After(10 * time.Second):
		t.Fatalf("allowedSubsets over %d candidates at k = n and n-1 did not finish in 10s: it is walking prefixes that cannot reach k cards", n)
	}

	if len(got.all) != 1 || len(got.all[0]) != n {
		t.Fatalf("k = n: got %d subsets, want exactly the whole pool once", len(got.all))
	}
	for i, id := range got.all[0] {
		if id != pool[i] {
			t.Fatalf("k = n: subset[%d] = %v, want pool order (%v)", i, id, pool[i])
		}
	}
	if len(got.allButOne) != limit {
		t.Fatalf("k = n-1: got %d subsets, want %d (the cap)", len(got.allButOne), limit)
	}
	for _, s := range got.allButOne {
		if len(s) != n-1 {
			t.Fatalf("k = n-1: a subset has %d cards, want %d", len(s), n-1)
		}
	}
}

// TestAllowedSubsetsKeepsPoolOrderAfterPruning pins that pruning changes
// how much is walked, not what is found: every 2-subset of four cards,
// in the same lexicographic pool order the unpruned walk produced.
func TestAllowedSubsetsKeepsPoolOrderAfterPruning(t *testing.T) {
	pool := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	want := [][2]int{{0, 1}, {0, 2}, {0, 3}, {1, 2}, {1, 3}, {2, 3}}
	got := allowedSubsets(pool, 2, 12, nil)
	if len(got) != len(want) {
		t.Fatalf("got %d pairs, want %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i][0] != pool[w[0]] || got[i][1] != pool[w[1]] {
			t.Fatalf("pair %d = %v, want pool[%d], pool[%d]", i, got[i], w[0], w[1])
		}
	}
}
