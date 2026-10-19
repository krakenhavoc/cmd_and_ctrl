package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rules_batch_4082_test.go — #4082 (an ability whose only target went
// away does nothing at all, life gain included, CR 608.2b) and #4095
// (destroyEachLegalTarget destroys its targets as ONE event).

// sorinPlusDamageGain runs one "deals N damage to any target and you
// gain N life" walker ability's Effect against a creature that either
// stays or is already gone, and returns the controller's life change.
//
// The Effect is called directly: the stack's own "every target illegal"
// short-circuit (see Context.IsTargetLegal) already fizzles the item
// before the Effect runs in real play, so the Effect's own guard is a
// second line that only a direct call reaches (#4082).
func sorinPlusDamageGain(t *testing.T, name, oracle string, loyalty, idx int, leaves bool) int {
	t.Helper()
	g, _, me, sorin := sorinATable(t, name, oracle, loyalty)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	target := bear
	if leaves {
		// Not in any zone at all: an item carrying no announced clause is
		// re-checked for bare existence, and a graveyard card exists.
		target = uuid.New()
	}
	spec, ok := Lookup(oracle)
	if !ok {
		t.Fatalf("%s is not in the catalog", name)
	}
	item := &game.StackItem{
		Controller:   me.ID,
		SourceCardID: sorin,
		Targets:      b16TargetCard(target),
	}
	before := me.Life
	var err error
	g.WithWriteLock(func() { err = spec.Activated[idx].Effect(g, item) })
	if err != nil {
		t.Fatalf("effect: %v", err)
	}
	return me.Life - before
}

func TestSorinMarkovPlusTwoGainsNothingWhenItsTargetLeft(t *testing.T) {
	const o = "c151e8b7-4b28-4f03-8a81-9bf623f893d7"
	if got := sorinPlusDamageGain(t, "Sorin Markov", o, 4, 0, false); got != 2 {
		t.Errorf("target stayed: gained %d, want 2", got)
	}
	if got := sorinPlusDamageGain(t, "Sorin Markov", o, 4, 0, true); got != 0 {
		t.Errorf("target left in response: gained %d, want 0 (CR 608.2b)", got)
	}
}

func TestSorinMirthlessMinusSevenGainsNothingWhenItsTargetLeft(t *testing.T) {
	if got := sorinPlusDamageGain(t, "Sorin the Mirthless", sorinMirthlessOracle, 7, 2, false); got != 13 {
		t.Errorf("target stayed: gained %d, want 13", got)
	}
	if got := sorinPlusDamageGain(t, "Sorin the Mirthless", sorinMirthlessOracle, 7, 2, true); got != 0 {
		t.Errorf("target left in response: gained %d, want 0 (CR 608.2b)", got)
	}
}

func TestSorinVampireLordMinusTwoGainsNothingWhenItsTargetLeft(t *testing.T) {
	if got := sorinPlusDamageGain(t, "Sorin, Vampire Lord", sorinVampireLordOracle, 4, 1, false); got != 4 {
		t.Errorf("target stayed: gained %d, want 4", got)
	}
	if got := sorinPlusDamageGain(t, "Sorin, Vampire Lord", sorinVampireLordOracle, 4, 1, true); got != 0 {
		t.Errorf("target left in response: gained %d, want 0 (CR 608.2b)", got)
	}
}

// Two targets destroyed by one spell die together: a Zulaport Cutthroat
// among them sees BOTH deaths (it watches "another creature you control"
// too), so the opponent loses 2, not the 1 a one-at-a-time loop gave.
func TestCurtainsCallDestroysBothTargetsAsOneEvent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	z := pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue Ally", zulaportCutthroatOracle, false)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	before := opp.Life
	castCatalogSpell(t, g, "Curtains' Call", "Instant", curtainsCallOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: z}, {Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(z) || g.Battlefield.Contains(bear) {
		t.Fatal("both targets should be destroyed")
	}
	if got := before - opp.Life; got != 2 {
		t.Errorf("Zulaport saw %d deaths, want 2 (simultaneous, CR 700.4)", got)
	}
}

func TestWearDownPromisedDestroysBothTargetsAsOneEvent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	z := pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Artifact Creature — Human Rogue Ally", zulaportCutthroatOracle, false)
	golem := b12Permanent(g, me.ID, "Golem", "Artifact Creature — Golem")
	before := opp.Life
	mustCastWithGift(t, g, "Wear Down", "Sorcery", giftWearDownOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: z}, {Kind: game.TargetCard, ID: golem}}, opp.ID)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(z) || g.Battlefield.Contains(golem) {
		t.Fatal("both targets should be destroyed")
	}
	if got := before - opp.Life; got != 2 {
		t.Errorf("Zulaport saw %d deaths, want 2 (simultaneous, CR 700.4)", got)
	}
}
