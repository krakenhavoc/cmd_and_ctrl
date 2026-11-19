package game

import "testing"

// special_action_cost_increase_test.go — #5844: a special-action cost
// INCREASE (Exiled Doomsayer's "All morph costs cost {2} more") reaches
// a turn-face-up by its route, and gives a cost that prints no mana
// something to pay, the way an activation tax does (#5382).

const (
	zeroMorphOracle      = "test-zero-morph-card"
	morphTaxSourceOracle = "test-morph-tax-source"
)

// morphTax is a CostModifier in Exiled Doomsayer's shape.
func morphTax() CostModifier {
	return CostModifier{
		Kind:           CostIncrease,
		SpecialActions: true,
		Label:          "All morph costs cost {2} more.",
		AppliesTo: func(q CostQuery) bool {
			return q.SpecialAction != nil && q.SpecialAction.Kind == SpecialActionTurnFaceUp &&
				q.SpecialAction.Route == TurnFaceUpByMorph
		},
		Amount: fixed(2),
	}
}

// zeroMorph seats a face-down permanent whose card prints "Morph {0}".
func zeroMorph(t *testing.T, g *Game, me *Player) Card {
	t.Helper()
	withCatalogAlternativeCosts(t, altCostFor(zeroMorphOracle, morphOffer(FaceDownMorphed, "", false)))
	id := faceDownPermanent(t, g, me, "Zero Morph", "Creature — Shapeshifter", FaceDownMorphed)
	c := findBattlefieldCard(g, id)
	c.OracleID = zeroMorphOracle
	return *c
}

func TestAZeroMorphCostIsFreeWithNoTax(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	card := zeroMorph(t, g, me)
	offer := TurnFaceUpOffer(card)
	if offer == nil || offer.Cost != "" {
		t.Fatalf("offer = %+v, want a free morph", offer)
	}
	if g.SpecialActionChargesManaForEffect(me.ID, card, SpecialActionTurnFaceUp, *offer) {
		t.Error("a free morph owes mana with no tax on the board")
	}
	if err := g.PerformSpecialAction(me.ID, card.InstanceID, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn a free morph face up with an empty pool: %v", err)
	}
}

func TestAMorphTaxReachesAMorphCostThatPrintsNoMana(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	withCatalogCostModifiers(t, modifiersFor(morphTaxSourceOracle, morphTax()))
	modifierSource(t, g, me, "Exiled Doomsayer", morphTaxSourceOracle)
	card := zeroMorph(t, g, me)
	offer := TurnFaceUpOffer(card)
	if !g.SpecialActionChargesManaForEffect(me.ID, card, SpecialActionTurnFaceUp, *offer) {
		t.Fatal("a taxed free morph owes no mana")
	}
	cost, err := g.SpecialActionManaCostForEffect(me.ID, card, SpecialActionTurnFaceUp, *offer)
	if err != nil || cost.String() != "{2}" {
		t.Fatalf("price = %q, %v; want {2}", cost.String(), err)
	}
	if err := g.PerformSpecialAction(me.ID, card.InstanceID, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true}); err == nil {
		t.Fatal("turned a taxed free morph face up with an empty pool")
	}
	fillPool(g, me, 2)
	if err := g.PerformSpecialAction(me.ID, card.InstanceID, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up for the {2} tax: %v", err)
	}
	if n := len(me.ManaPool); n != 0 {
		t.Errorf("%d mana left over, want the {2} spent", n)
	}
}

// A tax on the morph route does not reach a free action by another
// route or of another kind: the cost stays empty and nothing is owed.
func TestAMorphTaxLeavesAFreeDisguiseAlone(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	withCatalogCostModifiers(t, modifiersFor(morphTaxSourceOracle, morphTax()))
	modifierSource(t, g, me, "Exiled Doomsayer", morphTaxSourceOracle)
	disguise := morphOffer(FaceDownDisguised, "", false)
	disguise.Key = "disguise"
	withCatalogAlternativeCosts(t, altCostFor(zeroMorphOracle, disguise))
	id := faceDownPermanent(t, g, me, "Zero Disguise", "Creature — Shapeshifter", FaceDownDisguised)
	c := findBattlefieldCard(g, id)
	c.OracleID = zeroMorphOracle
	offer := TurnFaceUpOffer(*c)
	if offer == nil {
		t.Fatal("no disguise offer")
	}
	if g.SpecialActionChargesManaForEffect(me.ID, *c, SpecialActionTurnFaceUp, *offer) {
		t.Error("a morph tax reached a disguise cost")
	}
}
