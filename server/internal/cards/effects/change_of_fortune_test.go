package effects

import "testing"

// change_of_fortune_test.go — #1112. "For each card you've
// discarded this turn" has to count a discard from EARLIER this
// turn too, not just the cards this spell itself pitches — the
// whole reason PlayerTurnTally.CardsDiscarded exists (#586's rule:
// "this turn" is read off the tally, never a walk of g.Events).

const changeOfFortuneOracle = "925feb7e-7876-430b-b95a-92b236a2e409"

func TestChangeOfFortuneDrawsForEveryCardDiscardedThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)

	// An earlier discard this turn, from something other than the
	// spell under test — pins that the tally is cumulative for the
	// turn, not just what Change of Fortune itself throws away.
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	if got := g.TurnTallyFor(me.ID).CardsDiscarded; got != 1 {
		t.Fatalf("setup: tally should show 1 discarded so far, got %d", got)
	}

	// Captured BEFORE castCatalogSpell pushes the card into hand and
	// casts it — by the time the spell resolves, Change of Fortune
	// itself is on the stack, not in the hand it discards, so this is
	// exactly how many cards the hand-discard bullet pitches.
	beforeHandSize := me.Hand.Size()

	castCatalogSpell(t, g, "Change of Fortune", "Sorcery", changeOfFortuneOracle, nil)
	passPriorityAroundTable(t, g)

	want := beforeHandSize + 1 // this spell's own discard, plus the earlier one this turn
	if got := me.Hand.Size(); got != want {
		t.Errorf("hand after discard-then-draw: %d, want %d", got, want)
	}
	if got := g.TurnTallyFor(me.ID).CardsDiscarded; got != beforeHandSize+1 {
		t.Errorf("the tally should now count every card discarded this turn: %d, want %d",
			got, beforeHandSize+1)
	}
}

func TestChangeOfFortuneDiscardingAnEmptyHandIsALegalNoOp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.Hand.Cards = nil })

	castCatalogSpell(t, g, "Change of Fortune", "Sorcery", changeOfFortuneOracle, nil)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != 0 {
		t.Errorf("nothing was discarded and nothing this turn before it, so nothing is drawn: hand %d, want 0", got)
	}
}
