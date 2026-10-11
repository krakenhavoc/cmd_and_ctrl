package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_independence_draws_test.go — ADR 0018's #3013 amendment,
// sub-PR 4, batch 5: draw triggers declare DrawCards. A draw takes the
// top of its controller's library, so it conflicts with any other item
// that touches that library (scry, surveil, mill, another draw), and a
// "whenever you draw" watcher keeps the batch asking.

const (
	llanowarVisionaryOracle = "f75ed312-3a23-4624-80c5-03980aa22d0b"
	chasmSkulkerOracle      = "69facbc7-3859-4716-b627-5199571fb3cf"
)

func TestTheDrawTriggersDeclareFootprints(t *testing.T) {
	for name, c := range map[string]struct {
		oracle string
		n      int
	}{
		"Llanowar Visionary":   {llanowarVisionaryOracle, 1},
		"Frostbite Pyromental": {"9a104f73-597b-49e3-8088-13db31f61900", 2},
	} {
		rows := game.CatalogTriggers(c.oracle)
		if len(rows) == 0 || len(rows[0].Footprint) != 1 ||
			rows[0].Footprint[0].Kind != game.FootprintDraw || rows[0].Footprint[0].N != c.n {
			t.Errorf("%s: rows %+v, want a draw step of %d first", name, rows, c.n)
		}
	}
}

// enterTwo puts a draw maker and a second entering permanent on
// the battlefield as one batch and returns the order prompt, if any.
func enterTwo(t *testing.T, g *game.Game, nameA, typeA, oracleA, nameB, typeB, oracleB string) *game.PendingChoice {
	t.Helper()
	me := g.Seats[0]
	toMain(t, g)
	a := pushCatalogPermanent(g, me.ID, nameA, typeA, oracleA, false)
	b := pushCatalogPermanent(g, me.ID, nameB, typeB, oracleB, false)
	enterTogether(g, me.ID, a, b)
	return settleUntilOrderPrompt(t, g)
}

// A draw and a drain touch different things, and nothing watches a
// draw: no prompt, and the card is drawn.
func TestADrawAndADrainOnOneEntryDoNotAsk(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	if ch := enterTwo(t, g, "Llanowar Visionary", "Creature — Elf Druid Shaman", llanowarVisionaryOracle,
		"Cauldron Familiar", "Creature — Bat", cauldronFamiliarOracle); ch != nil {
		t.Fatalf("a draw and a drain asked for an order: %+v", ch)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}
}

// A draw beside a scry reads and reorders the same library top.
func TestADrawAndAScryStillAsk(t *testing.T) {
	g := newCatalogGame(t)
	before := promptsByReason(t, game.TriggerOrderReasonLibrary)
	if ch := enterTwo(t, g, "Llanowar Visionary", "Creature — Elf Druid Shaman", llanowarVisionaryOracle,
		"Faerie Seer", "Creature — Faerie Wizard", faerieSeerOracle); ch == nil {
		t.Fatal("a draw and a scry touch one library and must keep the prompt")
	}
	if got := promptsByReason(t, game.TriggerOrderReasonLibrary) - before; got != 1 {
		t.Errorf("counted %v under library, want 1", got)
	}
}

// A "whenever you draw" watcher sees the draw, so the order shows.
func TestADrawAndADrainStillAskBesideADrawWatcher(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Chasm Skulker", "Creature — Squid Horror", chasmSkulkerOracle, false)
	before := promptsByReason(t, game.TriggerOrderReasonReactor)
	if ch := enterTwo(t, g, "Llanowar Visionary", "Creature — Elf Druid Shaman", llanowarVisionaryOracle,
		"Cauldron Familiar", "Creature — Bat", cauldronFamiliarOracle); ch == nil {
		t.Fatal("a draw watcher on the board must keep the prompt")
	}
	if got := promptsByReason(t, game.TriggerOrderReasonReactor) - before; got != 1 {
		t.Errorf("counted %v under reactor, want 1", got)
	}
}
