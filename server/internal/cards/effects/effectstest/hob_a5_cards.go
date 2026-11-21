package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- the "second card each turn" creatures -------------------------------

// hobA5Draw has p draw one card and lets the stack settle, answering a
// target prompt with pick when one is open.
func HobA5Draw(t *testing.T, g *game.Game, p *game.Player, pick *game.TargetRef) {
	t.Helper()
	g.WithWriteLock(func() { _ = g.DrawNForEffect(p.ID, 1) })
	AssistCardSettle(t, g, p.ID, pick, true)
}

func HobA5Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return RingBFCard(t, g, id).Counters[game.CounterPlusOne]
}

// hobA5SecondDrawGrows runs the shared shape of Knights of Dol Amroth and
// Lakeshore Apothecary: a counter on the second draw, none on the first or
// the third, and none for a draw by another player.
func HobA5SecondDrawGrows(t *testing.T, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	g := NewCatalogGame(t)
	me, other := g.Seats[1], g.Seats[2]
	id := PushCatalogPermanent(g, me.ID, name, typeLine, oracle, false)
	HobA5Draw(t, g, other, nil)
	HobA5Draw(t, g, other, nil)
	if n := HobA5Counters(t, g, id); n != 0 {
		t.Fatalf("%d counters after another player drew twice, want 0", n)
	}
	HobA5Draw(t, g, me, nil)
	if n := HobA5Counters(t, g, id); n != 0 {
		t.Fatalf("%d counters after the first draw, want 0", n)
	}
	HobA5Draw(t, g, me, nil)
	if n := HobA5Counters(t, g, id); n != 1 {
		t.Fatalf("%d counters after the second draw, want 1", n)
	}
	HobA5Draw(t, g, me, nil)
	if n := HobA5Counters(t, g, id); n != 1 {
		t.Fatalf("%d counters after the third draw, want it to stay 1", n)
	}
	return id
}
