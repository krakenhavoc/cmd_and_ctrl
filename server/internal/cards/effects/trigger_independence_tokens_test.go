package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_independence_tokens_test.go — ADR 0018's #3013 amendment,
// sub-PR 4, batch 4: token makers declare CreateToken. A token entering
// is an event (token created, entered, zone move), so a batch with a
// token maker goes on with no prompt only when nothing at the table
// watches an entering object.

const (
	openTheGravesOracle   = "28778958-a1f9-4fea-b551-c193d1257f18"
	beetlebackChiefOracle = "f5c5f64c-6911-430c-a825-b32b96d39c7d"
)

func TestTheTokenMakersDeclareFootprints(t *testing.T) {
	for name, c := range map[string]struct {
		oracle string
		n      int
	}{
		"Open the Graves":  {openTheGravesOracle, 1},
		"Beetleback Chief": {beetlebackChiefOracle, 2},
	} {
		rows := game.CatalogTriggers(c.oracle)
		if len(rows) != 1 || len(rows[0].Footprint) != 1 ||
			rows[0].Footprint[0].Kind != game.FootprintCreateToken || rows[0].Footprint[0].N != c.n {
			t.Errorf("%s: rows %+v, want one token step of %d", name, rows, c.n)
		}
	}
}

// zombieTokens counts the Zombie tokens on the battlefield.
func zombieTokens(g *game.Game) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.HasSubtype("Zombie") {
			n++
		}
	}
	return n
}

// Open the Graves and Zulaport Cutthroat on one nontoken creature
// dying: a Zombie and a drain. The order cannot show and nothing
// watches an entering object, so there is no prompt.
func TestATokenMakerAndADrainOnOneDeathDoNotAsk(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Open the Graves", "Enchantment", openTheGravesOracle, false)
	pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue Ally", zulaportOracle, false)
	bear := b21Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	toMain(t, g)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if ch := settleUntilOrderPrompt(t, g); ch != nil {
		t.Fatalf("a token maker and a drain asked for an order: %+v", ch)
	}
	if got := zombieTokens(g); got != 1 {
		t.Errorf("%d Zombie tokens, want 1", got)
	}
}

// The same board with a watcher of entering creatures (Impact Tremors)
// asks: the order decides whether the Zombie's entry sees the drain's
// life change first.
func TestATokenMakerAndADrainStillAskBesideAnEntryWatcher(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Open the Graves", "Enchantment", openTheGravesOracle, false)
	pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue Ally", zulaportOracle, false)
	pushCatalogPermanent(g, me.ID, "Impact Tremors", "Enchantment", impactTremorsOracle, false)
	bear := b21Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	toMain(t, g)

	before := promptsByReason(t, game.TriggerOrderReasonReactor)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if ch := settleUntilOrderPrompt(t, g); ch == nil {
		t.Fatal("an entry watcher on the board must keep the prompt")
	}
	if got := promptsByReason(t, game.TriggerOrderReasonReactor) - before; got != 1 {
		t.Errorf("counted %v under reactor, want 1", got)
	}
}
