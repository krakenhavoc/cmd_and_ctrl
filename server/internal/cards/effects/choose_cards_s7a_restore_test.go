package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// choose_cards_s7a_restore_test.go — #4361 slice S7a. Three engine
// choose-cards prompts whose continuations were closures are registered
// keys now: CR 502.3's untap determination (Winter Orb, Static Orb),
// Scheming Fence's as-enters choice and Gemstone Caverns' exile from hand.
// Each is captured with its prompt open, restored strictly through JSON,
// and answered in the restored game.

// A paused Winter Orb untap step is a restore point, and the restored
// prompt keeps its plan: the creature no cap counts still untaps, the
// chosen land untaps, the other stays tapped and the step ends.
func TestS7aWinterOrbUntapChoiceSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	forest := pushTappedForTest(g, owner.ID, "Forest", "", "Basic Land — Forest")
	island := pushTappedForTest(g, owner.ID, "Island", "", "Basic Land — Island")
	bear := pushTappedForTest(g, owner.ID, "Grizzly Bears", "", "Creature — Bear")
	pushPermanentForTest(g, owner.ID, "Winter Orb", oracleWinterOrb, "Artifact")
	choice := advanceToUntapChoiceOf(t, g, 1)

	r := restoreThroughJSON(t, g)
	if err := r.ResolveUntapChoice(choice.ID, owner.ID, []uuid.UUID{forest}); err != nil {
		t.Fatalf("answer the restored Winter Orb prompt: %v", err)
	}
	if tappedForTest(t, r, forest) {
		t.Error("the chosen land stayed tapped")
	}
	if !tappedForTest(t, r, island) {
		t.Error("the other land untapped under Winter Orb")
	}
	if tappedForTest(t, r, bear) {
		t.Error("the creature no cap counts stayed tapped")
	}
	if r.Turn.Step == game.StepUntap {
		t.Error("the restored untap step did not end")
	}
}

// The set rule rides the restored prompt as data: under Winter Orb and
// Static Orb two lands are still refused and a land and a creature are
// still accepted.
func TestS7aComposedOrbsKeepTheirSetRuleAcrossARestore(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	forest := pushTappedForTest(g, owner.ID, "Forest", "", "Basic Land — Forest")
	island := pushTappedForTest(g, owner.ID, "Island", "", "Basic Land — Island")
	bear := pushTappedForTest(g, owner.ID, "Grizzly Bears", "", "Creature — Bear")
	pushTappedForTest(g, owner.ID, "Ornithopter", "", "Artifact Creature — Thopter")
	pushPermanentForTest(g, owner.ID, "Winter Orb", oracleWinterOrb, "Artifact")
	pushPermanentForTest(g, owner.ID, "Static Orb", oracleStaticOrb, "Artifact")
	choice := advanceToUntapChoiceOf(t, g, 1)

	r := restoreThroughJSON(t, g)
	if err := r.ResolveUntapChoice(choice.ID, owner.ID, []uuid.UUID{forest, island}); err == nil {
		t.Fatal("two lands under Winter Orb was accepted after a restore")
	}
	if err := r.ResolveUntapChoice(choice.ID, owner.ID, []uuid.UUID{forest, bear}); err != nil {
		t.Fatalf("one land and one creature was refused after a restore: %v", err)
	}
	if tappedForTest(t, r, forest) || tappedForTest(t, r, bear) {
		t.Error("the chosen permanents stayed tapped")
	}
	if !tappedForTest(t, r, island) {
		t.Error("the land not chosen untapped under Static Orb")
	}
}

// Scheming Fence's as-enters choice is a restore point; the restored
// answer is stored on the Fence.
func TestS7aSchemingFenceChoiceSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ring := gaPut(g, opp.ID, "Sol Ring", "Artifact", b15SolRingOracle)
	fence := enterFromHand(t, g, me.ID, "Scheming Fence", "Creature — Human Citizen", slFenceOracle)

	r := restoreThroughJSON(t, g)
	slChoose(t, r, me.ID, ring)
	if got := gaLayered(t, r, fence).ChosenPermanent.ID; got != ring {
		t.Fatalf("chosen = %s, want the Sol Ring", got)
	}
}

// A candidate that left and came back before the restored answer is a
// new object (CR 400.7), so it is refused, and nothing is chosen.
func TestS7aSchemingFenceRefusesANewObjectAfterARestore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ring := gaPut(g, opp.ID, "Sol Ring", "Artifact", b15SolRingOracle)
	fence := enterFromHand(t, g, me.ID, "Scheming Fence", "Creature — Human Citizen", slFenceOracle)

	r := restoreThroughJSON(t, g)
	r.WithWriteLock(func() {
		for i := range r.Battlefield.Cards {
			if r.Battlefield.Cards[i].InstanceID == ring {
				r.Battlefield.Cards[i].ObjectEpoch++
			}
		}
	})
	for _, c := range r.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == me.ID {
			_ = r.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{ring})
		}
	}
	if got := gaLayered(t, r, fence).ChosenPermanent.ID; got != uuid.Nil {
		t.Errorf("chosen = %s, want nothing for a new object", got)
	}
}

// Gemstone Caverns' "exile a card from your hand" is a restore point;
// the restored pick exiles the card.
func TestS7aGemstoneCavernsExileSurvivesARestore(t *testing.T) {
	g := openingTable(t)
	ids := handWith(g, 3, gemstoneCaverns(), fillerCard("Keep"), fillerCard("Pitch"))
	keepAll(t, g)
	offers := choicesOf(g, game.PendingChoiceConfirm)
	if len(offers) != 1 {
		t.Fatalf("offers = %+v, want one", offers)
	}
	if err := g.ResolveConfirm(offers[0].ID, g.Seats[3].ID, true); err != nil {
		t.Fatalf("accept: %v", err)
	}
	picks := choicesOf(g, game.PendingChoiceChooseCards)
	if len(picks) != 1 {
		t.Fatalf("exile prompts = %+v, want one", picks)
	}

	r := restoreThroughJSON(t, g)
	if err := r.ResolveChooseCards(picks[0].ID, r.Seats[3].ID, nil); err == nil {
		t.Error("exiling nothing was accepted after a restore")
	}
	if err := r.ResolveChooseCards(picks[0].ID, r.Seats[3].ID, []uuid.UUID{ids[2]}); err != nil {
		t.Fatalf("exile pick on the restored game: %v", err)
	}
	if !r.Exile.Contains(ids[2]) || r.Seats[3].Hand.Contains(ids[2]) {
		t.Error("the picked card is not in exile")
	}
	if !r.Seats[3].Hand.Contains(ids[1]) {
		t.Error("the card that was not picked left the hand")
	}
}
