package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// terrasymbiosis_test.go — #3929. "Whenever you put one or more +1/+1
// counters on a creature you control, you may draw that many cards. Do
// this only once each turn."

const terrasymbiosisOracle = "4ac94586-3ac7-4ade-af31-ad0d0e3ffc4a"

func terrasymbiosisTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Terrasymbiosis", "Enchantment", terrasymbiosisOracle, false)
	bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	return g, me, opp, bear
}

// tsPut puts n counters of kind on id, placed by `by`.
func tsPut(t *testing.T, g *game.Game, by, id uuid.UUID, kind string, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(by, id, kind, n); err != nil {
			t.Fatalf("AddCounterByForEffect: %v", err)
		}
	})
}

func TestTerrasymbiosisDrawsThatManyOnceEachTurn(t *testing.T) {
	g, me, _, bear := terrasymbiosisTable(t)
	hand := me.Hand.Size()
	tsPut(t, g, me.ID, bear, game.CounterPlusOne, 3)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Fatalf("drew %d, want 3 (that many)", got)
	}

	hand = me.Hand.Size()
	tsPut(t, g, me.ID, bear, game.CounterPlusOne, 2)
	passPriorityAroundTable(t, g)
	if owedMays(g, me.ID) > 0 {
		t.Fatal("asked again after the turn's draw was used")
	}
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d more this turn, want 0", got)
	}
}

func TestTerrasymbiosisDecliningKeepsTheTurn(t *testing.T) {
	g, me, _, bear := terrasymbiosisTable(t)
	tsPut(t, g, me.ID, bear, game.CounterPlusOne, 1)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	// A new stack resolution starts a new batch, so this placement is a
	// second occurrence.
	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	hand := me.Hand.Size()
	tsPut(t, g, me.ID, bear, game.CounterPlusOne, 2)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d after a declined offer, want 2", got)
	}
}

// TestTerrasymbiosisCountsAfterReplacements: "that many" is what landed,
// so Hardened Scales' extra counter is drawn for.
func TestTerrasymbiosisCountsAfterReplacements(t *testing.T) {
	g, me, _, bear := terrasymbiosisTable(t)
	pushCatalogPermanent(g, me.ID, "Hardened Scales", "Enchantment", hardenedScalesOracle, false)
	hand := me.Hand.Size()
	tsPut(t, g, me.ID, bear, game.CounterPlusOne, 1)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Fatalf("Hardened Scales: %d counters, want 2", got)
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

func TestTerrasymbiosisIgnoresOtherPlacements(t *testing.T) {
	cases := map[string]func(g *game.Game, me, opp *game.Player, bear uuid.UUID){
		"an opponent puts them on your creature": func(g *game.Game, _, opp *game.Player, bear uuid.UUID) {
			tsPut(t, g, opp.ID, bear, game.CounterPlusOne, 2)
		},
		"another kind of counter": func(g *game.Game, me, _ *game.Player, bear uuid.UUID) {
			tsPut(t, g, me.ID, bear, "charge", 2)
		},
		"a -1/-1 counter": func(g *game.Game, me, _ *game.Player, bear uuid.UUID) {
			tsPut(t, g, me.ID, bear, game.CounterMinusOne, 1)
		},
		"on a creature you don't control": func(g *game.Game, me, opp *game.Player, _ uuid.UUID) {
			theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
			tsPut(t, g, me.ID, theirs, game.CounterPlusOne, 2)
		},
	}
	for name, put := range cases {
		t.Run(name, func(t *testing.T) {
			g, me, opp, bear := terrasymbiosisTable(t)
			put(g, me, opp, bear)
			passPriorityAroundTable(t, g)
			if owedMays(g, me.ID) > 0 {
				t.Error("Terrasymbiosis asked")
			}
		})
	}
}

// TestTerrasymbiosisOncePerCreature: one placement on each of two
// creatures is two triggers (CR 603.2c), and only the first "yes" draws.
func TestTerrasymbiosisOncePerCreature(t *testing.T) {
	g, me, _, bear := terrasymbiosisTable(t)
	other := pushVanillaCreature(g, me.ID, "My Other Bear", 2, 2)
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{bear, other} {
			if err := g.AddCounterByForEffect(me.ID, id, game.CounterPlusOne, 2); err != nil {
				t.Fatalf("AddCounterByForEffect: %v", err)
			}
		}
	})
	if n := owedMays(g, me.ID); n != 2 {
		t.Fatalf("%d triggers, want one per creature", n)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	for owedMays(g, me.ID) > 0 {
		answerLatestTriggerPrompt(t, g, me.ID, true)
		passPriorityAroundTable(t, g)
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2 (one of the two triggers)", got)
	}
}
