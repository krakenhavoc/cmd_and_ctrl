package effects

import (
	"testing"
)

// aang_g4_test.go — the "Aang is so flashy" deck's flicker / ETB /
// bounce group (#1306): the cards whose caveats closed and the cards
// that were missing.

// --- Deputy of Acquittals ----------------------------------------

// "Another target creature you control" excludes THIS Deputy by
// instance and nothing else: a second, same-named Deputy is offered
// (and returned when picked), the entering one is not.
func TestDeputyOfAcquittalsOffersAnotherCreatureButNeverItself(t *testing.T) {
	g := newCatalogGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]
	bear := pushCatalogPermanent(g, p0.ID, "Bears", "Creature — Bear", "", false)
	otherDeputy := pushCatalogPermanent(g, p0.ID, "Deputy of Acquittals", "Creature — Human Wizard", deputyOfAcquittalsOracle, false)
	theirs := pushCatalogPermanent(g, p1.ID, "Their Bear", "Creature — Bear", "", false)

	deputy := castAndResolveCreature(t, g, "Deputy of Acquittals", "Creature — Human Wizard", deputyOfAcquittalsOracle)
	answerLatestTriggerPrompt(t, g, p0.ID, true)

	pick := latestPickTarget(g, p0.ID)
	if pick == nil {
		t.Fatal("no pick_target prompt after answering yes")
	}
	if hasID(pick.PickTargetCards, deputy) {
		t.Error("the entering Deputy is offered as its own \"another\" target")
	}
	if !hasID(pick.PickTargetCards, otherDeputy) {
		t.Error("a second Deputy of Acquittals is not offered — \"another\" is not \"not named Deputy\"")
	}
	if !hasID(pick.PickTargetCards, bear) {
		t.Error("an ordinary creature you control is not offered")
	}
	if hasID(pick.PickTargetCards, theirs) {
		t.Error("an opponent's creature is offered to \"a creature you control\"")
	}

	pickCard(t, g, p0.ID, otherDeputy)
	passPriorityAroundTable(t, g)
	if !p0.Hand.Contains(otherDeputy) {
		t.Error("the picked Deputy did not return to its owner's hand")
	}
	if !onBattlefield(g, deputy) {
		t.Error("the entering Deputy left the battlefield")
	}
	if spec, _ := Lookup(deputyOfAcquittalsOracle); spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Deputy of Acquittals is %v with caveats %v, want full", spec.Completeness, spec.Caveats)
	}
}
