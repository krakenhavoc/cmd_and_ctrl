package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const cosmogoyfOracle = "99d8c5f9-ad8f-450e-9a7c-51a4c0ca5043"

func TestCosmogoyfCountsCardsYouOwnInExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	goyf := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Cosmogoyf", TypeLine: "Creature — Elemental Lhurgoyf",
		OracleID: cosmogoyfOracle, Owner: me, Controller: me,
	})
	if p, tt := effectivePower(t, g, goyf), effectiveToughness(t, g, goyf); p != 0 || tt != 1 {
		t.Fatalf("empty exile = %d/%d, want 0/1", p, tt)
	}
	g.WithWriteLock(func() {
		for _, o := range []uuid.UUID{me, me, opp} {
			g.Exile.PushTop(game.Card{InstanceID: uuid.New(), Name: "X", TypeLine: "Instant", Owner: o, Controller: o})
		}
		g.BumpLayerVersionForTest()
	})
	if p, tt := effectivePower(t, g, goyf), effectiveToughness(t, g, goyf); p != 2 || tt != 3 {
		t.Fatalf("two owned exiled = %d/%d, want 2/3 (an opponent's exiled card does not count)", p, tt)
	}
}
