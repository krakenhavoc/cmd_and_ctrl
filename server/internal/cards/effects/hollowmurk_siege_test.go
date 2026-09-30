package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const hollowmurkSiegeOracle = "31061e34-042e-40c3-99ab-752795ab4324"

// TestHollowmurkSiegeSultaiDrawsOncePerTurnOnAnyCounterKindOnYourCreatureOnly
// — placement (not removal), any counter kind, a creature you control
// only, and the once-each-turn limit.
func TestHollowmurkSiegeSultaiDrawsOncePerTurnOnAnyCounterKindOnYourCreatureOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	castSiege(t, g, "Hollowmurk Siege", hollowmurkSiegeOracle, "Sultai")
	mine := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	// An opponent's creature getting a counter draws nothing.
	handBefore := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(theirs, "stun", 1); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 0 {
		t.Errorf("a counter on an opponent's creature drew %d cards, want 0", got)
	}

	// A loyalty-style counter on your own creature draws — any kind.
	handBefore = me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(mine, "stun", 1); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("a counter on your creature drew %d cards, want 1", got)
	}

	// A second placement the same turn draws nothing (once each turn).
	handBefore = me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(mine, "+1/+1", 1); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 0 {
		t.Errorf("a second counter placement the same turn drew %d cards, want 0 (once each turn)", got)
	}

	// Removing a counter is not a placement and draws nothing.
	handBefore = me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(mine, "+1/+1", -1); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 0 {
		t.Errorf("removing a counter drew %d cards, want 0", got)
	}
}

// TestHollowmurkSiegeAbzanPutsOneCounterAndMenaceOnTheTargetedAttackerOncePerCombat
// — "whenever you attack" fires once per combat declaration (not once
// per attacker), and the target gets a +1/+1 counter and menace until
// end of turn.
func TestHollowmurkSiegeAbzanPutsOneCounterAndMenaceOnTheTargetedAttackerOncePerCombat(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	castSiege(t, g, "Hollowmurk Siege", hollowmurkSiegeOracle, "Abzan")
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatalf("first attack: %v", err)
	}
	if err := g.DeclareAttacker(b, opp.ID); err != nil {
		t.Fatalf("second attack: %v", err)
	}
	lockInAttacks(t, g)
	answerPickTarget(t, g, a)
	passPriorityAroundTable(t, g)

	if n := plusOneCounters(g, a); n != 1 {
		t.Errorf("Bear A has %d +1/+1 counters, want 1", n)
	}
	if n := plusOneCounters(g, b); n != 0 {
		t.Errorf("Bear B has %d +1/+1 counters, want 0 (one trigger for the whole attack, not one per attacker)", n)
	}
	if !effectiveAbilitiesContain(t, g, a, "menace") {
		t.Error("the targeted attacker lacks menace")
	}
	if effectiveAbilitiesContain(t, g, b, "menace") {
		t.Error("the untargeted attacker has menace")
	}
}

// TestHollowmurkSiegeModesDoNotLeak — a Sultai Siege has no Abzan
// attack trigger and an Abzan Siege has no Sultai counter trigger.
func TestHollowmurkSiegeModesDoNotLeak(t *testing.T) {
	g := newCatalogGame(t)
	id := castSiege(t, g, "Hollowmurk Siege", hollowmurkSiegeOracle, "Sultai")
	if keysMention(siegeTriggerKeys(g, id), "attacking creature") {
		t.Error("a Sultai Hollowmurk Siege has the Abzan trigger")
	}

	g = newCatalogGame(t)
	id = castSiege(t, g, "Hollowmurk Siege", hollowmurkSiegeOracle, "Abzan")
	if keysMention(siegeTriggerKeys(g, id), "draw a card") {
		t.Error("an Abzan Hollowmurk Siege has the Sultai trigger")
	}
}
