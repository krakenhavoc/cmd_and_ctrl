package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_independence_counters_test.go — ADR 0018's #3013 amendment,
// sub-PR 4, batch 2: the "put a +1/+1 counter on it" creatures declare
// CounterOnThis, so two of them growing off one entry go on with no
// prompt.

const (
	valleyMightcallerOracle = "16e9c452-6288-4da8-813d-2eb6b7a538c3"
)

func TestTheCounterGrowersDeclareFootprints(t *testing.T) {
	want := game.FootprintStep{Kind: game.FootprintCounterOnSource, N: 1, Counter: game.CounterPlusOne}
	for name, oracle := range map[string]string{
		"Champion of Lambholt": championOfLambholtOracle,
		"Valley Mightcaller":   valleyMightcallerOracle,
		"Ajani's Pridemate":    b24AjanisPridemateOracle,
	} {
		rows := game.CatalogTriggers(oracle)
		if len(rows) != 1 || len(rows[0].Footprint) != 1 || rows[0].Footprint[0].Kind != want.Kind ||
			rows[0].Footprint[0].N != want.N || rows[0].Footprint[0].Counter != want.Counter {
			t.Errorf("%s: rows %+v, want one row with %+v", name, rows, want)
		}
	}
}

func TestTwoGrowersOnOneEntryDoNotAsk(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	champ := pushCatalogPermanent(g, me.ID, "Champion of Lambholt", "Creature — Human Warrior", championOfLambholtOracle, false)
	valley := pushCatalogPermanent(g, me.ID, "Valley Mightcaller", "Creature — Frog Warrior", valleyMightcallerOracle, false)

	b20CastCreature(t, g, me, "Pond Frog", "Creature — Frog", "", 1, 1)
	if ch := settleUntilOrderPrompt(t, g); ch != nil {
		t.Fatalf("two growers asked for an order: %+v", ch)
	}
	for name, id := range map[string]uuid.UUID{"Champion of Lambholt": champ, "Valley Mightcaller": valley} {
		if got := counterCount(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("%s has %d counters, want 1", name, got)
		}
	}
}
