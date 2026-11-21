package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// apbCounters reads a counter off a battlefield permanent.
func ApbCounters(g *game.Game, id uuid.UUID, kind string) int {
	n := 0
	g.ReadSnapshot(func() {
		if c := FindBattlefieldCardForTest(g, id); c != nil {
			n = c.Counters[kind]
		}
	})
	return n
}

// apbRequireAnyPlayerRows fails unless every one of the card's rows
// listed in `want` is an any-player row, and no other is.
func ApbRequireAnyPlayerRows(t *testing.T, g *game.Game, id uuid.UUID, want ...int) {
	t.Helper()
	rows := game.ActivatedAbilitiesForCard(*FindBattlefieldCardForTest(g, id))
	isWanted := map[int]bool{}
	for _, i := range want {
		isWanted[i] = true
	}
	for i, r := range rows {
		if r.AnyPlayer != isWanted[i] {
			t.Errorf("row %d (%q) AnyPlayer = %v, want %v", i, r.Label, r.AnyPlayer, isWanted[i])
		}
		// Answers (ADR 0142) is not a bot price; only the amounts are.
		priced := r.Purpose
		priced.Answers = 0
		// #4302: nor is a target entry, which says what the row does to
		// its pick and buys the activator nothing.
		priced.Targets = nil
		if !priced.IsZero() {
			t.Errorf("row %d (%q) declares a purpose %+v; none of this batch's rows should", i, r.Label, r.Purpose)
		}
	}
}
