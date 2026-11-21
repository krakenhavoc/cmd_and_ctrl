package testsbr

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

const captainHowlerOracle = "46c99d67-8307-4af0-afd0-377714806614"

// TestCaptainHowlerSeaScourgeTwoCardsAreOneTriggerAndPumpTwiceAsMuch
// pins the batch reading (#3172): discarding two cards at once is ONE
// trigger (one target prompt), and the pump is +2/+0 for each card
// discarded — +4/+0 — as printed.
func TestCaptainHowlerSeaScourgeTwoCardsAreOneTriggerAndPumpTwiceAsMuch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	howler := effectstest.PushDiesCreatureForTest(g, me.ID, "Captain Howler, Sea Scourge", captainHowlerOracle,
		"Legendary Creature — Shark Pirate", 5, 4)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)

	if me.Hand.Size() < 2 {
		t.Fatalf("need two cards in hand to discard: have %d", me.Hand.Size())
	}
	discarded := []uuid.UUID{me.Hand.Cards[0].InstanceID, me.Hand.Cards[1].InstanceID}
	g.WithWriteLock(func() { g.DiscardChoiceForEffect(me.ID, 2) })
	effectstest.AnswerDiscard(t, g, me.ID, discarded...)

	effectstest.PickCard(t, g, me.ID, bear)
	if got := effectstest.TriggersOnStackFrom(g, howler); got != 1 {
		t.Fatalf("a two-card discard must fire the trigger once, not once per card: got %d", got)
	}

	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 6 {
		t.Errorf("two discarded cards are +4/+0: power %d, want 6", p)
	}
}

// One card is the exact +2/+0.
func TestCaptainHowlerSeaScourgeOneCardIsExact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushDiesCreatureForTest(g, me.ID, "Captain Howler, Sea Scourge", captainHowlerOracle,
		"Legendary Creature — Shark Pirate", 5, 4)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)

	g.WithWriteLock(func() { g.DiscardChoiceForEffect(me.ID, 1) })
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("one discarded card is +2/+0 exactly: power %d, want 4", p)
	}
}

// "Whenever that creature deals combat damage to a player this turn,
// you draw a card": the pumped creature's hit draws, another creature's
// does not.
func TestCaptainHowlerSeaScourgeDrawsWhenTheChosenCreatureConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushDiesCreatureForTest(g, me.ID, "Captain Howler, Sea Scourge", captainHowlerOracle,
		"Legendary Creature — Shark Pirate", 5, 4)
	chosen := effectstest.PushVanillaCreature(g, me.ID, "Chosen Bear", 2, 2)
	other := effectstest.PushVanillaCreature(g, me.ID, "Other Bear", 2, 2)

	g.WithWriteLock(func() { g.DiscardChoiceForEffect(me.ID, 1) })
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, chosen)
	effectstest.PassPriorityAroundTable(t, g)

	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Fatalf("an unchosen creature connected: hand %d, want %d", got, hand)
	}
}

func TestCaptainHowlerSeaScourgeChosenCreatureHitDrawsACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushDiesCreatureForTest(g, me.ID, "Captain Howler, Sea Scourge", captainHowlerOracle,
		"Legendary Creature — Shark Pirate", 5, 4)
	chosen := effectstest.PushVanillaCreature(g, me.ID, "Chosen Bear", 2, 2)

	g.WithWriteLock(func() { g.DiscardChoiceForEffect(me.ID, 1) })
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, chosen)
	effectstest.PassPriorityAroundTable(t, g)

	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, chosen)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("the chosen creature connected: hand %d, want %d", got, hand+1)
	}
}

// Compound ward closes the final printed-rule gap (#2945).
func TestCaptainHowlerSeaScourgeIsComplete(t *testing.T) {
	spec, ok := effects.Lookup(captainHowlerOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("want full implementation without caveats: %v %v", spec.Completeness, spec.Caveats)
	}
}
