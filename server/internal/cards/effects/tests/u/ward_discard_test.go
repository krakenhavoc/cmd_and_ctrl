package testsu

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ward_discard_test.go — "Ward—Discard a card" (#3172): Graveyard
// Trespasser, Graveyard Glutton, Gideon the Oathless.
//
// ward_test.go pins what every ward shares and ward_life_sacrifice_test.go
// the other non-mana costs. These pin what a discard cost changes: the
// payer picks the card from their own hand, an empty hand cannot pay and
// is countered with no prompt, and paying actually puts the card in the
// graveyard before the spell goes on to resolve.

var wardDiscardCards = []struct {
	name, typeLine, oracle string
}{
	{"Graveyard Trespasser", "Creature — Human Werewolf", "0bbd6cad-9b6f-45a6-9f2e-d7b4853586ae"},
	{"Graveyard Glutton", "Creature — Werewolf", "0bbd6cad-9b6f-45a6-9f2e-d7b4853586ae#1"},
	{"Gideon the Oathless", "Legendary Creature — Human Mercenary", "7afa530f-ea51-4bee-af3b-59b83ea5cb25"},
}

// Declining counters the spell and discards nothing.
func TestDiscardWardDeclinedCountersTheSpell(t *testing.T) {
	for _, w := range wardDiscardCards {
		t.Run(w.name, func(t *testing.T) {
			g, _, opp, warded := effectstest.WardTable(t, w.name, w.typeLine, w.oracle)
			hand := effectstest.FillHand(opp, "Filler A", "Filler B")
			bolt := effectstest.BoltWardedCreature(t, g, opp, warded)

			prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
			if prompt == nil {
				t.Fatal("no ward prompt for the spell's controller")
			}
			if err := g.ResolveConfirm(prompt.ID, opp.ID, false); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)

			if !opp.Graveyard.Contains(bolt) {
				t.Error("the countered spell should be in its owner's graveyard")
			}
			for _, id := range hand {
				if !opp.Hand.Contains(id) {
					t.Error("declining a discard ward must not discard anything")
				}
			}
			if effectstest.BoltResolved(g, warded) {
				t.Error("a countered Lightning Bolt must not damage the creature")
			}
		})
	}
}

// Paying is a two-link chain — the confirm, then the caster's own
// pick from their own hand — and the bolt then resolves.
func TestDiscardWardPaidDiscardsTheChosenCardAndTheSpellResolves(t *testing.T) {
	for _, w := range wardDiscardCards {
		t.Run(w.name, func(t *testing.T) {
			g, me, opp, warded := effectstest.WardTable(t, w.name, w.typeLine, w.oracle)
			hand := effectstest.FillHand(opp, "Filler A", "Filler B")
			effectstest.BoltWardedCreature(t, g, opp, warded)

			prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
			if prompt == nil {
				t.Fatal("no ward prompt for the spell's controller")
			}
			if effectstest.ConfirmChoiceFor(g, me.ID) != nil {
				t.Error("the ward permanent's controller must not be asked to pay")
			}
			if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			if effectstest.DiscardChoiceFor(g, me.ID) != nil {
				t.Error("the ward permanent's controller must not be asked to discard")
			}
			if n := effectstest.DiscardOwed(g, opp.ID); n != 1 {
				t.Fatalf("the payer owes %d cards, want 1", n)
			}
			// The payer chooses: pitch the SECOND card, not the top one.
			effectstest.AnswerDiscard(t, g, opp.ID, hand[1])
			effectstest.PassPriorityAroundTable(t, g)

			if !opp.Graveyard.Contains(hand[1]) {
				t.Error("the chosen card should be in the graveyard")
			}
			if !opp.Hand.Contains(hand[0]) {
				t.Error("only the chosen card goes")
			}
			if !effectstest.BoltResolved(g, warded) {
				t.Error("the ward was paid — the Lightning Bolt should have resolved")
			}
		})
	}
}

// A caster with an empty hand cannot pay: the spell is countered and
// nobody is asked (CR 118.3).
func TestDiscardWardCountersWithoutAPromptWhenTheHandIsEmpty(t *testing.T) {
	g, _, opp, warded := effectstest.WardTable(t, "Gideon the Oathless", "Legendary Creature — Human Mercenary", wardDiscardCards[2].oracle)
	for opp.Hand.Size() > 0 {
		if _, err := opp.Hand.PopTop(); err != nil {
			t.Fatal(err)
		}
	}
	bolt := effectstest.BoltWardedCreature(t, g, opp, warded)

	if prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID); prompt != nil {
		t.Fatalf("a caster with no cards was offered a discard payment: %q", prompt.Reason)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(bolt) || effectstest.BoltResolved(g, warded) {
		t.Error("a ward the caster cannot pay must counter the spell")
	}
}

// Your own spell on your own warded creature does not trigger it.
func TestDiscardWardIgnoresItsControllersSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	warded := effectstest.PushCatalogPermanent(g, me.ID, "Gideon the Oathless", "Legendary Creature — Human Mercenary", wardDiscardCards[2].oracle, false)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.FillHand(me, "Filler A")
	effectstest.BoltWardedCreature(t, g, me, warded)
	if prompt := effectstest.PassUntilConfirmFor(t, g, me.ID); prompt != nil {
		t.Fatalf("a ward asked its own controller to pay: %q", prompt.Reason)
	}
}

// The cards are Full: the ward was the only declared gap.
func TestDiscardWardCardsAreComplete(t *testing.T) {
	for _, w := range wardDiscardCards {
		spec, ok := effects.Lookup(w.oracle)
		if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: want CompletenessFull with no caveats, got %v %v", w.name, spec.Completeness, spec.Caveats)
		}
	}
}
