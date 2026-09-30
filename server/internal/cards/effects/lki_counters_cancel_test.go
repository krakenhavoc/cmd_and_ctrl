package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #1630: a departed permanent's counters come from the engine's
// departure record (the live Card.Counters, copied as it leaves), not
// from a walk over EventCounterPlaced. The CR 704.5q state-based action
// removes a +1/+1 and a -1/-1 counter in pairs WITHOUT emitting a
// counter event, so the walk kept reporting counters the rules had
// already removed.

// cancelledPairCreature is a 3/3 catalog creature that has taken two
// +1/+1 counters and one -1/-1 counter; 704.5q has cancelled one pair,
// so it holds ONE +1/+1 counter and no -1/-1 counter.
func cancelledPairCreature(t *testing.T, g *game.Game, owner uuid.UUID, name, oracle string) uuid.UUID {
	t.Helper()
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Artifact Creature — Construct", OracleID: oracle,
		Power: 3, Toughness: 3, Owner: owner, Controller: owner,
	})
	if err := g.AddCounter(id, game.CounterPlusOne, 2); err != nil {
		t.Fatalf("AddCounter +1/+1: %v", err)
	}
	if err := g.AddCounter(id, game.CounterMinusOne, 1); err != nil {
		t.Fatalf("AddCounter -1/-1: %v", err)
	}
	runStateChecksViaDraw(t, g)
	if got := counterCount(g, id, game.CounterPlusOne); got != 1 {
		t.Fatalf("704.5q should leave one +1/+1 counter, got %d", got)
	}
	if got := counterCount(g, id, game.CounterMinusOne); got != 0 {
		t.Fatalf("704.5q should leave no -1/-1 counter, got %d", got)
	}
	return id
}

// Marketback Walker draws for the +1/+1 counters it HAD when it died: one
// after the cancel, where the log walk said two. Rules-side change: the
// draw follows the counters the rules left on it.
func TestMarketbackWalkerDrawsForCountersLeftAfterACancelledPair(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	walker := cancelledPairCreature(t, g, me.ID, "Marketback Walker", b39MarketbackWalkerOracle)

	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(walker) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1 (one +1/+1 counter survived 704.5q)", hand, me.Hand.Size())
	}
}

// Blowfly Infestation triggers on a creature that died WITH a -1/-1
// counter. One that had it cancelled against a +1/+1 counter did not.
func TestBlowflyInfestationIgnoresAMinusCounterThatWasCancelled(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b29Push(g, me.ID, "Blowfly Infestation", "Enchantment", b29BlowflyInfestationOracle, "{2}{B}", 0, 0, "B")
	pushCounterCreature(g, opp.ID, "Bystander", "", 0)
	cancelled := cancelledPairCreature(t, g, opp.ID, "Cancelled", "")

	b27Kill(g, cancelled)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("the -1/-1 counter was cancelled before it died: no trigger")
	}
}

// Twitching Doll makes a Spider for each counter on it as it was
// sacrificed, every kind counted, read off its departure record.
func TestTwitchingDollCountsTheCountersItHadWhenSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	doll := cancelledPairCreature(t, g, me.ID, "Twitching Doll", "fd6e1967-237a-41f6-bbf4-2c869f9447c8")
	if err := g.AddCounter(doll, "nest", 2); err != nil {
		t.Fatalf("AddCounter nest: %v", err)
	}
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, doll, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := b29CountNamed(g, me.ID, "Spider"); got != 3 {
		t.Errorf("Spiders = %d, want 3 (one +1/+1 counter that survived 704.5q, two nest)", got)
	}
}
