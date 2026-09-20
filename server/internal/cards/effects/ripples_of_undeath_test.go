package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const ripplesOfUndeathOracle = "2acf8e34-9215-4a72-a24f-09d3bdd0083a"

// TestRipplesOfUndeathMillsAtYourFirstMainPhaseOnly is the trigger's
// whole contract: it fires at its controller's precombat main and at
// nobody else's.
func TestRipplesOfUndeathMillsAtYourFirstMainPhaseOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "Ripples of Undeath", ripplesOfUndeathOracle, "Enchantment")

	before := me.Graveyard.Size()
	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)

	// The mill has happened and the optional payment is waiting.
	if got := me.Graveyard.Size() - before; got != 3 {
		t.Fatalf("milled %d cards at your own first main phase, want 3", got)
	}
	ask := latestChoiceOfKind(g, game.PendingChoicePayUnless)
	if ask == nil {
		t.Fatal("the trigger offers the optional payment")
	}
	if ask.PayCost != "{1}" {
		t.Errorf("the mana half of the optional cost is {1}: got %q", ask.PayCost)
	}
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	// An opponent's first main phase is not yours.
	mid := me.Graveyard.Size()
	advanceToPrecombatMainOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if me.Graveyard.Size() != mid {
		t.Errorf("an opponent's first main phase milled %d cards, want 0", me.Graveyard.Size()-mid)
	}
}

// TestRipplesOfUndeathPaymentKeepsOneOfTheMilledCards is the "if you
// do" half: {1} off the pool, 3 life off the total, and exactly one
// of the three milled cards — not any other graveyard card — comes
// back to hand.
func TestRipplesOfUndeathPaymentKeepsOneOfTheMilledCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "Ripples of Undeath", ripplesOfUndeathOracle, "Enchantment")
	// A card already in the graveyard is NOT "among those cards".
	older := pushGraveyardCardForTest(me, "Older Card")

	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)

	ask := latestChoiceOfKind(g, game.PendingChoicePayUnless)
	if ask == nil {
		t.Fatal("the trigger offers the optional payment")
	}
	life, hand := me.Life, me.Hand.Size()
	b06AddMana(me, "B")
	if err := g.ResolvePayUnless(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}

	if me.Life != life-3 {
		t.Errorf("paying costs 3 life: %d -> %d", life, me.Life)
	}
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("paying asks which of the milled cards to keep")
	}
	if len(pick.ChooseCards) != 3 {
		t.Fatalf("offered %d candidates, want the 3 cards milled this way", len(pick.ChooseCards))
	}
	for _, id := range pick.ChooseCards {
		if id == older {
			t.Error("a card already in the graveyard was offered")
		}
	}
	keep := pick.ChooseCards[0]
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{keep}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(keep) {
		t.Error("the chosen card did not reach the hand")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want exactly one card back", hand, me.Hand.Size())
	}
}

// TestRipplesOfUndeathDoesNotOfferAPaymentYouCannotMake is CR 119.4:
// below 3 life the payment is not a payment, so the offer is not
// made — and the mill still happens.
func TestRipplesOfUndeathDoesNotOfferAPaymentYouCannotMake(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "Ripples of Undeath", ripplesOfUndeathOracle, "Enchantment")
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2-me.Life) })

	before := me.Graveyard.Size()
	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)

	if got := me.Graveyard.Size() - before; got != 3 {
		t.Errorf("milled %d cards, want 3 — the mill is not conditional on the payment", got)
	}
	if ask := latestChoiceOfKind(g, game.PendingChoicePayUnless); ask != nil {
		t.Error("a player at 2 life was offered a 3-life payment")
	}
}
