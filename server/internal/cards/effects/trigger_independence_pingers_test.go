package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_independence_pingers_test.go — ADR 0018's #3013 amendment,
// sub-PR 4, batch 3: "deals N damage to each opponent" triggers declare
// DamageEachOpponent, so two of them off one entry go on with no
// prompt.

func TestThePingersDeclareFootprints(t *testing.T) {
	for name, c := range map[string]struct {
		oracle string
		n      int
	}{
		"Impact Tremors":       {impactTremorsOracle, 1},
		"Molten Gatekeeper":    {moltenGatekeeperOracle, 1},
		"Black Waltz No. 3":    {"9c84c681-40c1-4ff3-86e6-48b4ab7491e5", 2},
		"Purphoros, God/Forge": {"4fdbbec2-e921-4b63-958d-f9ba1e417197", 2},
	} {
		rows := game.CatalogTriggers(c.oracle)
		if len(rows) != 1 || len(rows[0].Footprint) != 1 ||
			rows[0].Footprint[0].Kind != game.FootprintDamageEachOpponent || rows[0].Footprint[0].N != c.n {
			t.Errorf("%s: rows %+v, want one damage-to-each-opponent step of %d", name, rows, c.n)
		}
	}
}

func TestTwoPingersOnOneEntryDoNotAsk(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Impact Tremors", "Enchantment", impactTremorsOracle, false)
	pushCatalogPermanent(g, me.ID, "Molten Gatekeeper", "Creature — Elemental", moltenGatekeeperOracle, false)
	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = p.Life
	}

	b20CastCreature(t, g, me, "Walking Corpse", "Creature — Zombie", "", 2, 2)
	if ch := settleUntilOrderPrompt(t, g); ch != nil {
		t.Fatalf("two pingers asked for an order: %+v", ch)
	}
	for i, p := range g.Seats {
		want := before[i]
		if p.ID != me.ID {
			want -= 2
		}
		if p.Life != want {
			t.Errorf("seat %d life %d → %d, want %d", i, before[i], p.Life, want)
		}
	}
}

func TestTwoPingersStillAskBesideALifeWatcher(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Impact Tremors", "Enchantment", impactTremorsOracle, false)
	pushCatalogPermanent(g, me.ID, "Molten Gatekeeper", "Creature — Elemental", moltenGatekeeperOracle, false)
	pushCatalogPermanent(g, opp.ID, "Ajani's Pridemate", "Creature — Cat Soldier", b24AjanisPridemateOracle, false)
	before := promptsByReason(t, game.TriggerOrderReasonReactor)
	b20CastCreature(t, g, me, "Walking Corpse", "Creature — Zombie", "", 2, 2)
	if ch := settleUntilOrderPrompt(t, g); ch == nil {
		t.Fatal("a life watcher on the board must keep the prompt")
	}
	if got := promptsByReason(t, game.TriggerOrderReasonReactor) - before; got != 1 {
		t.Errorf("counted %v under reactor, want 1", got)
	}
}
